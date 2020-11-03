package benchmark

import (
	"monkey/ast"
	"monkey/evaluator"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"testing"
)

func testBuild(input string) (*ast.Program, *object.Environment) {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	return program, env
}

func BenchmarkFib(b *testing.B) {
	program, env := testBuild(`let fib = fn(x) { if (x == 1) { 1 } else { if (x ==0) { 1} else { fib(x - 1) + fib(x - 2) }}};fib(8)`)
	for i := 0; i < b.N; i++ {
		evaluator.Eval(program, env)
	}
}

func BenchmarkPureFib(b *testing.B) {
	program, env := testBuild(`let fib = pfn(x) { if (x == 1) { 1 } else { if (x ==0) { 1} else { fib(x - 1) + fib(x - 2) }}};fib(8)`)
	for i := 0; i < b.N; i++ {
		evaluator.Eval(program, env)
	}
}

func BenchmarkCollatz(b *testing.B) {
	program, env := testBuild(`
	let collatz = fn(n) { if (n==1) { 0 } else { if (n%2 == 0) { collatz(n/2) + 1 } else { collatz(3*n + 1) + 1 }}};
	let compute_sum_of_first_n_collatz = fn(n) { if (n== 1) { 0 } else { collatz(n) + compute_sum_of_first_n_collatz(n - 1) }};
	compute_sum_of_first_n_collatz(100)`)
	for i := 0; i < b.N; i++ {
		evaluator.Eval(program, env)
	}
}

func BenchmarkPureCollatz(b *testing.B) {
	program, env := testBuild(`let collatz = pfn(n) { if (n==1) { 0 } else { if (n%2 == 0) { collatz(n/2) + 1 } else { collatz(3*n + 1) + 1 }}};
	let compute_sum_of_first_n_collatz = fn(n) { if (n== 1) { 0 } else { collatz(n) + compute_sum_of_first_n_collatz(n - 1) }};
	compute_sum_of_first_n_collatz(100)`)
	for i := 0; i < b.N; i++ {
		evaluator.Eval(program, env)
	}
}
