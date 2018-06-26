package go2cpp

import (
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
)

func (g *generator) dumpBlock(stmt *ast.BlockStmt) {
	for _, s := range stmt.List {
		g.dumpStmt(s)
	}
}

func (g *generator) dumpStmt(stmt ast.Stmt) {
	switch s := stmt.(type) {

	case *ast.IfStmt:
		if s.Init != nil && !reflect.ValueOf(s.Init).IsNil() {
			g.dumpStmt(s.Init)
		}
		fmt.Fprint(g.cppWriter, g.indent)
		fmt.Fprint(g.cppWriter, "if (")
		g.dumpExpr(g.cppWriter, s.Cond)
		fmt.Fprintln(g.cppWriter, ") {")
		g.enter()
		g.dumpBlock(s.Body)
		g.leave()
		if s.Else != nil && !reflect.ValueOf(s.Else).IsNil() {
			fmt.Fprint(g.cppWriter, g.indent)
			fmt.Fprintln(g.cppWriter, "} else {")
			g.enter()
			g.dumpStmt(s.Else)
			g.leave()
		}
		fmt.Fprint(g.cppWriter, g.indent)
		fmt.Fprintln(g.cppWriter, "}")

	case *ast.BlockStmt:
		g.dumpBlock(s)

	case *ast.ReturnStmt:
		fmt.Fprint(g.cppWriter, g.indent)
		fmt.Fprint(g.cppWriter, "return")
		for i, r := range s.Results {
			if i != 0 {
				fmt.Fprint(g.cppWriter, ",")
			}
			fmt.Fprint(g.cppWriter, " ")
			g.dumpExpr(g.cppWriter, r)
		}
		fmt.Fprintln(g.cppWriter, ";")

	case *ast.BranchStmt:
		// @todo ラベル
		fmt.Fprint(g.cppWriter, g.indent)
		fmt.Fprint(g.cppWriter, s.Tok.String())
		if s.Label != nil {
			fmt.Fprintf(g.cppWriter, " %s", s.Label.Name)
		}
		fmt.Fprintln(g.cppWriter, ";")

	case *ast.AssignStmt:
		fmt.Fprint(g.cppWriter, g.indent)
		if s.Tok == token.DEFINE {
			fmt.Fprint(g.cppWriter, "auto ")
		}
		for i, e := range s.Lhs {
			if i != 0 {
				fmt.Fprint(g.cppWriter, ", ")
			}
			g.dumpExpr(g.cppWriter, e)
		}
		fmt.Fprintf(g.cppWriter, " = ")
		for i, e := range s.Rhs {
			if i != 0 {
				fmt.Fprint(g.cppWriter, ", ")
			}
			g.dumpExpr(g.cppWriter, e)
		}
		fmt.Fprintln(g.cppWriter, ";")

	case *ast.ExprStmt:
		fmt.Fprint(g.cppWriter, g.indent)
		g.dumpExpr(g.cppWriter, s.X)
		fmt.Fprintln(g.cppWriter, ";")

	case *ast.DeclStmt:
		g.dumpDecl(s.Decl)

	case *ast.SwitchStmt:
		// @todo 変数のスコープ
		if s.Init != nil && !reflect.ValueOf(s.Init).IsNil() {
			g.dumpStmt(s.Init)
		}
		hasTag := false
		if s.Tag != nil && !reflect.ValueOf(s.Tag).IsNil() {
			hasTag = true
			fmt.Fprint(g.cppWriter, g.indent)
			fmt.Fprint(g.cppWriter, "auto __tag = ")
			g.dumpExpr(g.cppWriter, s.Tag)
			fmt.Fprintln(g.cppWriter, ";")
		}
		for i, c := range s.Body.List {
			cc, _ := c.(*ast.CaseClause)
			fmt.Fprint(g.cppWriter, g.indent)
			if len(cc.List) == 0 {
				fmt.Fprintln(g.cppWriter, "} else {")
			} else {
				if i == 0 {
					fmt.Fprint(g.cppWriter, "if (")
				} else {
					fmt.Fprint(g.cppWriter, "} else if (")
				}
				for j, e := range cc.List {
					if j != 0 {
						fmt.Fprint(g.cppWriter, " || ")
					}
					if hasTag {
						fmt.Fprint(g.cppWriter, "__tag == ")
					}
					g.dumpExpr(g.cppWriter, e)
				}
				fmt.Fprintln(g.cppWriter, ") {")
			}
			g.enter()
			for _, s := range cc.Body {
				g.dumpStmt(s)
			}
			g.leave()
		}
		fmt.Fprint(g.cppWriter, g.indent)
		fmt.Fprintln(g.cppWriter, "}")

	case *ast.ForStmt:
		// @todo 変数のスコープ
		if s.Init != nil && !reflect.ValueOf(s.Init).IsNil() {
			g.dumpStmt(s.Init)
		}
		fmt.Fprint(g.cppWriter, g.indent)
		fmt.Fprint(g.cppWriter, "while (")
		if s.Cond != nil && !reflect.ValueOf(s.Cond).IsNil() {
			g.dumpExpr(g.cppWriter, s.Cond)
		} else {
			fmt.Fprint(g.cppWriter, "true")
		}
		fmt.Fprintln(g.cppWriter, ") {")
		g.enter()
		g.dumpBlock(s.Body)
		if s.Post != nil && !reflect.ValueOf(s.Post).IsNil() {
			g.dumpStmt(s.Post)
		}
		g.leave()
		fmt.Fprint(g.cppWriter, g.indent)
		fmt.Fprintln(g.cppWriter, "}")

	case *ast.RangeStmt:
		// @todo 変数のスコープ
		if g.isArrayType(s.X) {
			// 配列のrange
			fmt.Fprint(g.cppWriter, g.indent)
			fmt.Fprint(g.cppWriter, "for (int ")
			g.dumpExpr(g.cppWriter, s.Key)
			fmt.Fprint(g.cppWriter, " = 0; ")
			g.dumpExpr(g.cppWriter, s.Key)
			fmt.Fprint(g.cppWriter, " < sizeof(")
			g.dumpExpr(g.cppWriter, s.X)
			fmt.Fprint(g.cppWriter, ") / sizeof(")
			g.dumpExpr(g.cppWriter, s.X)
			fmt.Fprint(g.cppWriter, "[0]); ")
			g.dumpExpr(g.cppWriter, s.Key)
			fmt.Fprint(g.cppWriter, "++")
			fmt.Fprintln(g.cppWriter, ") {")
			g.enter()
			if s.Value != nil && !reflect.ValueOf(s.Value).IsNil() {
				fmt.Fprint(g.cppWriter, g.indent)
				fmt.Fprint(g.cppWriter, "auto ")
				g.dumpExpr(g.cppWriter, s.Value)
				fmt.Fprint(g.cppWriter, " = ")
				g.dumpExpr(g.cppWriter, s.X)
				fmt.Fprint(g.cppWriter, "[")
				g.dumpExpr(g.cppWriter, s.Key)
				fmt.Fprint(g.cppWriter, "]")
				fmt.Fprintln(g.cppWriter, ";")
			}
		} else if g.isSliceType(s.X) {
			// スライスのrange
			fmt.Fprint(g.cppWriter, g.indent)
			fmt.Fprint(g.cppWriter, "for (int ")
			g.dumpExpr(g.cppWriter, s.Key)
			fmt.Fprint(g.cppWriter, " = 0; ")
			g.dumpExpr(g.cppWriter, s.Key)
			fmt.Fprint(g.cppWriter, " < (int)")
			g.dumpExpr(g.cppWriter, s.X)
			fmt.Fprint(g.cppWriter, ".size(); ")
			g.dumpExpr(g.cppWriter, s.Key)
			fmt.Fprint(g.cppWriter, "++")
			fmt.Fprintln(g.cppWriter, ") {")
			g.enter()
			if s.Value != nil && !reflect.ValueOf(s.Value).IsNil() {
				fmt.Fprint(g.cppWriter, g.indent)
				fmt.Fprint(g.cppWriter, "auto ")
				g.dumpExpr(g.cppWriter, s.Value)
				fmt.Fprint(g.cppWriter, " = ")
				g.dumpExpr(g.cppWriter, s.X)
				fmt.Fprint(g.cppWriter, "[")
				g.dumpExpr(g.cppWriter, s.Key)
				fmt.Fprint(g.cppWriter, "]")
				fmt.Fprintln(g.cppWriter, ";")
			}
		} else if g.isMapType(s.X) {
			// mapのrange
			fmt.Fprint(g.cppWriter, g.indent)
			fmt.Fprint(g.cppWriter, "for (auto __p : ")
			g.dumpExpr(g.cppWriter, s.X)
			fmt.Fprintln(g.cppWriter, ") {")
			g.enter()
			if s.Key != nil && !reflect.ValueOf(s.Key).IsNil() {
				fmt.Fprint(g.cppWriter, g.indent)
				fmt.Fprint(g.cppWriter, "auto ")
				g.dumpExpr(g.cppWriter, s.Key)
				fmt.Fprintln(g.cppWriter, " = __p.first;")
			}
			if s.Value != nil && !reflect.ValueOf(s.Value).IsNil() {
				fmt.Fprint(g.cppWriter, g.indent)
				fmt.Fprint(g.cppWriter, "auto ")
				g.dumpExpr(g.cppWriter, s.Value)
				fmt.Fprintln(g.cppWriter, " = __p.second;")
			}
		} else {
			// ?
			fmt.Fprintln(g.cppWriter, "for (true) {")
			g.enter()
		}
		g.dumpBlock(s.Body)
		g.leave()
		fmt.Fprint(g.cppWriter, g.indent)
		fmt.Fprintln(g.cppWriter, "}")

	case *ast.IncDecStmt:
		fmt.Fprint(g.cppWriter, g.indent)
		g.dumpExpr(g.cppWriter, s.X)
		fmt.Fprint(g.cppWriter, s.Tok.String())
		fmt.Fprintln(g.cppWriter, ";")

	default:
		fmt.Fprint(g.cppWriter, g.indent)
		g.debugInspect(g.cppWriter, stmt, "stmt")
		fmt.Fprintln(g.cppWriter)
	}
}
