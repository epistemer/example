// Command stub is the workaround of findings F4: it rewrites the Go files
// of a handler package in place so the package type-checks without the
// routes package it generates. Every function body becomes panic(0),
// package-level values that use that package (transitively) are dropped,
// and so are imports nothing uses any more. Run it on a copy, or commit
// first and restore after:
//
//	go run ./scripts/stub ./internal/admin adminroutes && go tool genx && git checkout -- internal/admin
package main

import (
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var tainted = map[string]bool{}

func main() {
	tainted[os.Args[2]] = true
	for i := 0; i < 5; i++ {
		run()
	}
}

func run() {
	dir := os.Args[1]
	files, _ := filepath.Glob(filepath.Join(dir, "*.go"))
	for _, p := range files {
		if strings.HasSuffix(p, "_test.go") {
			continue
		}
		fs := token.NewFileSet()
		f, err := parser.ParseFile(fs, p, nil, parser.ParseComments)
		if err != nil {
			panic(err)
		}
		stub := &ast.BlockStmt{List: []ast.Stmt{&ast.ExprStmt{X: &ast.CallExpr{Fun: ast.NewIdent("panic"), Args: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}}}}}}
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.FuncDecl:
				if x.Body != nil {
					x.Body = stub
				}
			case *ast.FuncLit:
				x.Body = stub
			}
			return true
		})
		// Package-level vars whose values use the generated package: drop
		// the value, keep a typed zero when a type is written.
		for _, d := range f.Decls {
			g, ok := d.(*ast.GenDecl)
			if !ok || g.Tok != token.VAR {
				continue
			}
			for _, s := range g.Specs {
				vs := s.(*ast.ValueSpec)
				uses := false
				for _, v := range vs.Values {
					ast.Inspect(v, func(n ast.Node) bool {
						if id, ok := n.(*ast.Ident); ok && tainted[id.Name] {
							uses = true
						}
						return true
					})
				}
				if uses {
					for _, nm := range vs.Names {
						tainted[nm.Name] = true
					}
					for i := range vs.Values {
						vs.Values[i] = &ast.CallExpr{Fun: &ast.FuncLit{Type: &ast.FuncType{Params: &ast.FieldList{}, Results: &ast.FieldList{List: []*ast.Field{{Type: ast.NewIdent("any")}}}}, Body: stub}}
					}
					vs.Type = nil
				}
			}
		}
		used := map[string]bool{}
		ast.Inspect(f, func(n ast.Node) bool {
			if se, ok := n.(*ast.SelectorExpr); ok {
				if id, ok := se.X.(*ast.Ident); ok {
					used[id.Name] = true
				}
			}
			return true
		})
		var keep []ast.Spec
		for _, d := range f.Decls {
			g, ok := d.(*ast.GenDecl)
			if !ok || g.Tok != token.IMPORT {
				continue
			}
			keep = keep[:0]
			for _, s := range g.Specs {
				is := s.(*ast.ImportSpec)
				path, _ := strconv.Unquote(is.Path.Value)
				name := path[strings.LastIndex(path, "/")+1:]
				if is.Name != nil {
					name = is.Name.Name
				}
				if name == "_" || name == "." || used[name] || strings.HasPrefix(name, "v") && used[strings.Split(path, "/")[len(strings.Split(path, "/"))-2]] {
					keep = append(keep, s)
				}
			}
			g.Specs = append([]ast.Spec(nil), keep...)
		}
		out, err := os.Create(p)
		if err != nil {
			panic(err)
		}
		format.Node(out, fs, f)
		out.Close()
	}
}
