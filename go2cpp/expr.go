package go2cpp

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io"
	"reflect"
	"regexp"
)

func (g *generator) iota(reset bool) int {
	if reset {
		g.iotaSequence = 0
	}
	g.iotaSequence++
	return g.iotaSequence - 1
}

func (g *generator) debugInspect(writer io.Writer, v interface{}, tag string) {
	fmt.Fprint(writer, g.debugSInspect(v, tag))
}

func (g *generator) debugSInspect(v interface{}, tag string) string {
	return fmt.Sprintf("\x1b[41m%s\x1b[0m\x1b[31m<%T>(%#v)\x1b[0m", tag, v, v)
}

func (g *generator) objectOf(expr ast.Expr) (types.Object, bool) {
	switch e := expr.(type) {
	case *ast.Ident:
		return g.info.ObjectOf(e), true
	case *ast.ParenExpr:
		return g.objectOf(e.X)
	case *ast.StarExpr:
		return g.objectOf(e.X)
	case *ast.SelectorExpr:
		return nil, false
	default:
		g.debugInspect(g.cppWriter, expr, "objectOf")
		return nil, false
	}
}

func (g *generator) identOf(expr ast.Expr) (string, *types.Scope, types.Object, bool) {
	switch e := expr.(type) {

	case *ast.Ident:
		scope, object := g.pkg.Scope().Innermost(e.Pos()).LookupParent(e.Name, e.Pos())
		return e.Name, scope, object, true

	default:
		return "", nil, nil, false
	}
}

var isInPrintfRe = regexp.MustCompile(`[pP]rintf\W*$`)

func (g *generator) isInPrintf() bool {
	n := len(g.currentCall)
	return 0 < n && isInPrintfRe.MatchString(g.currentCall[n-1])
}

var isInMakeRe = regexp.MustCompile(`^\W*make\W*$`)

func (g *generator) isInMake() bool {
	n := len(g.currentCall)
	return 0 < n && isInMakeRe.MatchString(g.currentCall[n-1])
}

func (g *generator) dumpExpr(writer io.Writer, expr ast.Expr) {
	switch e := expr.(type) {

	case *ast.Ident:
		name, scope, _, _ := g.identOf(e)
		if scope != nil && scope.Parent() == nil {
			switch name {
			case "println", "append", "len", "new", "make", "complex", "real", "imag", "panic":
				fmt.Fprint(writer, name)
			case "iota":
				fmt.Fprint(writer, g.iota(true))
			case "string", "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64":
				fmt.Fprint(writer, "__mg."+name)
			case "true", "false":
				fmt.Fprint(writer, name)
			case "nil":
				fmt.Fprint(writer, "nullptr")
			default:
				g.debugInspect(writer, scope, "Scope")
				fmt.Fprint(writer, name)
			}
		} else {
			fmt.Fprint(writer, name)
		}
		if g.isInPrintf() && g.isStringType(g.info.TypeOf(e)) {
			fmt.Fprint(writer, ".c_str()")
		}

	case *ast.BasicLit:
		g.dumpLiteral(writer, e)

	case *ast.CallExpr:
		var recv ast.Expr
		skip := false
		comma := false
		var o types.Object
		switch g.info.TypeOf(e.Fun).(type) {
		case *types.Signature:
			switch f := e.Fun.(type) {
			case *ast.SelectorExpr:
				o, _ = g.objectOf(f.X)
				switch o.(type) {
				case *types.TypeName:
					// noop
				case *types.PkgName:
					// noop
				default:
					// ドットの左辺が式なのでレシーバを第1引数に移動
					recv = e.Fun.(*ast.SelectorExpr).X
					comma = true
				}
			}
		default:
			o, _ = g.objectOf(e.Fun)
			switch o.(type) {
			case *types.TypeName:
				fmt.Fprint(writer, o.Name())
				fmt.Fprint(writer, "(")
				skip = true
			default:
				g.debugInspect(writer, o, "CallExpr")
			}
		}
		if !skip {
			g.dumpExpr(writer, e.Fun)
			fmt.Fprint(writer, "(")
		}
		fnName := fmt.Sprint(e.Fun)
		// fmt.Fprintf(writer, "/*%s*/", fnName)
		g.currentCall = append(g.currentCall, fnName)
		if recv != nil {
			g.dumpExpr(writer, recv)
		}
		for i, a := range e.Args {
			// @todo ポインタでない構造体や構造体以外へのポインタを引数に渡すのは禁止
			if i != 0 || comma {
				fmt.Fprint(writer, ", ")
			}
			g.dumpExpr(writer, a)
		}
		g.currentCall = g.currentCall[:len(g.currentCall)-1]
		fmt.Fprint(writer, ")")

	case *ast.SelectorExpr:
		// ドット演算子
		switch g.info.TypeOf(e).(type) {
		case *types.Signature:
			// 関数へのアクセス
			done := false
			o, ok := g.objectOf(e.X)
			if ok {
				switch o.(type) {
				case *types.PkgName:
					fmt.Fprint(writer, o.Name())
					fmt.Fprint(writer, "::")
					fmt.Fprint(writer, e.Sel.Name)
					done = true
				case *types.TypeName:
				case *types.Var:
				default:
					g.debugInspect(writer, o, "SelectorExpr")
				}
			}
			if !done {
				name := g.typedFuncName(e.X, e.Sel.Name)
				fmt.Fprint(writer, name)
			}
		default:
			// 変数・フィールドへのアクセス
			g.dumpExpr(writer, e.X)
			switch g.info.TypeOf(e.X).(type) {
			case *types.Basic:
				fmt.Fprint(writer, "::")
			case *types.Pointer:
				fmt.Fprint(writer, "->")
			default:
				fmt.Fprintf(writer, "/*%s*/->", reflect.TypeOf(g.info.TypeOf(e.X)))
			}
			fmt.Fprint(writer, e.Sel.Name)
			if g.isInPrintf() && g.isStringType(g.info.TypeOf(e)) {
				fmt.Fprint(writer, ".c_str()")
			}
		}

	case *ast.BinaryExpr:
		// @todo 演算子の優先順位
		switch e.Op {
		case token.AND:
			g.dumpExpr(writer, e.X)
			fmt.Fprint(writer, " & ")
			g.dumpExpr(writer, e.Y)
		case token.OR:
			g.dumpExpr(writer, e.X)
			fmt.Fprint(writer, " | ")
			g.dumpExpr(writer, e.Y)
		case token.XOR:
			g.dumpExpr(writer, e.X)
			fmt.Fprint(writer, " ^ ")
			g.dumpExpr(writer, e.Y)
		case token.SHL:
			g.dumpExpr(writer, e.X)
			fmt.Fprint(writer, " << ")
			g.dumpExpr(writer, e.Y)
		case token.SHR:
			// @todo unsignedに対しては rshift
			g.dumpExpr(writer, e.X)
			fmt.Fprint(writer, " >> ")
			g.dumpExpr(writer, e.Y)
		case token.AND_NOT:
			g.dumpExpr(writer, e.X)
			fmt.Fprint(writer, "&& !(")
			g.dumpExpr(writer, e.Y)
			fmt.Fprint(writer, ")")
		case token.MUL, token.QUO:
			// @todo ビット幅でオーバーフローを考慮
			g.dumpExpr(writer, e.X)
			fmt.Fprint(writer, e.Op.String())
			g.dumpExpr(writer, e.Y)
		case token.ADD, token.SUB, token.EQL, token.LSS, token.GTR, token.LEQ, token.GEQ:
			// @todo ビット幅でオーバーフローを考慮
			// @todo 文字列の結合は .. で置換
			g.dumpExpr(writer, e.X)
			fmt.Fprint(writer, " ")
			fmt.Fprint(writer, e.Op.String())
			fmt.Fprint(writer, " ")
			g.dumpExpr(writer, e.Y)
		case token.NEQ:
			g.dumpExpr(writer, e.X)
			fmt.Fprint(writer, " != ")
			g.dumpExpr(writer, e.Y)
		case token.LAND:
			g.dumpExpr(writer, e.X)
			fmt.Fprint(writer, " && ")
			g.dumpExpr(writer, e.Y)
		case token.LOR:
			g.dumpExpr(writer, e.X)
			fmt.Fprint(writer, " || ")
			g.dumpExpr(writer, e.Y)
		default:
			g.dumpExpr(writer, e.X)
			fmt.Fprint(writer, " ")
			g.debugInspect(writer, e.Op, "BinaryExpr")
			fmt.Fprint(writer, e.Op.String())
			fmt.Fprint(writer, " ")
			g.dumpExpr(writer, e.Y)
		}

	case *ast.UnaryExpr:
		// @todo 構造体以外へのポインタを取るのは禁止
		switch e.Op {
		case token.ADD, token.SUB:
			fmt.Fprint(writer, e.Op.String())
			g.dumpExpr(writer, e.X)
		case token.NOT:
			fmt.Fprint(writer, "!")
			g.dumpExpr(writer, e.X)
		case token.XOR:
			// @todo ビット幅を考慮
			fmt.Fprint(writer, "~")
			g.dumpExpr(writer, e.X)
		case token.AND:
			fmt.Fprint(writer, "__ptr(")
			g.dumpExpr(writer, e.X)
			fmt.Fprint(writer, ")")
		default:
			g.debugInspect(writer, e.Op, "UnaryExpr")
			fmt.Fprint(writer, e.Op.String())
			g.dumpExpr(writer, e.X)
		}

	case *ast.ParenExpr:
		fmt.Fprint(writer, "(")
		g.dumpExpr(writer, e.X)
		fmt.Fprint(writer, ")")

	case *ast.FuncLit:
		g.dumpFunc(writer, true, "", nil, e.Type, e.Body)

	case *ast.CompositeLit:
		switch e.Type.(type) {
		case *ast.ArrayType:
			fmt.Fprint(writer, "{")
			if 0 < len(e.Elts) {
				fmt.Fprintln(writer)
				g.enter()
				for _, v := range e.Elts {
					fmt.Fprint(writer, g.indent)
					g.dumpExpr(writer, v)
					fmt.Fprint(writer, ",")
					fmt.Fprintln(writer)
				}
				g.leave()
				fmt.Fprint(writer, g.indent)
			}
			fmt.Fprint(writer, "}")
		default:
			n, s, ok := g.formatType(g.info.TypeOf(e))
			fmt.Fprint(writer, "(const ")
			if ok {
				fmt.Fprint(writer, n)
				fmt.Fprintf(writer, s, "")
			} else {
				g.dumpExpr(writer, e)
			}
			fmt.Fprintln(writer, "){")
			g.enter()
			byKey := false
			isStruct := g.isStructType(e)
			fieldByKey := map[string]*ast.KeyValueExpr{}
			for _, v := range e.Elts {
				switch vv := v.(type) {
				case *ast.KeyValueExpr:
					if isStruct {
						byKey = true
						key := fmt.Sprintf("%s", vv.Key)
						fieldByKey[key] = vv
					} else {
						fmt.Fprint(writer, g.indent)
						g.dumpExpr(writer, v)
						fmt.Fprintln(writer, ",")
					}
				default:
					fmt.Fprint(writer, g.indent)
					g.dumpExpr(writer, v)
					fmt.Fprintln(writer, ",")
				}
			}
			if byKey {
				if strc, ok := g.structInfo(g.info.TypeOf(e)); ok {
					for i := 0; i < strc.NumFields(); i++ {
						f := strc.Field(i)
						fmt.Fprint(writer, g.indent)
						if kv, ok := fieldByKey[f.Name()]; ok {
							g.dumpExpr(writer, kv.Value)
						} else {
							g.dumpZeroValue(writer, f.Type())
						}
						fmt.Fprintln(writer, ",")
					}
				} else {
					fmt.Fprintf(writer, "%s//%#v | %#v\n", g.indent, strc, fieldByKey)
				}
			}
			g.leave()
			fmt.Fprint(writer, g.indent)
			fmt.Fprint(writer, "}")
		}

	case *ast.KeyValueExpr:
		g.dumpExpr(writer, e.Key)
		fmt.Fprint(writer, ": ")
		g.dumpExpr(writer, e.Value)

	case *ast.IndexExpr:
		containerType := g.info.TypeOf(e.X)
		indexType := g.info.TypeOf(e.Index)
	LOOP_IndexExpr:
		for {
			switch ct := containerType.(type) {
			case *types.Slice, *types.Array:
				switch it := indexType.(type) {
				case *types.Basic:
					break LOOP_IndexExpr
				case *types.Named:
					indexType = it.Underlying()
				default:
					g.debugInspect(writer, it, "IndexExpr2")
					break LOOP_IndexExpr
				}
			case *types.Map:
				// noop
				break LOOP_IndexExpr
			case *types.Named:
				containerType = ct.Underlying()
			default:
				g.debugInspect(writer, ct, "IndexExpr1")
				break LOOP_IndexExpr
			}
		}
		g.dumpExpr(writer, e.X)
		fmt.Fprint(writer, "[")
		g.dumpExpr(writer, e.Index)
		fmt.Fprint(writer, "]")

	case *ast.SliceExpr:
	//e.

	case *ast.ArrayType:
		if g.isInMake() {
			fmt.Fprint(writer, "(")
		}
		g.dumpTypeAndName(writer, e.Elt, "")
		// g.debugInspect(writer, g.info.TypeOf(e), "ArrayType")
		if g.isInMake() {
			fmt.Fprint(writer, "*)NULL")
		}

	default:
		g.debugInspect(writer, expr, "expr")
	}

	//t := g.info.TypeOf(expr)
	//if t != nil {
	//	fmt.Fprintf(g.writer, "(%s)", t.String())
	//}
}

func (g *generator) structInfo(typ types.Type) (*types.Struct, bool) {
	switch t := typ.(type) {
	case *types.Named:
		return g.structInfo(t.Underlying())
	case *types.Struct:
		return t, true
	default:
		return nil, false
	}
}
