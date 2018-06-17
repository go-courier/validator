package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strings"

	"github.com/go-courier/codegen"
)

func main() {
	fset := token.NewFileSet()
	f, _ := parser.ParseFile(fset, "strfmt.go", nil, parser.ParseComments)

	file := codegen.NewFile("strfmt", "strfmt_validators.go")

	constNames := make([]string, 0)
	for key, obj := range f.Scope.Objects {
		if obj.Kind == ast.Con {
			constNames = append(constNames, key)
		}
	}
	sort.Strings(constNames)

	for _, constName := range constNames {

		name := bytes.Replace([]byte(constName), []byte("regexpString"), []byte(""), 1)
		validatorName := strings.Replace(codegen.LowerSnakeCase(string(name)), "_", "-", -1)
		validatorAlias := codegen.LowerCamelCase(string(name))

		args := []codegen.Snippet{
			codegen.Id(constName),
			codegen.Val(validatorName),
		}

		if validatorName != validatorAlias {
			args = append(args, codegen.Val(validatorAlias))
		}

		file.WriteBlock(
			codegen.Expr("var ? = ?",
				codegen.Id(codegen.UpperCamelCase(string(name))+"Validator"),
				codegen.Call(
					file.Use("github.com/go-courier/validator", "NewRegexpStrfmtValidator"),
					args...,
				),
			),
		)
	}

	file.WriteFile()
}
