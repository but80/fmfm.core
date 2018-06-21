package go2cpp

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
)

func (g *generator) iota(reset bool) int {
	if reset {
		g.iotaSequence = 0
	}
	g.iotaSequence++
	return g.iotaSequence - 1
}

func (g *generator) debugInspect(v interface{}, tag string) {
	fmt.Fprint(g.cppWriter, g.debugSInspect(v, tag))
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
		g.debugInspect(expr, "objectOf")
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

func (g *generator) dumpExpr(expr ast.Expr) {
	switch e := expr.(type) {

	case *ast.Ident:
		name, scope, _, _ := g.identOf(e)
		if scope != nil && scope.Parent() == nil {
			switch name {
			case "println", "append", "len", "new", "make", "complex", "real", "imag", "panic":
				fmt.Fprint(g.cppWriter, name)
			case "iota":
				fmt.Fprint(g.cppWriter, g.iota(true))
			case "string", "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64":
				fmt.Fprint(g.cppWriter, "__mg."+name)
			case "true", "false":
				fmt.Fprint(g.cppWriter, name)
			case "nil":
				fmt.Fprint(g.cppWriter, "nullptr")
			default:
				g.debugInspect(scope, "Scope")
				fmt.Fprint(g.cppWriter, name)
			}
		} else {
			fmt.Fprint(g.cppWriter, name)
		}

	case *ast.BasicLit:
		g.dumpLiteral(e)

	case *ast.CallExpr:
		var recv ast.Expr
		skip := false
		comma := false
		switch g.info.TypeOf(e.Fun).(type) {
		case *types.Signature:
			switch f := e.Fun.(type) {
			case *ast.SelectorExpr:
				o, _ := g.objectOf(f.X)
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
			o, _ := g.objectOf(e.Fun)
			switch o.(type) {
			case *types.TypeName:
				fmt.Fprint(g.cppWriter, o.Name())
				fmt.Fprint(g.cppWriter, "(")
				skip = true
			default:
				g.debugInspect(o, "CallExpr")
			}
		}
		if !skip {
			g.dumpExpr(e.Fun)
			fmt.Fprint(g.cppWriter, "(")
		}
		if recv != nil {
			g.dumpExpr(recv)
		}
		for i, a := range e.Args {
			// @todo ポインタでない構造体や構造体以外へのポインタを引数に渡すのは禁止
			if i != 0 || comma {
				fmt.Fprint(g.cppWriter, ", ")
			}
			g.dumpExpr(a)
		}
		fmt.Fprint(g.cppWriter, ")")

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
					fmt.Fprint(g.cppWriter, o.Name())
					fmt.Fprint(g.cppWriter, "::")
					fmt.Fprint(g.cppWriter, e.Sel.Name)
					done = true
				case *types.TypeName:
				case *types.Var:
				default:
					g.debugInspect(o, "SelectorExpr")
				}
			}
			if !done {
				name := g.typedFuncName(e.X, e.Sel.Name)
				fmt.Fprint(g.cppWriter, name)
			}
		default:
			// 変数・フィールドへのアクセス
			g.dumpExpr(e.X)
			fmt.Fprintf(g.cppWriter, "->%s", e.Sel.Name)
		}

	case *ast.BinaryExpr:
		// @todo 演算子の優先順位
		switch e.Op {
		case token.AND:
			g.dumpExpr(e.X)
			fmt.Fprint(g.cppWriter, " & ")
			g.dumpExpr(e.Y)
		case token.OR:
			g.dumpExpr(e.X)
			fmt.Fprint(g.cppWriter, " | ")
			g.dumpExpr(e.Y)
		case token.XOR:
			g.dumpExpr(e.X)
			fmt.Fprint(g.cppWriter, " ^ ")
			g.dumpExpr(e.Y)
		case token.SHL:
			g.dumpExpr(e.X)
			fmt.Fprint(g.cppWriter, " << ")
			g.dumpExpr(e.Y)
		case token.SHR:
			// @todo unsignedに対しては rshift
			g.dumpExpr(e.X)
			fmt.Fprint(g.cppWriter, " >> ")
			g.dumpExpr(e.Y)
		case token.AND_NOT:
			g.dumpExpr(e.X)
			fmt.Fprint(g.cppWriter, "&& !(")
			g.dumpExpr(e.Y)
			fmt.Fprint(g.cppWriter, ")")
		case token.MUL, token.QUO:
			// @todo ビット幅でオーバーフローを考慮
			g.dumpExpr(e.X)
			fmt.Fprint(g.cppWriter, e.Op.String())
			g.dumpExpr(e.Y)
		case token.ADD, token.SUB, token.EQL, token.LSS, token.GTR, token.LEQ, token.GEQ:
			// @todo ビット幅でオーバーフローを考慮
			// @todo 文字列の結合は .. で置換
			g.dumpExpr(e.X)
			fmt.Fprint(g.cppWriter, " ")
			fmt.Fprint(g.cppWriter, e.Op.String())
			fmt.Fprint(g.cppWriter, " ")
			g.dumpExpr(e.Y)
		case token.NEQ:
			g.dumpExpr(e.X)
			fmt.Fprint(g.cppWriter, " != ")
			g.dumpExpr(e.Y)
		case token.LAND:
			g.dumpExpr(e.X)
			fmt.Fprint(g.cppWriter, " && ")
			g.dumpExpr(e.Y)
		case token.LOR:
			g.dumpExpr(e.X)
			fmt.Fprint(g.cppWriter, " || ")
			g.dumpExpr(e.Y)
		default:
			g.dumpExpr(e.X)
			fmt.Fprint(g.cppWriter, " ")
			g.debugInspect(e.Op, "BinaryExpr")
			fmt.Fprint(g.cppWriter, e.Op.String())
			fmt.Fprint(g.cppWriter, " ")
			g.dumpExpr(e.Y)
		}

	case *ast.UnaryExpr:
		// @todo 構造体以外へのポインタを取るのは禁止
		switch e.Op {
		case token.ADD, token.SUB:
			fmt.Fprint(g.cppWriter, e.Op.String())
			g.dumpExpr(e.X)
		case token.NOT:
			fmt.Fprint(g.cppWriter, "!")
			g.dumpExpr(e.X)
		case token.XOR:
			// @todo ビット幅を考慮
			fmt.Fprint(g.cppWriter, "~")
			g.dumpExpr(e.X)
		case token.AND:
			fmt.Fprint(g.cppWriter, "__ptr(")
			g.dumpExpr(e.X)
			fmt.Fprint(g.cppWriter, ")")
		default:
			g.debugInspect(e.Op, "UnaryExpr")
			fmt.Fprint(g.cppWriter, e.Op.String())
			g.dumpExpr(e.X)
		}

	case *ast.ParenExpr:
		fmt.Fprint(g.cppWriter, "(")
		g.dumpExpr(e.X)
		fmt.Fprint(g.cppWriter, ")")

	case *ast.FuncLit:
		g.dumpFunc(g.cppWriter, true, "", nil, e.Type, e.Body)

	case *ast.CompositeLit:
		switch e.Type.(type) {
		case *ast.ArrayType:
			fmt.Fprint(g.cppWriter, "{")
			if 0 < len(e.Elts) {
				fmt.Fprintln(g.cppWriter)
				g.enter()
				for _, v := range e.Elts {
					fmt.Fprint(g.cppWriter, g.indent)
					g.dumpExpr(v)
					fmt.Fprint(g.cppWriter, ",")
					fmt.Fprintln(g.cppWriter)
				}
				g.leave()
				fmt.Fprint(g.cppWriter, g.indent)
			}
			fmt.Fprint(g.cppWriter, "}")
		default:
			fields := []string{}
			n, s, ok := g.formatType(g.info.TypeOf(e))
			fmt.Fprint(g.cppWriter, "(const ")
			if ok {
				fmt.Fprint(g.cppWriter, n)
				fmt.Fprintf(g.cppWriter, s, "")
			} else {
				g.dumpExpr(e)
			}
			fmt.Fprintln(g.cppWriter, "){")
			g.enter()
			for i, v := range e.Elts {
				fmt.Fprint(g.cppWriter, g.indent)
				switch v.(type) {
				case *ast.KeyValueExpr:
					g.dumpExpr(v)
				default:
					if i < len(fields) {
						fmt.Fprintf(g.cppWriter, "%s: ", fields[i])
					}
					g.dumpExpr(v)
				}
				fmt.Fprint(g.cppWriter, ",")
				fmt.Fprintln(g.cppWriter)
			}
			g.leave()
			fmt.Fprint(g.cppWriter, g.indent)
			fmt.Fprint(g.cppWriter, "}")
		}

	case *ast.KeyValueExpr:
		g.dumpExpr(e.Key)
		fmt.Fprint(g.cppWriter, ": ")
		g.dumpExpr(e.Value)

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
					g.debugInspect(it, "IndexExpr2")
					break LOOP_IndexExpr
				}
			case *types.Map:
				// noop
				break LOOP_IndexExpr
			case *types.Named:
				containerType = ct.Underlying()
			default:
				g.debugInspect(ct, "IndexExpr1")
				break LOOP_IndexExpr
			}
		}
		g.dumpExpr(e.X)
		fmt.Fprint(g.cppWriter, "[")
		g.dumpExpr(e.Index)
		fmt.Fprint(g.cppWriter, "]")

	case *ast.SliceExpr:
	//e.

	default:
		g.debugInspect(expr, "expr")
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
