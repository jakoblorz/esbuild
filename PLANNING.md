Please explain the different processing steps of esbuild and what makes it fast. Include code examples of the go source code
_______
Here’s the mental model for esbuild’s pipeline:
	1.	API entry
	2.	Scan phase: load files, resolve imports, parse each file into ASTs
	3.	Link phase: connect imports/exports, tree-shake, split chunks
	4.	Print/codegen phase: turn ASTs back into JS/CSS, minify names, write source maps

That’s also how esbuild’s own architecture doc describes it: scan first, then compile; and the FAQ further groups the heavy work into parsing, linking, and code generation. Parsing and code generation are designed to parallelize well, while linking is the more serial coordination step.  ￼

1) API entry

At the public API level, Build() is intentionally thin. It creates a context, triggers a rebuild, and disposes the context afterward. In pkg/api/api.go that flow looks like this:

// pkg/api/api.go
func Build(options BuildOptions) BuildResult {
    ctx, errors := contextImpl(options)
    result := ctx.Rebuild()
    ctx.Dispose()
    return result
}

That snippet matches the current public API structure shown in the repository. The important point is that the public API is mostly orchestration; the real work happens in the internal bundler, resolver, parser, linker, and printer packages.  ￼

2) Scan phase: build the module graph

The bundler source says the core has two phases: ScanBundle and Compile. The scan phase walks entry points, resolves imports, loads file contents, chooses loaders, and parses files. That phase is described directly in internal/bundler/bundler.go.  ￼

A tiny excerpt from the current source shape:

// internal/bundler/bundler.go
// The first phase scans the module graph...
// represented by the "ScanBundle" function.

And each file is processed through parseFile(args parseArgs), which sets up a logger.Source, handles stdin specially, and otherwise runs onLoad plugins before parsing. You can see that in the current parseFile implementation.  ￼

Conceptually, the scan phase does this:
	•	resolve each import path
	•	read file contents
	•	choose a loader (js, ts, tsx, css, json, etc.)
	•	parse into AST + symbol tables + import records
	•	enqueue newly discovered dependencies

Because files are mostly independent at this stage, esbuild can parse many of them concurrently. That concurrency is one of the main reasons it is fast.  ￼

3) Parsing: esbuild compresses a lot into very few passes

A big performance trick is that the parser does more than just “produce an AST.” The architecture doc says esbuild condensed parsing-related work into two passes:
	1.	lex/parse + build the scope tree + declare symbols
	2.	bind identifiers to symbols + perform lowering/transforms

That avoids extra full-AST walks that many toolchains do later as separate passes.  ￼

The doc also notes that lexer lookahead is kept to one token in almost all cases, with TypeScript being the main exception where arbitrary lookahead is sometimes necessary. That matters because bounded lookahead keeps the parser simpler and cheaper.  ￼

A simplified sketch of the parser’s role is:

// conceptualized from internal/js_parser
ast := parse(source)
declareSymbols(ast)
bindIdentifiers(ast)
lowerSyntax(ast, target)

That is not the literal source, but it matches the architecture doc’s description of what the parser packs into those two passes.  ￼

4) Linking: imports/exports, tree shaking, code splitting

After parsing, esbuild enters the compile/link stage. The architecture doc says linking merges modules so imports in one file can reference exports in another. That stage also performs tree shaking and optionally code splitting.  ￼

The important internal trick here is symbol merging. For ES module scope hoisting, imported and exported symbols are merged so they become the same logical symbol in the final output. The doc explains that this is implemented with a Link field and a FollowSymbols() step, similar to a union-find style structure.  ￼

That is a big deal for speed because it avoids repeatedly rewriting AST nodes across files. Instead of mutating everything over and over, esbuild can keep compact symbol references and resolve them later during linking/printing.  ￼

For CommonJS, esbuild may wrap a module in a helper closure; for pure ES modules, it often does scope hoisting and merges them into a broader module scope. The architecture doc explicitly contrasts those two strategies.  ￼

5) Printing/code generation

Once the linked graph is known, esbuild prints final JS/CSS and source maps. The FAQ says code generation is one of the phases that is heavily parallelized. The bundler source also notes that some source map data is computed in parallel with linking so it is ready by print time.  ￼

The DataForSourceMap structure in bundler.go shows that esbuild precomputes things such as line offset tables and quoted original contents because those are expensive enough to be worth doing ahead of time.  ￼

A small excerpt from the source structure:

// internal/bundler/bundler.go
type DataForSourceMap struct {
    LineOffsetTables []sourcemap.LineOffsetTable
    QuotedContents   [][]byte
}

That is a good example of esbuild’s style: push expensive, mechanical work into compact data structures and parallel precomputation so the printer can stay fast.  ￼

6) Minification: fast symbol renaming

The architecture doc has a useful section on symbol minification. Instead of doing heavyweight global name analysis in a slow, repeated way, esbuild assigns symbols to integer slots, counts usage frequencies, sorts by frequency, and then assigns the shortest identifiers to the most-used slots. The doc also notes that parts of this slot assignment can be done in parallel.  ￼

That approach is fast because:
	•	integer slots are cheaper than rich per-symbol objects everywhere
	•	counting is array-based
	•	chunk-local renaming can be parallelized
	•	the printer only needs a compact slot→name mapping at the end

This is classic esbuild design: favor simple data layouts and linear scans over repeated object-heavy passes.  ￼

What specifically makes esbuild fast?

The short answer is: algorithm choices + data layout + native execution + parallelism.

A. Native Go + shared-memory concurrency

The FAQ says esbuild tries to saturate all CPU cores, and because threads share memory, work can be shared efficiently between entry points importing the same libraries. That avoids the cost you’d get from isolated worker processes exchanging lots of serialized data.  ￼

B. Fewer passes over the AST

The parser folds together lexing, parsing, scope construction, symbol declaration, binding, and some lowering in just two passes. Fewer full-tree walks means less CPU and less memory traffic.  ￼

C. Compact symbol machinery

The array-of-arrays symbol map and link-following model let esbuild merge symbols across files without constantly rewriting the whole tree. That keeps cross-module linking relatively cheap.  ￼

D. Parallel parse and print

The FAQ explicitly says parsing and code generation are the bulk of the work and are fully parallelizable.  ￼

E. Source-map work is precomputed

bundler.go documents that source map data is computed in parallel with linking because it is “somewhat expensive.” That reduces print-time latency.  ￼

F. Plugin overhead is deliberately constrained

The plugin docs warn that calling from esbuild’s highly parallel internals into single-threaded JavaScript is expensive. That is why plugins require filters, so esbuild can avoid invoking JS callbacks unless necessary.  ￼

A representative example:

// plugin design idea from docs
build.OnResolve(api.OnResolveOptions{
    Filter: `^react$`,
}, func(args api.OnResolveArgs) ...)

The point of the filter is not just ergonomics; it is a performance boundary.  ￼

End-to-end flow in one picture

This is the practical pipeline:

Build()
  -> ScanBundle()
       -> resolve imports
       -> run onLoad plugins
       -> choose loader
       -> parse file
       -> collect symbols/import records
  -> Compile()
       -> link imports to exports
       -> tree shake
       -> split chunks
       -> minify symbols
       -> print JS/CSS
       -> emit source maps

That summary is directly consistent with esbuild’s architecture doc, FAQ, and the current source layout.  ￼

The most important takeaway

esbuild is not just “fast because it is written in Go.” Go helps, but the bigger reason is that the implementation is structured to avoid expensive work:
	•	minimal parsing passes
	•	compact symbol representation
	•	shared-memory parallelism
	•	selective plugin crossing into JS
	•	precomputed source map/minification metadata
	•	linking that reuses symbol indirection instead of repeatedly rewriting ASTs

That combination is what makes it feel dramatically faster than many JS-based build pipelines.  ￼

I can also turn this into a file-by-file walkthrough of the current esbuild repository, starting with pkg/api/api.go, internal/bundler/bundler.go, internal/js_parser/*, and internal/linker/*.
_______
Let's say I want to add a pre-processing / transform step to esbuild (in my personal fork). I want to enable reflect.metadata / decorators by essentially transforming the code first to wrap the function on which the decorator is built on. How would I do that?
_______
Yes — but the clean way in an esbuild fork is not to add a text pre-pass. The best place is to add a new lowering step in the parser/lowering pipeline, right next to where esbuild already lowers TypeScript’s experimentalDecorators. esbuild’s architecture is explicitly optimized to avoid “print JS → reparse JS” intermediate stages and to keep the number of full-AST passes low, so an AST-level transform fits the project much better than a string rewrite.  ￼

The important constraint is this: upstream esbuild supports TypeScript’s experimentalDecorators, but does not support emitDecoratorMetadata, and the stated reason is that emitting metadata requires type information. esbuild parses TS syntax, but it is not a type checker. So if your fork wants Reflect.metadata("design:type" | "design:paramtypes" | "design:returntype", ...), you either need to accept a syntax-only approximation or you need to bolt on some external type source.  ￼

What I would change

There are really two implementation options.

Option A: quick-and-dirty “preprocess source text”

Hook into the scan/load path before parsing and rewrite the file contents. This is the fastest thing to prototype, but it’s the wrong long-term fit because it makes source maps, symbol binding, decorator ordering, class field semantics, and tree shaking harder to keep correct. esbuild’s scan phase parses each file into an AST in parallel, and the architecture docs strongly suggest keeping transformations in shared data structures instead of converting between text and AST repeatedly.  ￼

Option B: add a new AST transform in class lowering

This is the one I’d recommend. esbuild already has decorator lowering machinery and runtime helpers for it. The architecture docs call out the runtime helper library as the place for helpers such as decorator support, and current releases/changelogs show generated helper calls such as __decorateClass(...) and __decorateParam(...).  ￼

That means the path is:
	1.	Parse decorators as esbuild already does.
	2.	Preserve enough metadata on class/method/field declarations.
	3.	During lowering, generate extra helper calls or wrapper expressions.
	4.	Inject any new helper into internal/runtime/runtime.go.
	5.	Optionally add a fork-only tsconfig/build flag to enable it. esbuild already parses experimentalDecorators out of tsconfig.json, so there is precedent for plumbing a TS-specific flag through config.  ￼

Where to patch in the codebase

The rough places are:
	•	internal/resolver/tsconfig_json.go — parse a fork-only flag such as emitDecoratorMetadata or reflectMetadata
	•	config structs under the internal config/api plumbing — carry that flag through the build
	•	internal/js_parser/... — collect enough info from decorated declarations
	•	internal/js_parser/js_parser_lower_class.go — emit the transformed form
	•	internal/runtime/runtime.go — add a helper if you want a helper-based lowering instead of raw AST for Reflect.metadata(...) calls  ￼

That location also matches how esbuild already handles decorator lowering. The changelog shows actual emitted code of the form:

class Class {
  method(arg) {}
}
__decorate([
  __param(0, decorator())
], Class.prototype, "method", null);

and newer code paths use helper forms such as __decorateClass(...). So you’re not inventing a brand-new stage; you’re extending an existing one.  ￼

The design choice: wrap the function, or inject metadata calls?

You said you want to “wrap the function on which the decorator is built on.” That can work, but for TS-style metadata I would usually inject metadata decorators into the existing decorator array instead of wrapping the function body.

For example, this:

class UserService {
  @Log
  save(id: number, name: string): Promise<void> {}
}

would become conceptually:

class UserService {
  save(id, name) {}
}

__decorateClass([
  Log,
  Reflect.metadata("design:type", Function),
  Reflect.metadata("design:paramtypes", [Number, String]),
  Reflect.metadata("design:returntype", Promise),
], UserService.prototype, "save", 1)

That approach is much closer to how esbuild already lowers decorators, so it should integrate better with ordering, tree shaking, and helper injection. The catch is that Number, String, Promise, and so on require type knowledge for correctness, which upstream esbuild intentionally doesn’t have.  ￼

If all you really need is “wrap the decorated function,” then you can lower:

class C {
  @dec
  method(x) { return x + 1 }
}

into something equivalent to:

class C {
  method(x) { return x + 1 }
}

C.prototype.method = dec(C.prototype, "method", C.prototype.method) ?? C.prototype.method

But that is not TypeScript-compatible decorator lowering. esbuild’s existing decorator transform already handles the subtleties around class elements and parameter decorators, including scope behavior that changed in past releases. So I would piggyback on that machinery instead of replacing it with an ad hoc wrapper transform.  ￼

A practical fork plan

1) Add a fork-only flag

In internal/resolver/tsconfig_json.go, esbuild already reads experimentalDecorators. Add another bool for your fork:

// pseudo-code patterned after esbuild's tsconfig parsing
if valueJSON, _, ok := getProperty(compilerOptionsJSON, "emitDecoratorMetadata"); ok {
    if value, ok := getBool(valueJSON); ok {
        result.Settings.EmitDecoratorMetadata = value
    }
}

That mirrors the existing parsing path for TS compiler options.  ￼

Then thread that through the internal config structs until it is visible in the parser/lowerer.

2) Reuse the class-lowering pass

In class lowering, find the code path that collects decorators and emits helper calls. Instead of inventing a separate transform stage, extend the existing lowering result with extra decorator expressions.

Pseudo-Go:

type MetadataExprs struct {
    DesignType       js_ast.Expr
    DesignParamTypes []js_ast.Expr
    DesignReturnType js_ast.Expr
}

func (p *parser) maybeGenerateReflectMetadata(
    class *js_ast.Class,
    prop *js_ast.Property,
) []js_ast.Expr {
    if !p.options.EmitDecoratorMetadata {
        return nil
    }

    // Syntax-only approximation:
    // method => Function
    // getter/setter => Function or inferred
    // property with ": string" => String
    // parameter ": Foo" => Foo
    // unknown => void 0 / Object
    md := inferMetadataFromTSSyntax(prop)

    var extra []js_ast.Expr
    if md.DesignType.Data != nil {
        extra = append(extra, makeReflectMetadataCall("design:type", md.DesignType))
    }
    if len(md.DesignParamTypes) > 0 {
        extra = append(extra, makeReflectMetadataArrayCall("design:paramtypes", md.DesignParamTypes))
    }
    if md.DesignReturnType.Data != nil {
        extra = append(extra, makeReflectMetadataCall("design:returntype", md.DesignReturnType))
    }
    return extra
}

Then, where esbuild currently builds the decorator helper call, append extra to the decorator list before emitting __decorateClass(...).

3) Emit AST, not strings

Build Reflect.metadata(...) as normal AST nodes.

Pseudo-Go:

func makeReflectMetadataCall(key string, value js_ast.Expr) js_ast.Expr {
    return js_ast.Expr{Data: &js_ast.ECall{
        Target: js_ast.Expr{Data: &js_ast.EDot{
            Target:  js_ast.Expr{Data: &js_ast.EIdentifier{Ref: reflectRef}},
            Name:    "metadata",
            NameLoc: logger.Loc{},
        }},
        Args: []js_ast.Expr{
            {Data: &js_ast.EString{Value: helpers.StringToUTF16(key)}},
            value,
        },
    }}
}

The exact node constructors may differ from current internal names, but the idea is the important part: generate esbuild AST directly, so printing and source maps continue to work with the existing pipeline. That matches the architecture doc’s “avoid unnecessary work / avoid extra passes” philosophy.  ￼

4) Mark the runtime symbol usage if needed

If you only emit Reflect.metadata(...), you may not need a runtime helper at all. But if you want a helper such as __metadata(k, v), put it in internal/runtime/runtime.go, because that file is specifically where esbuild keeps injected helpers and lets tree shaking strip unused ones.  ￼

For example:

// internal/runtime/runtime.go (conceptual)
export var __metadata = (k, v) =>
  typeof Reflect === "object" && typeof Reflect.metadata === "function"
    ? Reflect.metadata(k, v)
    : void 0

Then lower to:

__decorateClass([
  dec,
  __metadata("design:paramtypes", [Number, String])
], C.prototype, "method", 1)

That gives you graceful behavior when reflect-metadata isn’t loaded.

What “type inference” can you do without a type checker?

This is the hard part. Upstream esbuild declined emitDecoratorMetadata specifically because proper output needs the TS type checker.  ￼

Still, for a personal fork, a syntax-only approximation is often good enough:
	•	string → String
	•	number → Number
	•	boolean → Boolean
	•	bigint → BigInt
	•	symbol → Symbol
	•	any, unknown, object literal types, unions, intersections → Object
	•	Foo → identifier Foo
	•	Foo[] / Array<Foo> → Array
	•	method declarations → Function
	•	constructors → omit or map as needed
	•	return type Promise<T> → Promise

Pseudo-Go:

func inferRuntimeTypeFromTSType(expr js_ast.TSType) js_ast.Expr {
    switch t := expr.Data.(type) {
    case *js_ast.TSStringKeyword:
        return ident("String")
    case *js_ast.TSNumberKeyword:
        return ident("Number")
    case *js_ast.TSBooleanKeyword:
        return ident("Boolean")
    case *js_ast.TSTypeReference:
        return typeRefToRuntimeExpr(t)
    case *js_ast.TSArrayType:
        return ident("Array")
    default:
        return ident("Object")
    }
}

That will not match tsc in all cases, but it is a realistic compromise for a fork.

Why not do it in parseFile() as a source rewrite?

You can hook parseFile(args parseArgs) during scan and rewrite the loaded source before passing it into the parser. That’s likely the easiest place for a prototype because scan/load happens before AST creation. But esbuild’s scan phase is a parallel worklist that parses each file into an AST and then feeds the compile phase, and the project’s design principles explicitly favor operating on shared internal data structures instead of converting back and forth through text. So a source rewrite is workable for an experiment, but it fights the architecture.  ￼

My recommendation

For a personal fork, I’d do this:
	•	add a new internal flag
	•	extend the existing decorator-lowering path in js_parser_lower_class.go
	•	emit extra decorator expressions for metadata
	•	optionally add __metadata() to runtime.go
	•	keep it syntax-only unless you also plan to integrate the TS checker

That gives you something that feels native to esbuild instead of bolted on.

Minimal conceptual example

Input:

class Foo {
  @inject()
  bar(name: string, count: number): Promise<void> {}
}

Lowering target in your fork:

class Foo {
  bar(name, count) {}
}

__decorateClass([
  inject(),
  __metadata("design:type", Function),
  __metadata("design:paramtypes", [String, Number]),
  __metadata("design:returntype", Promise)
], Foo.prototype, "bar", 1)

And helper:

var __metadata = (k, v) =>
  typeof Reflect === "object" && typeof Reflect.metadata === "function"
    ? Reflect.metadata(k, v)
    : void 0

That is the closest fit to esbuild’s current decorator architecture.  ￼

The main thing I would not do is a wrapper transform that rewrites methods as arbitrary closures before class lowering. You’ll end up fighting class semantics, super, fields, parameter decorators, and ordering — all areas where esbuild has already accumulated special-case lowering logic over multiple releases.  ￼

I can sketch the exact files and pseudo-diffs for a fork, centered on tsconfig_json.go, js_parser_lower_class.go, and runtime.go.
_______