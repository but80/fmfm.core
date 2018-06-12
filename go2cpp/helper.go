package go2cpp

import (
	"fmt"
	"go/ast"
)

func (g *generator) warn(node ast.Node, msg string) {
	p := g.fset.Position(node.Pos())
	g.Warnings = append(g.Warnings, fmt.Sprintf("%s(%d:%d): %s", p.Filename, p.Line, p.Column, msg))
}

func (g *generator) enter() {
	g.indent += "\t"
}

func (g *generator) leave() {
	g.indent = g.indent[0 : len(g.indent)-1]
}
