package go2cpp

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
)

func (g *generator) stringLiteral(s string) string {
	j, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	return string(j)
}

func (g *generator) dumpLiteral(lit *ast.BasicLit) {
	switch lit.Kind {

	case token.STRING:
		s, err := strconv.Unquote(lit.Value)
		if err != nil {
			panic(err)
		}
		fmt.Fprint(g.cppWriter, g.stringLiteral(s))

	case token.CHAR:
		s, err := strconv.Unquote(lit.Value)
		if err != nil {
			panic(err)
		}
		fmt.Fprintf(g.cppWriter, "string.byte(%s)", g.stringLiteral(s))

	case token.INT, token.FLOAT:
		fmt.Fprint(g.cppWriter, lit.Value)

	case token.IMAG:
		fmt.Fprintf(g.cppWriter, "complex(0, %s)", lit.Value[:len(lit.Value)-1])

	default:
		g.debugInspect(lit, "literal")
	}
}
