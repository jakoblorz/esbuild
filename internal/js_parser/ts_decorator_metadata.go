package js_parser

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/evanw/esbuild/internal/ast"
	"github.com/evanw/esbuild/internal/config"
	"github.com/evanw/esbuild/internal/js_ast"
	"github.com/evanw/esbuild/internal/logger"
)

func (p *parser) shouldCaptureTypeScriptDecoratorMetadata() bool {
	return p.options.ts.Parse &&
		p.options.ts.Config.ExperimentalDecorators == config.True &&
		p.options.ts.Config.EmitDecoratorMetadata == config.True
}

func (p *parser) decoratorMetadataIdentifierExpr(loc logger.Loc, name string) js_ast.Expr {
	var ref ast.Ref
	var ok bool
	if ref, ok = p.decoratorMetadataGlobals[name]; !ok {
		ref = p.newSymbol(ast.SymbolUnbound, name)
		scope := p.moduleScope
		if scope == nil {
			scope = p.currentScope
		}
		scope.Generated = append(scope.Generated, ref)
		p.decoratorMetadataGlobals[name] = ref
	}
	return js_ast.Expr{Loc: loc, Data: &js_ast.EIdentifier{Ref: ref}}
}

func (p *parser) decoratorMetadataObjectExpr(loc logger.Loc) js_ast.Expr {
	return p.decoratorMetadataIdentifierExpr(loc, "Object")
}

func (p *parser) decoratorMetadataFunctionExpr(loc logger.Loc) js_ast.Expr {
	return p.decoratorMetadataIdentifierExpr(loc, "Function")
}

func (p *parser) decoratorMetadataArrayExpr(loc logger.Loc) js_ast.Expr {
	return p.decoratorMetadataIdentifierExpr(loc, "Array")
}

func (p *parser) decoratorMetadataPromiseExpr(loc logger.Loc) js_ast.Expr {
	return p.decoratorMetadataIdentifierExpr(loc, "Promise")
}

func (p *parser) decoratorMetadataVoidExpr(loc logger.Loc) js_ast.Expr {
	return js_ast.Expr{Loc: loc, Data: &js_ast.EUnary{
		Op:    js_ast.UnOpVoid,
		Value: js_ast.Expr{Loc: loc, Data: &js_ast.ENumber{Value: 0}},
	}}
}

func (p *parser) skipTypeScriptTypeAndCaptureDecoratorMetadata() js_ast.Expr {
	start := p.lexer.Range().Loc.Start
	p.skipTypeScriptType(js_ast.LLowest)
	end := p.lexer.Range().Loc.Start
	return p.decoratorMetadataExprFromTypeRange(start, end, logger.Loc{Start: start})
}

func (p *parser) skipTypeScriptReturnTypeAndCaptureDecoratorMetadata() js_ast.Expr {
	start := p.lexer.Range().Loc.Start
	p.skipTypeScriptReturnType()
	end := p.lexer.Range().Loc.Start
	return p.decoratorMetadataExprFromTypeRange(start, end, logger.Loc{Start: start})
}

func (p *parser) decoratorMetadataExprFromTypeRange(start int32, end int32, loc logger.Loc) js_ast.Expr {
	if start < 0 || end < start || end > int32(len(p.source.Contents)) {
		return p.decoratorMetadataObjectExpr(loc)
	}
	text := strings.TrimSpace(p.source.Contents[start:end])
	return p.decoratorMetadataExprFromTypeText(text, loc)
}

func (p *parser) resolveLocalTypeMetadata(name string, seen map[string]bool) (js_ast.Expr, bool) {
	expr, ok := p.localTypeMetadata[name]
	if !ok {
		return js_ast.Expr{}, false
	}
	if seen[name] {
		return p.decoratorMetadataObjectExpr(expr.Loc), true
	}
	seen[name] = true
	if id, ok := expr.Data.(*js_ast.EIdentifier); ok {
		idName := ""
		if (id.Ref.SourceIndex & 0x80000000) != 0 {
			idName = p.loadNameFromRef(id.Ref)
		} else if int(id.Ref.InnerIndex) < len(p.symbols) {
			idName = p.symbols[id.Ref.InnerIndex].OriginalName
		}
		if idName != "" {
			if resolved, ok := p.resolveLocalTypeMetadata(idName, seen); ok {
				return resolved, true
			}
		}
	}
	return expr, true
}

func (p *parser) decoratorMetadataExprFromTypeText(text string, loc logger.Loc) js_ast.Expr {
	text = strings.TrimSpace(text)
	if text == "" {
		return p.decoratorMetadataObjectExpr(loc)
	}

	for {
		trimmed := strings.TrimSpace(text)
		if len(trimmed) < 2 || trimmed[0] != '(' || trimmed[len(trimmed)-1] != ')' || !typeTextHasOuterParens(trimmed) {
			break
		}
		text = strings.TrimSpace(trimmed[1 : len(trimmed)-1])
	}

	if typeTextHasTopLevel(text, '|') || typeTextHasTopLevel(text, '&') || typeTextHasTopLevelConditional(text) {
		return p.decoratorMetadataObjectExpr(loc)
	}

	if typeTextHasTopLevelArrow(text) || strings.HasPrefix(text, "new ") || strings.HasPrefix(text, "abstract new ") {
		return p.decoratorMetadataFunctionExpr(loc)
	}

	for strings.HasSuffix(strings.TrimSpace(text), "[]") {
		text = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(text), "[]"))
		if text == "" {
			break
		}
		return p.decoratorMetadataArrayExpr(loc)
	}

	if len(text) >= 2 && text[0] == '[' && text[len(text)-1] == ']' && typeTextHasOuterBrackets(text, '[', ']') {
		return p.decoratorMetadataArrayExpr(loc)
	}

	switch text {
	case "string":
		return p.decoratorMetadataIdentifierExpr(loc, "String")
	case "number":
		return p.decoratorMetadataIdentifierExpr(loc, "Number")
	case "boolean":
		return p.decoratorMetadataIdentifierExpr(loc, "Boolean")
	case "symbol":
		return p.decoratorMetadataIdentifierExpr(loc, "Symbol")
	case "bigint":
		return p.decoratorMetadataIdentifierExpr(loc, "BigInt")
	case "any", "unknown", "object", "this":
		return p.decoratorMetadataObjectExpr(loc)
	case "void", "undefined", "null", "never":
		return p.decoratorMetadataVoidExpr(loc)
	case "true", "false":
		return p.decoratorMetadataIdentifierExpr(loc, "Boolean")
	}

	if strings.HasPrefix(text, "typeof ") || strings.HasPrefix(text, "keyof ") || strings.HasPrefix(text, "readonly ") ||
		strings.HasPrefix(text, "infer ") || strings.HasPrefix(text, "import(") || strings.HasPrefix(text, "{") {
		return p.decoratorMetadataObjectExpr(loc)
	}

	if typeTextIsStringLiteral(text) {
		return p.decoratorMetadataIdentifierExpr(loc, "String")
	}

	if strings.HasSuffix(text, "n") {
		if _, err := strconv.ParseUint(strings.TrimSuffix(text, "n"), 10, 64); err == nil {
			return p.decoratorMetadataIdentifierExpr(loc, "BigInt")
		}
	}

	if _, err := strconv.ParseFloat(text, 64); err == nil {
		return p.decoratorMetadataIdentifierExpr(loc, "Number")
	}

	if expr, baseName, ok := p.decoratorMetadataEntityNameToExpr(text, loc); ok {
		if resolved, ok := p.resolveLocalTypeMetadata(baseName, make(map[string]bool)); ok {
			return resolved
		}
		if p.localEnumNames[baseName] {
			return p.decoratorMetadataIdentifierExpr(loc, "Number")
		}
		if p.localTypeNames[baseName] {
			return p.decoratorMetadataObjectExpr(loc)
		}
		if baseName == "Array" || baseName == "ReadonlyArray" {
			return p.decoratorMetadataArrayExpr(loc)
		}
		if baseName == "Promise" {
			return p.decoratorMetadataPromiseExpr(loc)
		}
		return expr
	}

	return p.decoratorMetadataObjectExpr(loc)
}

func typeTextIsStringLiteral(text string) bool {
	if len(text) < 2 {
		return false
	}
	quote := text[0]
	if (quote != '\'' && quote != '"' && quote != '`') || text[len(text)-1] != quote {
		return false
	}
	return true
}

func (p *parser) decoratorMetadataEntityNameToExpr(text string, loc logger.Loc) (expr js_ast.Expr, baseName string, ok bool) {
	text = strings.TrimSpace(text)
	i := 0

	readIdent := func() (string, bool) {
		for i < len(text) && unicode.IsSpace(rune(text[i])) {
			i++
		}
		start := i
		if i >= len(text) {
			return "", false
		}
		r := rune(text[i])
		if !(r == '_' || r == '$' || unicode.IsLetter(r)) {
			return "", false
		}
		i++
		for i < len(text) {
			r = rune(text[i])
			if !(r == '_' || r == '$' || unicode.IsLetter(r) || unicode.IsDigit(r)) {
				break
			}
			i++
		}
		return text[start:i], true
	}

	skipTypeArgs := func() bool {
		if i >= len(text) || text[i] != '<' {
			return true
		}
		depth := 1
		i++
		for i < len(text) {
			c := text[i]
			switch c {
			case '\'', '"', '`':
				quote := c
				i++
				for i < len(text) && text[i] != quote {
					if text[i] == '\\' {
						i++
					}
					i++
				}
			case '<':
				depth++
			case '>':
				if i > 0 && text[i-1] == '=' {
					break
				}
				depth--
				if depth == 0 {
					i++
					return true
				}
			}
			i++
		}
		return false
	}

	name, ok := readIdent()
	if !ok {
		return js_ast.Expr{}, "", false
	}
	baseName = name
	expr = p.decoratorMetadataIdentifierExpr(loc, name)

	if !skipTypeArgs() {
		return js_ast.Expr{}, "", false
	}

	for {
		for i < len(text) && unicode.IsSpace(rune(text[i])) {
			i++
		}
		if i >= len(text) || text[i] != '.' {
			break
		}
		i++
		name, ok = readIdent()
		if !ok {
			return js_ast.Expr{}, "", false
		}
		expr = js_ast.Expr{Loc: loc, Data: &js_ast.EDot{Target: expr, Name: name, NameLoc: loc}}
		if !skipTypeArgs() {
			return js_ast.Expr{}, "", false
		}
	}

	for i < len(text) && unicode.IsSpace(rune(text[i])) {
		i++
	}
	for i < len(text) && text[i] == '!' {
		i++
		for i < len(text) && unicode.IsSpace(rune(text[i])) {
			i++
		}
	}

	if i != len(text) {
		return js_ast.Expr{}, "", false
	}
	return expr, baseName, true
}

func typeTextHasOuterParens(text string) bool {
	depthRound := 0
	depthSquare := 0
	depthCurly := 0
	depthAngle := 0
	for i := 0; i < len(text); i++ {
		c := text[i]
		switch c {
		case '\'', '"', '`':
			quote := c
			i++
			for i < len(text) && text[i] != quote {
				if text[i] == '\\' {
					i++
				}
				i++
			}
		case '(':
			depthRound++
		case ')':
			depthRound--
			if depthRound == 0 && i != len(text)-1 {
				return false
			}
		case '[':
			depthSquare++
		case ']':
			depthSquare--
		case '{':
			depthCurly++
		case '}':
			depthCurly--
		case '<':
			depthAngle++
		case '>':
			if i > 0 && text[i-1] == '=' {
				break
			}
			depthAngle--
		}
		if depthRound < 0 || depthSquare < 0 || depthCurly < 0 || depthAngle < 0 {
			return false
		}
	}
	return depthRound == 0 && depthSquare == 0 && depthCurly == 0 && depthAngle == 0
}

func typeTextHasOuterBrackets(text string, open byte, close byte) bool {
	depth := 0
	for i := 0; i < len(text); i++ {
		c := text[i]
		switch c {
		case '\'', '"', '`':
			quote := c
			i++
			for i < len(text) && text[i] != quote {
				if text[i] == '\\' {
					i++
				}
				i++
			}
		case open:
			depth++
		case close:
			depth--
			if depth == 0 && i != len(text)-1 {
				return false
			}
		}
		if depth < 0 {
			return false
		}
	}
	return depth == 0
}

func typeTextHasTopLevel(text string, target byte) bool {
	depthRound := 0
	depthSquare := 0
	depthCurly := 0
	depthAngle := 0
	for i := 0; i < len(text); i++ {
		c := text[i]
		switch c {
		case '\'', '"', '`':
			quote := c
			i++
			for i < len(text) && text[i] != quote {
				if text[i] == '\\' {
					i++
				}
				i++
			}
		case '(':
			depthRound++
		case ')':
			depthRound--
		case '[':
			depthSquare++
		case ']':
			depthSquare--
		case '{':
			depthCurly++
		case '}':
			depthCurly--
		case '<':
			depthAngle++
		case '>':
			if i > 0 && text[i-1] == '=' {
				break
			}
			depthAngle--
		}

		if depthRound == 0 && depthSquare == 0 && depthCurly == 0 && depthAngle == 0 && c == target {
			return true
		}
	}
	return false
}

func typeTextHasTopLevelConditional(text string) bool {
	return typeTextHasTopLevel(text, '?') && typeTextHasTopLevel(text, ':')
}

func typeTextHasTopLevelArrow(text string) bool {
	depthRound := 0
	depthSquare := 0
	depthCurly := 0
	depthAngle := 0
	for i := 0; i+1 < len(text); i++ {
		c := text[i]
		switch c {
		case '\'', '"', '`':
			quote := c
			i++
			for i < len(text) && text[i] != quote {
				if text[i] == '\\' {
					i++
				}
				i++
			}
		case '(':
			depthRound++
		case ')':
			depthRound--
		case '[':
			depthSquare++
		case ']':
			depthSquare--
		case '{':
			depthCurly++
		case '}':
			depthCurly--
		case '<':
			depthAngle++
		case '>':
			if i > 0 && text[i-1] == '=' {
				break
			}
			depthAngle--
		}

		if depthRound == 0 && depthSquare == 0 && depthCurly == 0 && depthAngle == 0 && text[i] == '=' && text[i+1] == '>' {
			return true
		}
	}
	return false
}
