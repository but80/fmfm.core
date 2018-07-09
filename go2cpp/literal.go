package go2cpp

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/token"
	"io"
	"strconv"
)

func (g *generator) stringLiteral(s string) string {
	j, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	return string(j)
}

func (g *generator) dumpLiteral(writer io.Writer, lit *ast.BasicLit) {
	switch lit.Kind {

	case token.STRING:
		s, err := strconv.Unquote(lit.Value)
		if err != nil {
			panic(err)
		}
		fmt.Fprintf(writer, "string(%s)", g.stringLiteral(s))

	case token.CHAR:
		s, err := strconv.Unquote(lit.Value)
		if err != nil {
			panic(err)
		}
		fmt.Fprintf(writer, "string.byte(%s)", g.stringLiteral(s))

	case token.INT, token.FLOAT:
		fmt.Fprint(writer, lit.Value)

	case token.IMAG:
		fmt.Fprintf(writer, "complex(0, %s)", lit.Value[:len(lit.Value)-1])

	default:
		g.debugInspect(writer, lit, "literal")
	}
}
