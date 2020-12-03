package evaluator

import (
	"fmt"
	"koko/lexer"
	"koko/object"
	"koko/parser"
	"math/rand"
	"sort"
	"strconv"
	"testing"
)

func testEvalAndGetDeps(input string, fname string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()
	return Eval(program, env)
}

func strEquals(t *testing.T, testStr string, truthStr string) {
	if testStr != truthStr {
		t.Errorf("string has wrong value. got %s, want %s",
			testStr, truthStr)
	}
}

/*func assertProgramDidNotError(t *testing.T, res object.Object) object.Object {
	if ok, _ := res.(object.Error); ok {

	}
}*/

func assertObjectDepsEqual(t *testing.T, res object.Object, expectedDeps []string) {
	debugTraceObj, ok := res.(*object.DebugTraceMetadata)
	if !ok {
		t.Errorf("Expected Debug Trace Object got %+v\n", res)
		return
	}
	deps := debugTraceObj.GetDebugMetadata().Dependencies
	expectedDepsMap := make(map[string]bool)
	for _, d := range expectedDeps {
		expectedDepsMap[d] = true
		ok, val := deps[d]
		if !(ok && val) {
			t.Errorf("Expected Dependency %s missing on object %+v\n", d, res)
			return
		}
	}
	// check that we have no other deps
	for k, v := range deps {
		if v {
			if ok, _ := expectedDepsMap[k]; !ok {
				t.Errorf("Extra Dependency %s on object %+v\n", k, res)
				return
			}
		}
	}
}

func TestDependencyTrackingInBasicFunctionWithIntegers(t *testing.T) {
	program := "let f = pfn(a, b) { b }; deps(f, 1, 2)"
	res := testEval(program)
	assertObjectDepsEqual(t, res, []string{"1"})
}

func TestDependencyTrackingInBasicFunctionWithIntegerAddition(t *testing.T) {
	program := "let f = pfn(a, b, c) { a + c }; deps(f, 1, 2, 3)"
	res := testEval(program)
	assertObjectDepsEqual(t, res, []string{"0", "2"})
}

func TestDependencyTrackingInBasicFunctionWithIntegerMultiplication(t *testing.T) {
	program := "let f = pfn(a, b, c) { a * b * c }; deps(f, 1, 2, 0)"
	res := testEval(program)
	assertObjectDepsEqual(t, res, []string{"2"})
}

func TestDependencyTrackingInBasicFunctionWithConditional(t *testing.T) {
	program := "let f = pfn(a, b, c) { if (a > 0) { b } else { c } }; deps(f, 1, 2, 0)"
	res := testEval(program)
	assertObjectDepsEqual(t, res, []string{"0", "1"})
	program = "let f = pfn(a, b, c) { if (a > 0) { b } else { c } }; deps(f, -1, 2, 0)"
	res = testEval(program)
	assertObjectDepsEqual(t, res, []string{"0", "2"})
}

func TestDependencyTrackingInSubFunctions(t *testing.T) {
	program := "let g = pfn(a, b) { b }; let f = pfn(a, b, c) { g(c, a) }; deps(f, 1, 2, 3)"
	res := testEval(program)
	assertObjectDepsEqual(t, res, []string{"0"})
}

func TestDependencyTrackingInBasicFunctionWithArrays(t *testing.T) {
	program := "let f = pfn(a) { a[2] + a[3] }; deps(f, [1,2,3,4,5])"
	res := testEval(program)
	assertObjectDepsEqual(t, res, []string{"0|2", "0|3"})
}

func TestDependencyTrackingInFunctionWithArraysWhichReturnAnArray(t *testing.T) {
	program := "let f = pfn(a) { [a[0],a[1],a[2],a[3]] }; deps(f, [1,2,3,4,5])"
	res := testEval(program)
	assertObjectDepsEqual(t, res, []string{"0|0", "0|1", "0|2", "0|3"})
}

func TestDependencyTrackingInSubFunctionsWithArrays(t *testing.T) {
	program := "let f = pfn(a) { [a[0],a[1],a[2],a[3]] }; let g = pfn(a) { f(a)[0] + f(a)[2] }; deps(g, [1,2,3,4,5])"
	res := testEval(program)
	assertObjectDepsEqual(t, res, []string{"0|0", "0|2"})
}

func TestDependencyTrackingInSubFunctionsWithArrayConcatenation(t *testing.T) {
	program := "let f = pfn(a, b) { (b + a)[2] }; deps(f, [1,2], [3, 4])"
	res := testEval(program)
	assertObjectDepsEqual(t, res, []string{"0|0", "1#"})
	program = "let f = pfn(a, b) { (b + a)[1] }; deps(f, [1,2], [3, 4])"
	res = testEval(program)
	assertObjectDepsEqual(t, res, []string{"1|1"})
}

/*
* TESTING BUILTINS HERE
 */
func TestStringToArrayConversion(t *testing.T) {
	program := "let f = pfn(s) { array(s)[2] }; deps(f, \"hello word\")"
	res := testEval(program)
	assertObjectDepsEqual(t, res, []string{"0"})
	program = "let f = pfn(s) { len(array(s)) }; deps(f, \"hello word\")"
	res = testEval(program)
	assertObjectDepsEqual(t, res, []string{"0"})
}

func TestStringToArrayConversionForOtherType(t *testing.T) {
	program := "let f = pfn(s) { len(array(s)) }; deps(f, 1)"
	res := testEval(program)
	assertObjectDepsEqual(t, res, []string{"0"})
	program = "let f = pfn(s) { len(array(s)) }; deps(f, 1)"
	res = testEval(program)
	assertObjectDepsEqual(t, res, []string{"0"})
}

/**
This section contains larger "integration tests".
**/
func TestMergeSortOnInts(t *testing.T) {
	rand.Seed(1)
	for i := 0; i < 1; i++ {
		arrLen := 2
		arr := make([]int, arrLen)
		for j := 0; j < arrLen; j++ {
			arr[j] = rand.Intn(200000) - 100000
		}

		expectedResult := make([]int, len(arr))
		copy(expectedResult, arr)
		sort.Ints(expectedResult)

		strInput := "["
		for j := 0; j < arrLen; j++ {
			strInput += strconv.Itoa(arr[j])
			if j < arrLen-1 {
				strInput += ","
			}
		}
		strInput += "]"
		fmt.Printf("in: %s", strInput)
		mergeSortProgram := fmt.Sprintf(`
		let get_n_elements = fn(arr, offset, number_of_elements) { if (number_of_elements == 0) { [] } else { [arr[offset]] + get_n_elements(arr, offset + 1, number_of_elements - 1) } }
		let merge_elements = fn(res_lower, res_upper) { if (len(res_lower) == 0) { if (len(res_upper) == 0) { [] } else { res_upper } } else { if (len(res_upper) == 0) { res_lower } else { if (first(res_upper) < first(res_lower)) { [first(res_upper)] + merge_elements(res_lower, rest(res_upper)) } else { [first(res_lower)] + merge_elements(res_upper, rest(res_lower)) } } } }

		let merge_sort = fn(arr) { if (len(arr) < 2) { return arr } else { let half = int(len(arr)/2); let res_lower = get_n_elements(arr, 0, half); let res_upper = get_n_elements(arr, half, len(arr) - half); merge_elements(merge_sort(res_lower), merge_sort(res_upper)) } }
		merge_sort(%s)
		`, strInput)

		outList := testEval(mergeSortProgram)
		outStr := "["
		expectedStr := "["
		for j := 0; j < arrLen; j++ {
			elem := outList.(*object.Array).Elements[j]
			outStr += elem.Inspect()
			expectedStr += strconv.Itoa(expectedResult[j])
			if j < arrLen-1 {
				outStr += ","
				expectedStr += ","
			}
		}
		outStr += "]"
		expectedStr += "]"
		strEquals(t, outStr, expectedStr)
	}
}
