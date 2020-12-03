package benchmark

import (
	"koko/ast"
	"koko/evaluator"
	"koko/lexer"
	"koko/object"
	"koko/parser"
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

func BenchmarkMergeSort(b *testing.B) {
	program, env := testBuild(`
	let get_n_elements = fn(arr, offset, number_of_elements) { if (number_of_elements == 0) { [] } else { [arr[offset]] + get_n_elements(arr, offset + 1, number_of_elements - 1) } }
	let merge_elements = fn(res_lower, res_upper) { if (len(res_lower) == 0) { if (len(res_upper) == 0) { [] } else { res_upper } } else { if (len(res_upper) == 0) { res_lower } else { if (first(res_upper) < first(res_lower)) { [first(res_upper)] + merge_elements(res_lower, rest(res_upper)) } else { [first(res_lower)] + merge_elements(res_upper, rest(res_lower)) } } } }

	let merge_sort = fn(arr) { if (len(arr) == 1) { return arr } else { let half = int(len(arr)/2); let res_lower = get_n_elements(arr, 0, half); let res_upper = get_n_elements(arr, half, len(arr) - half); merge_elements(merge_sort(res_lower), merge_sort(res_upper)) } }

	let RAND_CONST = 10000
	let random_array = fn(len) { if (len == 0) { [] } else { [rando(RAND_CONST)] + random_array(len - 1) } }
    let ra = random_array(200)

	let repeat_merge_sort_with_modifications = fn(repeats, arr) {
		if (repeats != 0) {
		   let mod_ind = rando(len(ra))
		   merge_sort(arr)
		   repeat_merge_sort_with_modifications(repeats - 1, get_n_elements(ra, 0, mod_ind) + [rando(RAND_CONST)] + get_n_elements(ra, mod_ind, len(ra) - mod_ind))
	   }
   }
   repeat_merge_sort_with_modifications(200, ra)
	`)
	for i := 0; i < b.N; i++ {
		evaluator.Eval(program, env)
	}
}

func BenchmarkPureMergeSort(b *testing.B) {
	// merge_sort([-10,8,100,2,1])
	// merge_sort([10,4,-1000,7,-68,99,1])
	// merge_sort([2,-8])
	// merge_sort([-2,1,-11])
	program, env := testBuild(`
	let get_n_elements = pfn(arr, offset, number_of_elements) { if (number_of_elements == 0) { [] } else { [arr[offset]] + get_n_elements(arr, offset + 1, number_of_elements - 1) } }
	let merge_elements = pfn(res_lower, res_upper) { if (len(res_lower) == 0) { if (len(res_upper) == 0) { [] } else { res_upper } } else { if (len(res_upper) == 0) { res_lower } else { if (first(res_upper) < first(res_lower)) { [first(res_upper)] + merge_elements(res_lower, rest(res_upper)) } else { [first(res_lower)] + merge_elements(res_upper, rest(res_lower)) } } } }

	let merge_sort = pfn(arr) { if (len(arr) < 2) { return arr } else { let half = int(len(arr)/2); let res_lower = get_n_elements(arr, 0, half); let res_upper = get_n_elements(arr, half, len(arr) - half); merge_elements(merge_sort(res_lower), merge_sort(res_upper)) } }

	let RAND_CONST = 10000
	let random_array = fn(len) { if (len == 0) { [] } else { [rando(RAND_CONST)] + random_array(len - 1) } }
    let ra = random_array(100)

	let repeat_merge_sort_with_modifications = fn(repeats, arr) {
		 if (repeats != 0) {
			let mod_ind = rando(len(arr))
			merge_sort(arr)
			repeat_merge_sort_with_modifications(repeats - 1, get_n_elements(arr, 0, mod_ind) + [rando(RAND_CONST)] + get_n_elements(arr, mod_ind, len(arr) - mod_ind))
		}
	}
	repeat_merge_sort_with_modifications(10, ra)
	`)
	for i := 0; i < b.N; i++ {
		evaluator.Eval(program, env)
	}
}

func BenchmarkRockHopper(b *testing.B) {
	program, env := testBuild(`
	let RAND_CONST = 10
	let random_array = fn(len) { if (len == 0) { [] } else { [rando(RAND_CONST)] + random_array(len - 1) } }
	let ra = random_array(1000)
	
	let rock_hopper = fn(arr, pos) { if (pos > len(arr) - 1) { true } else { if (arr[pos] == 0) { false } else { rock_hopper(arr, pos + arr[pos]) }}}
	let repeat_rock_hopper_with_modifications = fn(repeats, arr) { if (repeats != 0) { let mod_ind = rando(len(arr)); rock_hopper(arr, 0); repeat_rock_hopper_with_modifications(repeats - 1, get_n_elements(arr, 0, mod_ind) + [rando(RAND_CONST)] + get_n_elements(arr, mod_ind, len(arr) - mod_ind))}}
	repeat_rock_hopper_with_modifications(10, ra)
	`)
	for i := 0; i < b.N; i++ {
		evaluator.Eval(program, env)
	}
}

func BenchmarkPureRockHopper(b *testing.B) {
	program, env := testBuild(`
	let RAND_CONST = 10
	let random_array = fn(len) { if (len == 0) { [] } else { [rando(RAND_CONST)] + random_array(len - 1) } }
	let ra = random_array(1000)
	let get_n_elements = fn(arr, offset, number_of_elements) { if (number_of_elements == 0) { [] } else { [arr[offset]] + get_n_elements(arr, offset + 1, number_of_elements - 1) } }

	
	let rock_hopper = pfn(arr, pos) { if (pos > len(arr) - 1) { true } else { if (arr[pos] == 0) { false } else { rock_hopper(arr, pos + arr[pos]) }}}
	let repeat_rock_hopper_with_modifications = fn(repeats, arr) { if (repeats != 0) { let mod_ind = rando(len(arr)); rock_hopper(arr, 0); repeat_rock_hopper_with_modifications(repeats - 1, get_n_elements(arr, 0, mod_ind) + [rando(RAND_CONST)] + get_n_elements(arr, mod_ind, len(arr) - mod_ind))}}
	repeat_rock_hopper_with_modifications(10, ra)
	`)
	for i := 0; i < b.N; i++ {
		evaluator.Eval(program, env)
	}
}

func BenchmarkPureFibPartialCache(b *testing.B) {
	program, env := testBuild(`let fib = pfn(x, b) { if (x==1005) { b } else { if (x == 1) { 1 } else { if (x == 0) { 1 } else { fib(x - 1, x * b + 1) + fib(x - 2, x * b + 1) }}}};fib(8)`)
	for i := 0; i < b.N; i++ {
		evaluator.Eval(program, env)
	}
}
