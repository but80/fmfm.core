package go2cpp

import (
	"fmt"
	"go/ast"
	"io"
)

func (g *generator) dumpFuncSig(fntype *ast.FuncType, firstComma bool) {
	for _, p := range fntype.Params.List {
		for _, name := range p.Names {
			if firstComma {
				fmt.Fprint(g.cppWriter, ", ")
			}
			g.dumpTypeAndName(g.cppWriter, p.Type, name.Name)
			firstComma = true
		}
	}
}

func (g *generator) typedFuncName(typ ast.Expr, name string) string {
	return fmt.Sprintf("%s__%s", g.typeIdent(g.info.TypeOf(typ)), name)
}

func (g *generator) dumpTypeAndName(writer io.Writer, typ ast.Expr, name string) {
	n, s, ok := g.formatType(g.info.TypeOf(typ))
	if ok {
		fmt.Fprintf(writer, "%s "+s, n, name)
	} else {
		fmt.Fprintf(writer, "UNKNOWN %s", name)
	}
}

func (g *generator) dumpFunc(name string, recv *ast.FieldList, sig *ast.FuncType, body *ast.BlockStmt) {
	hasNamedResult := false
	if sig.Results != nil {
		for _, r := range sig.Results.List {
			if 0 < len(r.Names) {
				hasNamedResult = true
				break
			}
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

	fmt.Fprint(g.cppWriter, g.indent)

	if sig.Results != nil && len(sig.Results.List) == 1 {
		g.dumpTypeAndName(g.cppWriter, sig.Results.List[0].Type, fnname)
	} else if sig.Results != nil && 2 <= len(sig.Results.List) {
		fmt.Fprintf(g.cppWriter, "MULTIRESULT %s", fnname)
	} else {
		fmt.Fprintf(g.cppWriter, "void %s", fnname)
	}

	fmt.Fprint(g.cppWriter, "(")
	if recv != nil && 0 < len(recv.List) {
		comma := false
		for _, r := range recv.List {
			for _, n := range r.Names {
				if comma {
					fmt.Fprint(g.cppWriter, ", ")
				}
				comma = true
				g.dumpTypeAndName(g.cppWriter, r.Type, n.Name)
			}
		}
		g.dumpFuncSig(sig, comma)
	} else {
		g.dumpFuncSig(sig, false)
	}
	fmt.Fprintln(g.cppWriter, ") {")
	if hasNamedResult {
		g.enter()
		fmt.Fprintf(g.cppWriter, "%slocal ", g.indent)
		for i, r := range sig.Results.List {
			if i != 0 {
				fmt.Fprint(g.cppWriter, ", ")
			}
			fmt.Fprint(g.cppWriter, r.Names[0].Name)
		}
		fmt.Fprintf(g.cppWriter, "\n%slocal __r = pack((function()\n", g.indent)
	}
	g.enter()
	g.dumpBlock(body)
	g.leave()
	if hasNamedResult {
		fmt.Fprint(g.cppWriter, g.indent)
		fmt.Fprintln(g.cppWriter, "end)())")
		fmt.Fprint(g.cppWriter, g.indent)
		fmt.Fprintln(g.cppWriter, "if 0 < #__r then")
		g.enter()
		fmt.Fprint(g.cppWriter, g.indent)
		for i, r := range sig.Results.List {
			if i != 0 {
				fmt.Fprint(g.cppWriter, ", ")
			}
			fmt.Fprint(g.cppWriter, r.Names[0].Name)
		}
		fmt.Fprintln(g.cppWriter, " = unpack(__r)")
		g.leave()
		fmt.Fprint(g.cppWriter, g.indent)
		fmt.Fprintln(g.cppWriter, "end")
		fmt.Fprint(g.cppWriter, g.indent)
		fmt.Fprint(g.cppWriter, "return ")
		for i, r := range sig.Results.List {
			if i != 0 {
				fmt.Fprint(g.cppWriter, ", ")
			}
			fmt.Fprint(g.cppWriter, r.Names[0].Name)
		}
		fmt.Fprintln(g.cppWriter)
		g.leave()
	}
	fmt.Fprint(g.cppWriter, g.indent)
	fmt.Fprint(g.cppWriter, "}")
}
