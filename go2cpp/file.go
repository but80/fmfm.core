package go2cpp

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type generator struct {
	cppWriter    io.Writer
	h0Writer     io.Writer
	h1Writer     io.Writer
	hWriter      io.Writer
	fileAST      *ast.File
	basePkg      string
	pkg          *types.Package
	info         *types.Info
	fset         *token.FileSet
	indent       string
	imports      map[string]string
	Warnings     []string
	Exports      []string
	iotaSequence int
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

func translateNamespace(ns string) string {
	ns = strings.Replace(ns, "/", "::", -1)
	ns = strings.Replace(ns, ".", "_", -1)
	return ns
}

var importNameAndPathRe = regexp.MustCompile(`.+/`)

func (g *generator) importNameAndPath(imp *ast.ImportSpec) (string, string, string, string) {
	path, err := strconv.Unquote(imp.Path.Value)
	if err != nil {
		panic(err)
	}
	relPath := relativePkg(path, g.basePkg)
	name := importNameAndPathRe.ReplaceAllString(relPath, "")
	if imp.Name != nil {
		name = imp.Name.Name
	}
	return name, path, relPath, translateNamespace(relPath)
}

func (g *generator) dumpImportSpec(imp *ast.ImportSpec) {
	name, path, relPath, ns := g.importNameAndPath(imp)
	if _, ok := g.imports[path]; ok {
		// TODO: 異なる名前でのインポート時はエラー
		return
	}
	fmt.Fprint(g.h0Writer, g.indent)
	fmt.Fprintf(g.h0Writer, "#include \"%s.h\"\n", relPath)
	if name != ns {
		fmt.Fprintf(g.h0Writer, "namespace %s = %s;\n", name, ns)
	}
	// g.imports[relPath] = name
	g.imports[path] = name
}

func (g *generator) dumpValueSpec(s *ast.ValueSpec) int {
	for i, n := range s.Names {
		if i != 0 {
			fmt.Fprint(g.cppWriter, ", ")
		}
		g.dumpTypeAndName(g.cppWriter, s.Type, n.Name)
		if i < len(s.Values) {
			fmt.Fprint(g.cppWriter, " = ")
			g.dumpExpr(g.cppWriter, s.Values[i])
		}
		g.export(n.Name)
	}
	return len(s.Values)
}

func (g *generator) dumpSpec(spec ast.Spec) {
	switch s := spec.(type) {

	case *ast.ImportSpec:
		g.dumpImportSpec(s)

	case *ast.ValueSpec:
		g.dumpValueSpec(s)

	case *ast.TypeSpec:
		g.dumpTypeSpec(s)

	default:
		g.debugInspect(g.cppWriter, spec, "spec")
	}
}

func (g *generator) dumpTypeSpec(typ *ast.TypeSpec) {
	if typ.Doc != nil {
		for _, c := range typ.Doc.List {
			fmt.Fprint(g.hWriter, g.indent)
			fmt.Fprintln(g.hWriter, c.Text)
		}
	}
	switch t := typ.Type.(type) {
	case *ast.StructType:
		if typ.Comment != nil {
			for _, c := range typ.Comment.List {
				fmt.Fprintf(g.hWriter, "%s// %s\n", g.indent, c.Text)
			}
		}
		fmt.Fprintf(g.h1Writer, "class %s;\n", typ.Name.Name)
		fmt.Fprintf(g.hWriter, "%sclass %s {\n", g.indent, typ.Name.Name)
		fmt.Fprintf(g.hWriter, "%spublic:\n", g.indent)
		g.enter()
		for _, field := range t.Fields.List {
			for _, name := range field.Names {
				fmt.Fprint(g.hWriter, g.indent)
				g.dumpTypeAndName(g.hWriter, field.Type, name.Name)
				fmt.Fprintln(g.hWriter, ";")
			}
		}
		g.leave()
		fmt.Fprintf(g.hWriter, "%s};\n", g.indent)
	case *ast.Ident:
		fmt.Fprintf(g.h1Writer, "%stypedef %s %s;\n", g.indent, t.Name, typ.Name.Name)
	default:
		fmt.Fprintf(g.h1Writer, "%s// %s\n", g.indent, reflect.TypeOf(t))
	}
	g.export(typ.Name.Name)
}

func (g *generator) dumpDecl(decl ast.Decl) {
	switch d := decl.(type) {

	case *ast.FuncDecl:
		if d.Doc != nil {
			for _, c := range d.Doc.List {
				fmt.Fprint(g.cppWriter, g.indent)
				fmt.Fprintln(g.cppWriter, c.Text)
			}
		}
		fmt.Fprint(g.cppWriter, g.indent)
		g.dumpFunc(g.cppWriter, true, d.Name.Name, d.Recv, d.Type, d.Body)
		g.dumpFunc(g.hWriter, false, d.Name.Name, d.Recv, d.Type, d.Body)
		fmt.Fprintln(g.cppWriter)

	case *ast.GenDecl:
		switch d.Tok {
		case token.VAR:
			for _, s := range d.Specs {
				fmt.Fprint(g.cppWriter, g.indent)
				g.dumpSpec(s)
				fmt.Fprintln(g.cppWriter, ";")
			}
		case token.IMPORT:
			fmt.Fprint(g.cppWriter, g.indent)
			for _, s := range d.Specs {
				g.dumpSpec(s)
			}
		case token.TYPE:
			fmt.Fprintf(g.hWriter, "/* %#v */\n", g.imports)
			for _, s := range d.Specs {
				g.dumpSpec(s)
			}
		case token.CONST:
			for _, s := range d.Specs {
				vs, ok := s.(*ast.ValueSpec)
				for i, n := range vs.Names {
					// @todo iota
					fmt.Fprint(g.hWriter, g.indent)
					if !ok {
						panic("const must have only ValueSpecs")
					}
					n2, s, ok := g.formatType(g.info.TypeOf(vs.Type))
					if ok && n2 != "auto" {
						fmt.Fprintf(g.hWriter, "const %s "+s+" = ", n2, n.Name)
					} else {
						fmt.Fprintf(g.hWriter, "const auto %s = ", n.Name)
					}
					if i < len(vs.Values) {
						g.dumpExpr(g.hWriter,vs.Values[i])
					} else {
						fmt.Fprintf(g.hWriter, "%d", g.iota(false))
					}
					fmt.Fprintln(g.hWriter, ";")
					g.export(n.Name)
				}
			}
		default:
			fmt.Fprint(g.cppWriter, g.indent)
			g.debugInspect(g.cppWriter, d, "decl:"+d.Tok.String())
			fmt.Fprintln(g.cppWriter)
		}

	default:
		fmt.Fprint(g.cppWriter, g.indent)
		g.debugInspect(g.cppWriter, decl, "decl")
		fmt.Fprintln(g.cppWriter)
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
