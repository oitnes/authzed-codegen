package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/danhtran94/authzed-codegen/internal/ast"
	"github.com/danhtran94/authzed-codegen/internal/generator"
	"github.com/danhtran94/authzed-codegen/internal/templates"
)

var outputPath string

func init() {
	flag.StringVar(&outputPath, "output", "zed", "output path for generated files")

}

func main() {
	if len(os.Args) < 2 {
		panic(fmt.Errorf("missing schema path"))
	}

	flag.Parse()

	schemePath := os.Args[len(os.Args)-1]

	input, err := os.ReadFile(schemePath)
	if err != nil {
		panic(err)
	}

	lex := ast.NewLexer(string(input))
	parser := ast.NewParser(lex.Lex())
	ast, err := parser.ParseDefinitions()
	if err != nil {
		panic(err)
	}

	g := generator.NewGenerator(ast)
	g.OutputPath = outputPath
	g.AddObjectTemplate("[object]", string(templates.ObjectTemplate))

	err = g.GenerateObjectSource("[object]")
	if err != nil {
		panic(err)
	}
}
