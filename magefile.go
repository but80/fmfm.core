// +build mage

package main

import (
	"os"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/mattn/go-shellwords"
	"github.com/mattn/go-zglob"
	"gopkg.in/but80/fmfm.core.v1/go2cpp"
)

func runVWithArgs(cmd string, args ...string) error {
	envArgs, err := shellwords.Parse(os.Getenv("ARGS"))
	if err != nil {
		return err
	}
	return sh.RunV(cmd, append(args, envArgs...)...)
}

func gen(basePkg, pkg string) error {
	fullpkg := basePkg
	if pkg != "" {
		fullpkg += "/" + pkg
	}

	p := go2cpp.NewPackage(fullpkg)
	if err := p.Load(); err != nil {
		return err
	}

	dir := "cpp"
	if pkg != "" {
		dir += "/" + pkg
	}
	if err := os.MkdirAll(filepath.FromSlash(dir), 0755); err != nil {
		return err
	}
	return p.ToCPP(dir, basePkg)
}

// Generates C++ code from Go code
func Gen() error {
	if err := gen("gopkg.in/but80/fmfm.core.v1", "sim"); err != nil {
		return err
	}
	if err := gen("gopkg.in/but80/fmfm.core.v1", "ymf/ymfdata"); err != nil {
		return err
	}
	return nil
}

func BuildCPP() error {
	mg.SerialDeps(Gen)
	err := os.Chdir("cpp")
	if err != nil {
		return err
	}
	defer os.Chdir("..")
	files, err := zglob.Glob("./**/*.cpp")
	if err != nil {
		return err
	}
	opts := append([]string{"-dynamiclib", "-o", "fmfm.dylib"}, files...)
	return sh.RunV("g++", opts...)
}

// Format code
func Fmt() error {
	files, err := zglob.Glob("./**/*.go")
	if err != nil {
		return err
	}
	for _, file := range files {
		if err := sh.RunV("goimports", "-w", file); err != nil {
			return err
		}
	}
	return nil
}

// Check coding style
func Lint() error {
	return sh.RunV("gometalinter", "--config=.gometalinter.json", "./...")
}

// Run test
func Test() error {
	return sh.RunV("go", "test", "./...")
}

// Run program
func Run() error {
	return runVWithArgs("go", "run", "cmd/fmfm-cli/main.go")
}

// Run program with profiling
func Prof() error {
	return runVWithArgs("go", "run", "cmd/fmfm-cli/*.go")
}

// Run program with profiling without inlining optimization
func Prof2() error {
	return runVWithArgs("go", "run", "-gcflags", "-N -l", "cmd/fmfm-cli/*.go")
}

// Build binary
func Build() error {
	if err := os.MkdirAll(filepath.FromSlash("build/fmfm-module"), 0755); err != nil {
		return err
	}
	return sh.RunV("go", "build", "-o", "build/fmfm-module/fmfm.so", "-buildmode=c-shared", "cmd/fmfm-module/main.go")
}
