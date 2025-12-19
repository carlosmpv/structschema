package generator_test

import (
	"encoding/json"
	"fmt"
	"go/printer"
	"go/token"
	"os"
	"strings"
	"testing"

	"codeberg.org/carlosmpv/schemagen/pkg/generator"
	"github.com/stretchr/testify/assert"
)

func TestThingTotalExpansion(t *testing.T) {
	file, err := os.Open("schemaorg-current-https.jsonld")
	assert.NoError(t, err)
	defer file.Close()

	schema, err := generator.Parse(file)
	assert.NoError(t, err)

	out := schema.GenFile("schema")

	fset := token.NewFileSet()
	var output strings.Builder
	if err := printer.Fprint(&output, fset, out); err != nil {
		panic(err)
	}
	fmt.Println(output.String())

	msg, err := json.MarshalIndent(schema, "", "\t")
	assert.NoError(t, err)
	os.WriteFile("out.json", msg, os.ModePerm)
}
