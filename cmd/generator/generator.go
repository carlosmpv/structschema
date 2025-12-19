package main

import (
	"go/format"
	"go/token"
	"os"
	"strings"

	"codeberg.org/carlosmpv/schemagen/pkg/generator"
)

func main() {
	file, err := os.Open("schemaorg-current-https.jsonld")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	schema, err := generator.Parse(file)
	if err != nil {
		panic(err)
	}

	result := schema.GenFile("structschema")

	fset := token.NewFileSet()
	var output strings.Builder

	if err := format.Node(&output, fset, result); err != nil {
		panic(err)
	}

	os.WriteFile("structschema.go", []byte(output.String()), os.ModePerm)
}
