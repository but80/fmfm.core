package go2cpp

import (
	"fmt"
	"go/ast"
	"io"
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

func (g *generator) dumpTypeAndName(writer io.Writer, typ ast.Expr, name string) {
	n, s, ok := g.formatType(g.info.TypeOf(typ))
	if ok {
		fmt.Fprintf(writer, "%s "+s, n, name)
	} else {
		fmt.Fprintf(writer, "UNKNOWN %s", name)
	}
}

func (g *generator) dumpFunc(writer io.Writer, withBody bool, name string, recv *ast.FieldList, sig *ast.FuncType, body *ast.BlockStmt) {
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

	fmt.Fprint(writer, g.indent)

	if sig.Results != nil && len(sig.Results.List) == 1 {
		g.dumpTypeAndName(writer, sig.Results.List[0].Type, fnname)
	} else if sig.Results != nil && 2 <= len(sig.Results.List) {
		fmt.Fprintf(writer, "MULTIRESULT %s", fnname)
	} else {
		fmt.Fprintf(writer, "void %s", fnname)
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
	if hasNamedResult {
		g.enter()
		fmt.Fprintf(writer, "%slocal ", g.indent)
		for i, r := range sig.Results.List {
			if i != 0 {
				fmt.Fprint(writer, ", ")
			}
			fmt.Fprint(writer, r.Names[0].Name)
		}
		fmt.Fprintf(writer, "\n%slocal __r = pack((function()\n", g.indent)
	}
	g.enter()
	g.dumpBlock(body)
	g.leave()
	if hasNamedResult {
		fmt.Fprint(writer, g.indent)
		fmt.Fprintln(writer, "end)())")
		fmt.Fprint(writer, g.indent)
		fmt.Fprintln(writer, "if 0 < #__r then")
		g.enter()
		fmt.Fprint(writer, g.indent)
		for i, r := range sig.Results.List {
			if i != 0 {
				fmt.Fprint(writer, ", ")
			}
			fmt.Fprint(writer, r.Names[0].Name)
		}
		fmt.Fprintln(writer, " = unpack(__r)")
		g.leave()
		fmt.Fprint(writer, g.indent)
		fmt.Fprintln(writer, "end")
		fmt.Fprint(writer, g.indent)
		fmt.Fprint(writer, "return ")
		for i, r := range sig.Results.List {
			if i != 0 {
				fmt.Fprint(writer, ", ")
			}
			fmt.Fprint(writer, r.Names[0].Name)
		}
		fmt.Fprintln(writer)
		g.leave()
	}
	fmt.Fprint(writer, g.indent)
	fmt.Fprint(writer, "}")
}
