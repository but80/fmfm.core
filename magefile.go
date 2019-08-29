// +build mage

package main

import (
	"os"
	"path/filepath"

	"github.com/magefile/mage/sh"
	"github.com/mattn/go-shellwords"
	"github.com/mattn/go-zglob"
)

func runVWithArgs(cmd string, args ...string) error {
	envArgs, err := shellwords.Parse(os.Getenv("ARGS"))
	if err != nil {
		return err
	}
	return sh.RunV(cmd, append(args, envArgs...)...)
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
	return sh.RunV("golangci-lint", "run")
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

// Build module version
func Buildmod() error {
	if err := os.MkdirAll(filepath.FromSlash("build/fmfm-module"), 0755); err != nil {
		return err
	}
	return sh.RunV(
		"go", "build",
		"-o", "build/fmfm-module/fmfm.so",
		"-buildmode=c-shared",
		"cmd/fmfm-module/main.go",
		"cmd/fmfm-module/helper.go",
	)
}

// Build WebAssembly version
func Buildwasm() error {
	if err := os.MkdirAll(filepath.FromSlash("build/fmfm-wasm"), 0755); err != nil {
		return err
	}
	return sh.RunWith(
		map[string]string{
			"GOOS":   "js",
			"GOARCH": "wasm",
		},
		"go", "build",
		"-o", "build/fmfm-wasm/fmfm.wasm",
		"cmd/fmfm-wasm/main.go",
	)
}
