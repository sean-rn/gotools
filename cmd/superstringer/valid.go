package main

import (
	"log"
)

// GenerateValidMethod produces the Valid() method for the named type.
func (g *Generator) GenerateValidMethod(typeName string) {
	values := g.GetValues(typeName)
	if len(values) == 0 {
		log.Fatalf("no values defined for type %s", typeName)
	}
	runs := splitIntoRuns(values)
	switch {
	case len(runs) == 1:
		g.buildValidOneRun(runs, typeName)
	case len(runs) <= 10:
		g.buildValidMultipleRuns(runs, typeName)
	default:
		g.buildValidMap(runs, typeName)
	}
}

// buildOneRun generates the variables and String method for a single run of contiguous values.
func (g *Generator) buildValidOneRun(runs [][]Value, typeName string) {
	values := runs[0]
	g.Printf("\n")
	// The generated code is simple enough to write as a Printf format.
	lessThanZero := ""
	if values[0].signed {
		lessThanZero = "i >= 0 && "
	}
	if values[0].value == 0 { // Signed or unsigned, 0 is still 0.
		g.Printf(validOneRun, typeName, usize(len(values)), lessThanZero)
	} else {
		g.Printf(validOneRunWithOffset, typeName, values[0].String(), usize(len(values)), lessThanZero)
	}
}

// Arguments to format are:
//	[1]: type name
//	[2]: size of index element (8 for uint8 etc.)
//	[3]: less than zero check (for signed types)
const validOneRun = `func (i %[1]s) Valid() bool {
	return %[3]si < %[1]s(len(_%[1]s_index)-1)
}
`

// Arguments to format are:
//	[1]: type name
//	[2]: lowest defined value for type, as a string
//	[3]: size of index element (8 for uint8 etc.)
//	[4]: less than zero check (for signed types)
const validOneRunWithOffset = `func (i %[1]s) Valid() bool {
	i -= %[2]s
	return %[4]si < %[1]s(len(_%[1]s_index)-1)
}
`

// buildValidMultipleRuns generates the valid method for multiple runs of contiguous values.
// For this pattern, a single Printf format won't do.
func (g *Generator) buildValidMultipleRuns(runs [][]Value, typeName string) {
	g.Printf("\n")
	g.Printf("func (i %s) String() string {\n", typeName)
	g.Printf("\tswitch {\n")
	for _, values := range runs {
		if len(values) == 1 {
			g.Printf("\tcase i == %s:\n", &values[0])
			g.Printf("\t\treturn true\n")
			continue
		}
		if values[0].value == 0 && !values[0].signed {
			// For an unsigned lower bound of 0, "0 <= i" would be redundant.
			g.Printf("\tcase i <= %s:\n", &values[len(values)-1])
		} else {
			g.Printf("\tcase %s <= i && i <= %s:\n", &values[0], &values[len(values)-1])
		}
		g.Printf("\t\treturn true\n")
	}
	g.Printf("\tdefault:\n")
	g.Printf("\t\treturn false\n")
	g.Printf("\t}\n")
	g.Printf("}\n")
}

// buildValidMap handles the case where the space is so sparse a map is a reasonable fallback.
// It's a rare situation but has simple code.
func (g *Generator) buildValidMap(runs [][]Value, typeName string) {
	g.Printf("\n")
	g.Printf(validMap, typeName)
}

// Argument to format is the type name.
const validMap = `func (i %[1]s) String() string {
	if _, ok := _%[1]s_map[i]; ok {
		return true
	}
	return false
}
`
