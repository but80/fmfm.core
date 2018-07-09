package go2cpp

import (
	"fmt"
	"go/ast"
	"go/types"
	"io"
	"strconv"
)

func (g *generator) dumpFuncSig(writer io.Writer, fntype *ast.FuncType, firstComma bool) {
	for _, p := range fntype.Params.List {
		for _, name := range p.Names {
			if firstComma {
				fmt.Fprint(writer, ", ")
			}
			g.dumpTypeAndName(writer, p.Type, name.Name)
			firstComma = true
		}
	}
}

func (g *generator) typedFuncName(typ ast.Expr, name string) string {
	return fmt.Sprintf("%s__%s", g.typeIdent(g.info.TypeOf(typ)), name)
}

func (g *generator) dumpTypeAndName(writer io.Writer, typ ast.Expr, name string) string {
	return g.dumpTypeAndNameImpl(writer, g.info.TypeOf(typ), name)
}

func (g *generator) dumpTypeAndNameImpl(writer io.Writer, typ types.Type, name string) string {
	n, s, ok := g.formatType(typ)
	if !ok {
		fmt.Fprintf(writer, "UNKNOWN %s", name)
		return "UNKNOWN"
	}
	fmt.Fprintf(writer, "%s "+s, n, name)
	return fmt.Sprintf("%s "+s, n, "")
}

func (g *generator) dumpFunc(writer io.Writer, withBody bool, name string, recv *ast.FieldList, sig *ast.FuncType, body *ast.BlockStmt) {
	returnsByStruct := false
	resultNames := []*ast.Ident{}
	if sig.Results != nil {
		n := 0
		for _, r := range sig.Results.List {
			if 0 < len(r.Names) {
				n += len(r.Names)
				returnsByStruct = true
				resultNames = append(resultNames, r.Names...)
			} else {
				n++
				resultNames = append(resultNames, nil)
			}
		}
		if 2 <= n {
			returnsByStruct = true
		}
	}

	fnname := "UNKNOWN"
	if name != "" {
		if recv != nil && 0 < len(recv.List) {
			fnname = g.typedFuncName(recv.List[0].Type, name)
		} else {
			fnname = name
			g.export(name)
		}
	}

	resultType := "void"

	if returnsByStruct {
		resultType = fnname + "__result"
		if !withBody {
			fmt.Fprintf(writer, "%sstruct %s {\n", g.indent, resultType)
			g.enter()
			for i, r := range sig.Results.List {
				if len(r.Names) == 0 {
					fmt.Fprint(writer, g.indent)
					g.dumpTypeAndName(writer, r.Type, "r"+strconv.Itoa(i))
					fmt.Fprintln(writer, ";")
				} else {
					for _, n := range r.Names {
						fmt.Fprint(writer, g.indent)
						g.dumpTypeAndName(writer, r.Type, n.Name)
						fmt.Fprintln(writer, ";")
					}
				}
			}
			g.leave()
			fmt.Fprintf(writer, "%s};\n", g.indent)
		}
		fmt.Fprint(writer, g.indent)
		fmt.Fprintf(writer, "%s %s", resultType, fnname)
	} else if sig.Results == nil || len(sig.Results.List) == 0 {
		fmt.Fprint(writer, g.indent)
		fmt.Fprintf(writer, "void %s", fnname)
	} else {
		fmt.Fprint(writer, g.indent)
		resultType = g.dumpTypeAndName(writer, sig.Results.List[0].Type, fnname)
	}

	fmt.Fprint(writer, "(")
	if recv != nil && 0 < len(recv.List) {
		comma := false
		for _, r := range recv.List {
			for _, n := range r.Names {
				if comma {
					fmt.Fprint(writer, ", ")
				}
				comma = true
				g.dumpTypeAndName(writer, r.Type, n.Name)
			}
		}
		g.dumpFuncSig(writer, sig, comma)
	} else {
		g.dumpFuncSig(writer, sig, false)
	}
	if !withBody {
		fmt.Fprintln(writer, ");")
		return
	}
	fmt.Fprintln(writer, ") {")
	g.enter()
	if returnsByStruct {
		fmt.Fprintf(writer, "%s%s __result; // multi-result\n", g.indent, resultType)
		for i, r := range resultNames {
			if r != nil {
				fmt.Fprintf(writer, "%sauto %s = &__result.r%d; // multi-result\n", g.indent, r, i)
			}
		}
	}
	g.currentFunc = append(g.currentFunc, fnname)
	g.currentFuncResultType = append(g.currentFuncResultType, resultType)
	g.dumpBlock(body)
	g.currentFuncResultType = g.currentFuncResultType[:len(g.currentFuncResultType)-1]
	g.currentFunc = g.currentFunc[:len(g.currentFunc)-1]
	g.leave()
	fmt.Fprint(writer, g.indent)
	fmt.Fprint(writer, "}")
}
