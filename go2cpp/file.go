package go2cpp

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io"
	"strings"
)

type generator struct {
	cppWriter             io.Writer
	h0Writer              io.Writer
	h1Writer              io.Writer
	hWriter               io.Writer
	fileAST               *ast.File
	basePkg               string
	pkg                   *types.Package
	info                  *types.Info
	fset                  *token.FileSet
	indent                string
	imports               map[string]string
	Warnings              []string
	Exports               []string
	iotaSequence          int
	currentFunc           []string
	currentFuncResultType []string
	currentCall           []string
}

func newGenerator(cppWriter, h0Writer, h1Writer, hWriter io.Writer, fileAST *ast.File, basePkg string, pkg *types.Package, info *types.Info, fset *token.FileSet, imports map[string]string) *generator {
	return &generator{
		cppWriter: cppWriter,
		h0Writer:  h0Writer,
		h1Writer:  h1Writer,
		hWriter:   hWriter,
		fileAST:   fileAST,
		basePkg:   basePkg,
		pkg:       pkg,
		info:      info,
		fset:      fset,
		imports:   imports,
	}
}

func (g *generator) export(symbol string) {
	if g.indent == "" && strings.ToUpper(symbol[:1]) == symbol[:1] {
		g.Exports = append(g.Exports, symbol)
	}
}

func (g *generator) Dump() {
	// for _, imp := range g.fileAST.Imports {
	// 	// @todo ドットインポート
	// 	name, path, relPath, _ := g.importNameAndPath(imp)
	// 	g.imports[path] = name
	// }

	for _, decl := range g.fileAST.Decls {
		g.dumpDecl(decl)
		fmt.Fprintln(g.cppWriter)
	}
}

func (g *generator) localPkgPrefix(pkg string) string {
	if pkg == g.pkg.Path() {
		return ""
	}
	if name, ok := g.imports[pkg]; ok {
		return name + "::"
	}
	return nonAlphaNumRe.ReplaceAllStringFunc(pkg, func(s string) string {
		switch s {
		case "/":
			return "::"
		case ".":
			return "_"
		default:
			return s
		}
	}) + "::"
}
