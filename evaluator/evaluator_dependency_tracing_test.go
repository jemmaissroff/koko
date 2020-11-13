package evaluator

import (
	"fmt"
	"math/rand"
	"monkey/object"
	"sort"
	"strconv"
	"testing"
)

func strEquals(t *testing.T, testStr string, truthStr string) {
	if testStr != truthStr {
		t.Errorf("string has wrong value. got %s, want %s",
			testStr, truthStr)
	}
}

/**
This section contains larger "integration tests".
**/
func TestMergeSortOnInts(t *testing.T) {
	rand.Seed(1)
	for i := 0; i < 20; i++ {
		arrLen := rand.Intn(100)
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
		mergeSortProgram := fmt.Sprintf(`
		let get_n_elements = pfn(arr, offset, number_of_elements) { if (number_of_elements == 0) { [] } else { [arr[offset]] + get_n_elements(arr, offset + 1, number_of_elements - 1) } }
		let merge_elements = pfn(res_lower, res_upper) { if (len(res_lower) == 0) { if (len(res_upper) == 0) { [] } else { res_upper } } else { if (len(res_upper) == 0) { res_lower } else { if (first(res_upper) < first(res_lower)) { [first(res_upper)] + merge_elements(res_lower, rest(res_upper)) } else { [first(res_lower)] + merge_elements(res_upper, rest(res_lower)) } } } }

		let merge_sort = pfn(arr) { if (len(arr) < 2) { return arr } else { let half = int(len(arr)/2); let res_lower = get_n_elements(arr, 0, half); let res_upper = get_n_elements(arr, half, len(arr) - half); merge_elements(merge_sort(res_lower), merge_sort(res_upper)) } }
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
