package js_parser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/evanw/esbuild/internal/ast"
	"github.com/evanw/esbuild/internal/config"
	"github.com/evanw/esbuild/internal/helpers"
	"github.com/evanw/esbuild/internal/js_ast"
	"github.com/evanw/esbuild/internal/js_printer"
	"github.com/evanw/esbuild/internal/logger"
	"github.com/evanw/esbuild/internal/renamer"
	"github.com/evanw/esbuild/internal/test"
)

type decoratorMetadataFixture struct {
	TypeScriptVersion string                    `json:"typescriptVersion"`
	Source            string                    `json:"source"`
	Records           []decoratorMetadataRecord `json:"records"`
}

type decoratorMetadataRecord struct {
	Kind     string                  `json:"kind"`
	Target   string                  `json:"target"`
	Key      *string                 `json:"key"`
	Metadata []decoratorMetadataPair `json:"metadata"`
}

type decoratorMetadataPair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func TestTSEmitDecoratorMetadataFixtures(t *testing.T) {
	rootDir := filepath.Join("testdata", "emit_decorator_metadata")
	casesDir := filepath.Join(rootDir, "cases")
	fixturesDir := filepath.Join(rootDir, "fixtures")

	entries, err := os.ReadDir(casesDir)
	if err != nil {
		t.Fatalf("failed to read cases directory: %v", err)
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".ts") {
			continue
		}
		names = append(names, strings.TrimSuffix(entry.Name(), ".ts"))
	}
	sort.Strings(names)

	for _, name := range names {
		name := name
		t.Run(name, func(t *testing.T) {
			casePath := filepath.Join(casesDir, name+".ts")
			fixturePath := filepath.Join(fixturesDir, name+".json")

			contents, err := os.ReadFile(casePath)
			if err != nil {
				t.Fatalf("failed to read case %q: %v", casePath, err)
			}

			fixtureBytes, err := os.ReadFile(fixturePath)
			if err != nil {
				t.Fatalf("failed to read fixture %q: %v", fixturePath, err)
			}

			var fixture decoratorMetadataFixture
			if err := json.Unmarshal(fixtureBytes, &fixture); err != nil {
				t.Fatalf("failed to parse fixture %q: %v", fixturePath, err)
			}

			generated := compileTSWithDecoratorMetadataForTest(t, string(contents))
			actual := extractDecoratorMetadataRecordsForTest(t, generated)

			actualBytes, err := json.MarshalIndent(actual, "", "  ")
			if err != nil {
				t.Fatalf("failed to encode generated records: %v", err)
			}
			expectedBytes, err := json.MarshalIndent(fixture.Records, "", "  ")
			if err != nil {
				t.Fatalf("failed to encode fixture records: %v", err)
			}

			test.AssertEqualWithDiff(t, string(actualBytes)+"\n", string(expectedBytes)+"\n")
		})
	}
}

func compileTSWithDecoratorMetadataForTest(t *testing.T, contents string) string {
	t.Helper()
	log := logger.NewDeferLog(logger.DeferLogNoVerboseOrDebug, nil)
	options := config.Options{
		TS: config.TSOptions{
			Parse: true,
			Config: config.TSConfig{
				ExperimentalDecorators:  config.True,
				EmitDecoratorMetadata:   config.True,
				UseDefineForClassFields: config.False,
			},
		},
		OmitRuntimeForTests: true,
	}

	tree, ok := Parse(log, test.SourceForTest(contents), OptionsFromConfig(&options))
	msgs := log.Done()

	var text strings.Builder
	for _, msg := range msgs {
		if msg.Kind != logger.Warning {
			text.WriteString(msg.String(logger.OutputOptions{}, logger.TerminalInfo{}))
		}
	}
	if text.Len() > 0 {
		t.Fatalf("unexpected parse diagnostics:\n%s", text.String())
	}
	if !ok {
		t.Fatal("parse failed")
	}

	symbols := ast.NewSymbolMap(1)
	symbols.SymbolsForSource[0] = tree.Symbols
	r := renamer.NewNoOpRenamer(symbols)
	js := js_printer.Print(tree, symbols, r, js_printer.Options{}).JS
	return string(js)
}

func extractDecoratorMetadataRecordsForTest(t *testing.T, contents string) []decoratorMetadataRecord {
	t.Helper()
	log := logger.NewDeferLog(logger.DeferLogNoVerboseOrDebug, nil)
	tree, ok := Parse(log, test.SourceForTest(contents), OptionsFromConfig(&config.Options{}))
	msgs := log.Done()
	if !ok || len(msgs) > 0 {
		var text strings.Builder
		for _, msg := range msgs {
			text.WriteString(msg.String(logger.OutputOptions{}, logger.TerminalInfo{}))
		}
		t.Fatalf("failed to parse generated output:\n%s", text.String())
	}

	var records []decoratorMetadataRecord
	for _, part := range tree.Parts {
		for _, stmt := range part.Stmts {
			walkStmtForDecoratorMetadataTest(stmt, tree.Symbols, &records)
		}
	}
	return records
}

func walkStmtForDecoratorMetadataTest(stmt js_ast.Stmt, symbols []ast.Symbol, records *[]decoratorMetadataRecord) {
	switch s := stmt.Data.(type) {
	case *js_ast.SExpr:
		walkExprForDecoratorMetadataTest(s.Value, symbols, records)

	case *js_ast.SLocal:
		for _, decl := range s.Decls {
			if decl.ValueOrNil.Data != nil {
				walkExprForDecoratorMetadataTest(decl.ValueOrNil, symbols, records)
			}
		}

	case *js_ast.SClass:
		for _, prop := range s.Class.Properties {
			if prop.Key.Data != nil {
				walkExprForDecoratorMetadataTest(prop.Key, symbols, records)
			}
			if prop.ValueOrNil.Data != nil {
				walkExprForDecoratorMetadataTest(prop.ValueOrNil, symbols, records)
			}
			if prop.InitializerOrNil.Data != nil {
				walkExprForDecoratorMetadataTest(prop.InitializerOrNil, symbols, records)
			}
		}
	}
}

func walkExprForDecoratorMetadataTest(expr js_ast.Expr, symbols []ast.Symbol, records *[]decoratorMetadataRecord) {
	if call, ok := expr.Data.(*js_ast.ECall); ok {
		if record, ok := parseDecoratorCallForMetadataTest(call, symbols); ok {
			*records = append(*records, record)
		}
		walkExprForDecoratorMetadataTest(call.Target, symbols, records)
		for _, arg := range call.Args {
			walkExprForDecoratorMetadataTest(arg, symbols, records)
		}
		return
	}

	switch e := expr.Data.(type) {
	case *js_ast.EArray:
		for _, item := range e.Items {
			walkExprForDecoratorMetadataTest(item, symbols, records)
		}

	case *js_ast.EBinary:
		walkExprForDecoratorMetadataTest(e.Left, symbols, records)
		walkExprForDecoratorMetadataTest(e.Right, symbols, records)

	case *js_ast.EUnary:
		walkExprForDecoratorMetadataTest(e.Value, symbols, records)

	case *js_ast.EDot:
		walkExprForDecoratorMetadataTest(e.Target, symbols, records)

	case *js_ast.EIndex:
		walkExprForDecoratorMetadataTest(e.Target, symbols, records)
		walkExprForDecoratorMetadataTest(e.Index, symbols, records)

	case *js_ast.EObject:
		for _, prop := range e.Properties {
			walkExprForDecoratorMetadataTest(prop.Key, symbols, records)
			if prop.ValueOrNil.Data != nil {
				walkExprForDecoratorMetadataTest(prop.ValueOrNil, symbols, records)
			}
			if prop.InitializerOrNil.Data != nil {
				walkExprForDecoratorMetadataTest(prop.InitializerOrNil, symbols, records)
			}
		}

	case *js_ast.EIf:
		walkExprForDecoratorMetadataTest(e.Test, symbols, records)
		walkExprForDecoratorMetadataTest(e.Yes, symbols, records)
		walkExprForDecoratorMetadataTest(e.No, symbols, records)

	case *js_ast.EFunction:
		for _, stmt := range e.Fn.Body.Block.Stmts {
			walkStmtForDecoratorMetadataTest(stmt, symbols, records)
		}

	case *js_ast.EClass:
		for _, prop := range e.Class.Properties {
			walkExprForDecoratorMetadataTest(prop.Key, symbols, records)
			if prop.ValueOrNil.Data != nil {
				walkExprForDecoratorMetadataTest(prop.ValueOrNil, symbols, records)
			}
			if prop.InitializerOrNil.Data != nil {
				walkExprForDecoratorMetadataTest(prop.InitializerOrNil, symbols, records)
			}
		}
	}
}

func parseDecoratorCallForMetadataTest(call *js_ast.ECall, symbols []ast.Symbol) (decoratorMetadataRecord, bool) {
	helper, ok := helperNameForMetadataTest(call.Target, symbols)
	if !ok || (helper != "__decorate" && helper != "__decorateClass") {
		return decoratorMetadataRecord{}, false
	}
	if len(call.Args) < 2 {
		return decoratorMetadataRecord{}, false
	}

	decoratorArray, ok := call.Args[0].Data.(*js_ast.EArray)
	if !ok {
		return decoratorMetadataRecord{}, false
	}

	metadata := []decoratorMetadataPair{}
	for _, item := range decoratorArray.Items {
		entry, ok := parseMetadataEntryForMetadataTest(item, symbols)
		if ok {
			metadata = append(metadata, entry)
		}
	}
	if len(metadata) == 0 {
		return decoratorMetadataRecord{}, false
	}

	record := decoratorMetadataRecord{
		Kind:     "class",
		Target:   serializeMetadataExprForTest(call.Args[1], symbols),
		Metadata: metadata,
	}
	if len(call.Args) >= 3 {
		record.Kind = "member"
		key := serializeMetadataExprForTest(call.Args[2], symbols)
		record.Key = &key
	}
	return record, true
}

func parseMetadataEntryForMetadataTest(expr js_ast.Expr, symbols []ast.Symbol) (decoratorMetadataPair, bool) {
	call, ok := expr.Data.(*js_ast.ECall)
	if !ok || len(call.Args) < 2 {
		return decoratorMetadataPair{}, false
	}

	helper, ok := helperNameForMetadataTest(call.Target, symbols)
	if !ok {
		if dot, ok := call.Target.Data.(*js_ast.EDot); ok {
			if targetName, ok := helperNameForMetadataTest(dot.Target, symbols); ok && targetName == "Reflect" && dot.Name == "metadata" {
				helper = "Reflect.metadata"
				ok = true
			}
		}
	}

	if !ok || (helper != "__metadata" && helper != "__legacyMetadata" && helper != "Reflect.metadata") {
		return decoratorMetadataPair{}, false
	}

	str, ok := call.Args[0].Data.(*js_ast.EString)
	if !ok {
		return decoratorMetadataPair{}, false
	}

	return decoratorMetadataPair{
		Key:   helpers.UTF16ToString(str.Value),
		Value: serializeMetadataExprForTest(call.Args[1], symbols),
	}, true
}

func helperNameForMetadataTest(expr js_ast.Expr, symbols []ast.Symbol) (string, bool) {
	switch e := expr.Data.(type) {
	case *js_ast.EIdentifier:
		return symbolNameForMetadataTest(e.Ref, symbols)
	case *js_ast.EImportIdentifier:
		return symbolNameForMetadataTest(e.Ref, symbols)
	}
	return "", false
}

func symbolNameForMetadataTest(ref ast.Ref, symbols []ast.Symbol) (string, bool) {
	for ref != ast.InvalidRef {
		if int(ref.InnerIndex) >= len(symbols) {
			return "", false
		}
		link := symbols[ref.InnerIndex].Link
		if link == ast.InvalidRef {
			return symbols[ref.InnerIndex].OriginalName, true
		}
		ref = link
	}
	return "", false
}

func serializeMetadataExprForTest(expr js_ast.Expr, symbols []ast.Symbol) string {
	if name, ok := helperNameForMetadataTest(expr, symbols); ok {
		return name
	}

	switch e := expr.Data.(type) {
	case *js_ast.EString:
		return strconv.Quote(helpers.UTF16ToString(e.Value))

	case *js_ast.ENumber:
		return strconv.FormatFloat(e.Value, 'g', -1, 64)

	case *js_ast.EBigInt:
		return e.Value

	case *js_ast.EBoolean:
		if e.Value {
			return "true"
		}
		return "false"

	case *js_ast.EDot:
		return serializeMetadataExprForTest(e.Target, symbols) + "." + e.Name

	case *js_ast.EIndex:
		return serializeMetadataExprForTest(e.Target, symbols) + "[" + serializeMetadataExprForTest(e.Index, symbols) + "]"

	case *js_ast.EArray:
		parts := make([]string, len(e.Items))
		for i, item := range e.Items {
			parts[i] = serializeMetadataExprForTest(item, symbols)
		}
		return "[" + strings.Join(parts, ", ") + "]"

	case *js_ast.EUnary:
		if e.Op == js_ast.UnOpVoid {
			return "void " + serializeMetadataExprForTest(e.Value, symbols)
		}
	}

	if expr.Data == js_ast.ENullShared {
		return "null"
	}
	if expr.Data == js_ast.EUndefinedShared {
		return "void 0"
	}

	return "<unsupported>"
}
