package go2cpp

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type analyzedFile struct {
	fileName string
	fileAST  *ast.File
}

// Package は、Goのパッケージです。
type Package struct {
	pkgPath string
	dir     string
	fset    *token.FileSet
	files   map[string]*analyzedFile
	pkg     *types.Package
	info    *types.Info
}

// NewPackage は、新しい Package を作成します。
func NewPackage(pkgPath string) *Package {
	return &Package{
		pkgPath: pkgPath,
		fset:    token.NewFileSet(),
		files:   map[string]*analyzedFile{},
		info: &types.Info{
			Selections: make(map[*ast.SelectorExpr]*types.Selection),
			Types:      make(map[ast.Expr]types.TypeAndValue),
			Defs:       make(map[*ast.Ident]types.Object),
			Uses:       make(map[*ast.Ident]types.Object),
		},
	}
}

func relativePkg(pkg, basePkg string) string {
	if strings.HasPrefix(pkg, basePkg+"/") {
		return pkg[len(basePkg)+1:]
	}
	return pkg
}

func (p *Package) listSourceFiles() (dir string, files []os.FileInfo, err error) {
	pkgPath := filepath.FromSlash(p.pkgPath)
	for _, gopath := range filepath.SplitList(build.Default.GOPATH) {
		dir := filepath.Join(gopath, "src", pkgPath)
		info, err := os.Stat(dir)
		if err != nil {
			continue
		}
		if !info.IsDir() {
			continue
		}
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			continue
		}
		return dir, files, nil
	}
	return "", nil, fmt.Errorf("Source file of package %s was not found", p.pkg)
}

// Load は、Goのソースコードからパッケージの内容を読み込みます。
func (p *Package) Load() error {
	dir, files, err := p.listSourceFiles()
	p.dir = dir
	if err != nil {
		return err
	}
	astFiles := []*ast.File{}
	for _, file := range files {
		fileName := file.Name()
		if len(fileName) <= 3 || fileName[len(fileName)-3:] != ".go" {
			continue
		}
		if 8 <= len(fileName) && fileName[len(fileName)-8:] == "_test.go" {
			continue
		}
		src, err := ioutil.ReadFile(filepath.Join(dir, fileName))
		if err != nil {
			return err
		}
		f := &analyzedFile{fileName: fileName}
		f.fileAST, err = parser.ParseFile(p.fset, fileName, src, parser.ParseComments)
		if err != nil {
			return err
		}
		p.files[fileName] = f
		astFiles = append(astFiles, f.fileAST)
	}
	errors := []error{}
	conf := types.Config{
		Importer: importer.Default(),
		Error:    func(err error) { errors = append(errors, err) },
	}
	pkg, err := conf.Check(p.pkgPath, p.fset, astFiles, p.info)
	p.pkg = pkg
	if 0 < len(errors) {
		for _, err := range errors {
			fmt.Println(err.Error())
		}
		panic("Error occured")
	}
	return nil
}

const fileHeader = `#include <string>
#include <memory>

using int8    = signed char;
using int16   = signed short;
using int32   = signed int;
using int64   = signed long long;
using uint8   = unsigned char;
using uint16  = unsigned short;
using uint32  = unsigned int;
using uint64  = unsigned long long;
using float32 = float;
using float64 = double;
using string  = std::string;

template <typename T>
inline std::shared_ptr<T> __ptr(T& t) {
    return &t;
}
`

// ToCPP は、このGoパッケージをC++コードに変換してファイルに保存します。
func (p *Package) ToCPP(dir, basePkg string) error {
	// fmt.Fprintln(writer, fileHeader)

	// ns := translateNamespace(relativePkg(a.pkgPath, basePkg))
	// segs := strings.Split(ns, "::")
	// for _, seg := range segs {
	// 	fmt.Fprintf(writer, "namespace %s {\n", seg)
	// }
	// fmt.Fprintln(writer, "")

	exports := []string{}
	warnings := []string{}
	fileNames := []string{}
	for fileName := range p.files {
		fileNames = append(fileNames, fileName)
	}
	sort.Strings(fileNames)
	for _, fileName := range fileNames {
		if !strings.HasSuffix(fileName, ".go") {
			continue
		}
		name := fileName[:len(fileName)-3]

		cppWriter, err := os.Create(filepath.FromSlash(dir + "/" + name + ".cpp"))
		if err != nil {
			return err
		}
		hWriter, err := os.Create(filepath.FromSlash(dir + "/" + name + ".h"))
		if err != nil {
			cppWriter.Close()
			return err
		}
		f := p.files[fileName]
		gen := newGenerator(cppWriter, hWriter, fileName, f.fileAST, basePkg, p.pkg, p.info, p.fset)
		gen.Dump()
		exports = append(exports, gen.Exports...)
		warnings = append(warnings, gen.Warnings...)
		hWriter.Close()
		cppWriter.Close()
	}

	// fmt.Fprintln(writer, "// ------------------------------------------------------------")
	// fmt.Fprintln(writer)
	// fmt.Fprintln(writer, "return {")
	// for i, exp := range exports {
	// 	fmt.Fprintf(writer, "\t%s = %s", exp, exp)
	// 	if i < len(exports)-1 {
	// 		fmt.Fprint(writer, ",")
	// 	}
	// 	fmt.Fprintln(writer)
	// }
	// fmt.Fprintln(writer, "}")

	// fmt.Fprintln(writer, "")
	// for range segs {
	// 	fmt.Fprintln(writer, "}")
	// }

	if 0 < len(warnings) {
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "// ------------------------------------------------------------ errors")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintf(os.Stderr, "%s", strings.Join(warnings, "\n"))
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr)
	}

	return nil
}
