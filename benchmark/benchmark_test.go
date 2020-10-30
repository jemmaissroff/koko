package benchmark

import (
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"monkey/ast"
	"monkey/evaluator"
	"testing"
)

func testBuild(input string) (*ast.Program, *object.Environment) {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	return program, env
}

var result object.Object

func BenchmarkFib(b *testing.B) {
	var r object.Object
	program, env := testBuild(`let fib = fn(x) { if (x == 1) { 1 } else { if (x ==0) { 1} else { fib(x - 1) + fib(x - 2) }}};fib(5)`)
	for i := 0; i < b.N; i++ {
		r = evaluator.Eval(program, env)
	}
	result = r
}