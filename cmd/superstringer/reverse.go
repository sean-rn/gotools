package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"log"
	"sort"
)

// Arguments to format are:
//	[1]: type name
const fromStringTemplate = `
// %[1]sFromString looks up the corresponding %[1]s from its String() value.
func %[1]sFromString(s string) (%[1]s, bool) {
	i := sort.Search(len(_%[1]s_sorted), func(i int) bool {
		return _%[1]s_sorted[i].String() >= s
	})
	if i < len(_%[1]s_sorted) && _%[1]s_sorted[i].String() == s {
		return _%[1]s_sorted[i], true
	}
	return 0, false
}
`

// GenerateFromString produces the FromString function for the named type.
func (g *Generator) GenerateFromString(typeName string) {
	values := g.GetValues(typeName)
	if len(values) == 0 {
		log.Fatalf("no values defined for type %s", typeName)
	}

	// Sort the values by their string value
	sort.SliceStable(values, func(i, j int) bool {
		return values[i].name < values[j].name
	})

	g.Printf("var %s\n", createSortedDecl(typeName, values))
	g.Printf(fromStringTemplate, typeName)
}

// createSortedDecl returns the declaration sorted by string value. The caller will add "var"
func createSortedDecl(typeName string, values []Value) string {
	b := new(bytes.Buffer)
	fmt.Fprintf(b, "_%s_sorted = [...]%s{", typeName, typeName)
	for i := range values {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(values[i].originalName)
	}
	b.WriteRune('}')
	return b.String()
}

// GetValues finds all the values for the type typeName
func (g *Generator) GetValues(typeName string) []Value {
	values := make([]Value, 0, 100)
	for _, file := range g.pkg.files {
		// Set the state for this run of the walker.
		file.typeName = typeName
		file.values = nil
		if file.file != nil {
			ast.Inspect(file.file, file.genDecl)
			values = append(values, file.values...)
		}
	}
	return values
}
