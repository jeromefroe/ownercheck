package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"strings"
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

type visitor struct {
	cmap ast.CommentMap

	// map from function name to params in that function which take ownership of the
	// value passed to them
	ownedParams map[string][]int

	// set of nodes which have been passed to functions which take ownership of them
	ownedNodes map[*ast.Object]struct{}

	violations []string
}

func (v *visitor) Visit(node ast.Node) ast.Visitor {

	switch node := node.(type) {
	case *ast.FuncDecl:
		v.checkFunc(node)

	case *ast.CallExpr:
		v.checkCall(node)
	}

	return v
}

func (v *visitor) checkFunc(decl *ast.FuncDecl) {
	t := decl.Type
	if t == nil {
		return
	}

	ps := t.Params
	if ps == nil {
		return
	}

	for i, p := range ps.List {
		if p == nil {
			continue
		}

		if cs, ok := v.cmap[p]; ok {
			for _, c := range cs {
				if strings.TrimSpace(c.Text()) == "owned" {
					if len(p.Names) == 0 || p.Names[0] == nil {
						continue
					}

					owned, ok := v.ownedParams[decl.Name.Name]
					if !ok {
						owned = make([]int, 0)
					}
					owned = append(owned, i)
					v.ownedParams[decl.Name.Name] = owned
				}
			}
		}
	}
}

func (v *visitor) checkCall(expr *ast.CallExpr) {
	f := expr.Fun
	if f == nil {
		return
	}

	ident, ok := f.(*ast.Ident)
	if !ok {
		return
	}

	owned, ok := v.ownedParams[ident.Name]
	if !ok {
		return
	}

	for _, o := range owned {
		if o > len(expr.Args) {
			return
		}

		arg, ok := expr.Args[o].(*ast.Ident)
		if !ok {
			continue
		}

		if arg.Obj == nil {
			continue
		}

		if _, ok := v.ownedNodes[arg.Obj]; ok {
			msg := fmt.Sprintf("variable %s has already been passed to a function which owns it, cannot pass it again at pos %v", arg.Name, arg.Pos())
			v.violations = append(v.violations, msg)
		} else {
			v.ownedNodes[arg.Obj] = struct{}{}
		}
	}
}
