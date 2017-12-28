package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
)

func main() {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "testdata/example.go", nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	cmap := ast.NewCommentMap(fset, f, f.Comments)
	v := &visitor{
		cmap:        cmap,
		ownedParams: make(map[string][]int),
		ownedNodes:  make(map[*ast.Object]struct{}),
	}
	ast.Walk(v, f)

	for _, violation := range v.violations {
		fmt.Println(violation)
	}
}
