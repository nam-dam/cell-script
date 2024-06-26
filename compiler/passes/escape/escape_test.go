package escape

import (
	"testing"

	"github.com/cell-labs/cell-script/compiler/lexer"
	"github.com/cell-labs/cell-script/compiler/parser"
	"github.com/stretchr/testify/assert"
)

func escapeTest(t *testing.T, input string, expected map[string]bool) {
	lexed := lexer.Lex(input)
	parsed := parser.Parse(lexed, false)
	parsed = Escape(parsed)

	var allocsChecked []string

	for _, ins := range parsed.Instructions {
		if defFuncNode, ok := ins.(*parser.DefineFuncNode); ok {
			for _, ins := range defFuncNode.Body {
				if allocNode, ok := ins.(*parser.AllocNode); ok {
					allocsChecked = append(allocsChecked, allocNode.Name[0])
					assert.Equal(t, expected[allocNode.Name[0]], allocNode.Escapes, allocNode.Name)
				}
			}
		}
	}

	assert.Equal(t, len(allocsChecked), len(expected))
}

func TestNoEscape(t *testing.T) {
	escapeTest(t, `package main

	func main() {
		a := 100
		b := 200
	}
`, map[string]bool{
		"a": false,
		"b": false,
	})
}

func TestEscapes(t *testing.T) {
	escapeTest(t, `package main

		func main() {
			a := 100
			b := 200
			return b
		}
	`, map[string]bool{
		"a": false,
		"b": true,
	})
}

func TestEscapesPointer(t *testing.T) {
	escapeTest(t, `package main

		func main() *int {
			a := 100
			b := 200
			return &b
		}
	`, map[string]bool{
		"a": false,
		"b": true,
	})
}

func TestEscapesStructPointer(t *testing.T) {
	escapeTest(t, `package main

		type mytype struct {
			a int
			b int
		}

		func main() *int {
			a := 100
			b := mytype{
				a: 100,
				b: 200,
			}
			return &b
		}
	`, map[string]bool{
		"a": false,
		"b": true,
	})
}

func TestEscapeNestedStruct(t *testing.T) {
	escapeTest(t, `package main

		type Bar struct {
			num int64
		}

		type Foo struct {
			num int64
			bar *Bar
		}

		func GetFooPtr() *Foo {
			f := Foo{
				num: 300,
				bar: &Bar{num: 400},
			}

			return &f
		}`,
		map[string]bool{
			"f": true,
		})
}

/*
TODO: Implement feature so that this case can pass
f can be stack allocated, but f.bar needs to allocqated on the heap
func TestNoEscapeNestedStruct(t *testing.T) {
	escapeTest(t, `package main

		type Bar struct {
			num int64
		}

		type Foo struct {
			num int64
			bar *Bar
		}

		func GetFooPtr() Foo {
			f := Foo{
				num: 300,
				bar: &Bar{num: 400},
			}

			return f
		}`,
		map[string]bool{
			"f": false,
		})
}
*/
