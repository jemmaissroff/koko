package evaluator

import (
	"monkey/ast"
	"monkey/object"
	"strconv"

	"fmt"
	"math"
)

var (
	NIL = &object.Nil{}

	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}

	EMPTY_STRING = &object.String{Value: ""}
	ZERO_INTEGER = &object.Integer{Value: 0}
	ZERO_FLOAT   = &object.Float{Value: 0}
)

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {

	case *ast.Program:
		return evalProgram(node, env)

	case *ast.BlockStatement:
		return evalBlockStatement(node, env)

	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)

	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		res := &object.Return{Value: val}
		// TODP (Peter) double check these references...
		res.SetMetadata(val.GetMetadata())
		return res

	case *ast.IntegerLiteral:
		switch node.Value {
		case 0:
			return ZERO_INTEGER
		default:
			return &object.Integer{Value: node.Value}
		}
	case *ast.FloatLiteral:
		switch node.Value {
		case 0:
			return ZERO_FLOAT
		default:
			return &object.Float{Value: node.Value}
		}
	case *ast.StringLiteral:
		switch node.Value {
		case "":
			return EMPTY_STRING
		default:
			return &object.String{Value: node.Value}

		}
	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		return env.Set(node.Name.Value, val)
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Env: env, Body: body}
		// JEM: You left off here. Need to implement a pure function literal
	case *ast.PureFunctionLiteral:
		params := node.Parameters
		body := node.Body
		return object.NewPureFunction(params, env, body)
	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}
		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return applyFunction(function, args)
	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}
	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		return evalIndexExpression(left, index)
	}
	return nil
}

func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range program.Statements {
		result = Eval(statement, env)

		switch result := result.(type) {
		case *object.Return:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range block.Statements {
		result = Eval(statement, env)

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}
	return result
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ:
		switch {
		case right.Type() == object.ARRAY_OBJ:
			return evalArrayInfixExpression(operator, left, right)
		case operator == "==":
			return FALSE
		case operator == "!=":
			return TRUE
		}
	case left.Type() == object.STRING_OBJ:
		switch {
		case right.Type() == object.STRING_OBJ:
			return evalStringInfixExpression(operator, left, right)
		case operator == "*" && right.Type() == object.INTEGER_OBJ:
			return multiplyStrings(left, right)
		case operator == "+":
			return addStrings(left, right)
		case operator == "==":
			return FALSE
		case operator == "!=":
			return TRUE
		}
	case right.Type() == object.STRING_OBJ:
		switch {
		case operator == "*" && left.Type() == object.INTEGER_OBJ:
			return multiplyStrings(right, left)
		case operator == "+":
			return addStrings(left, right)
		case operator == "==":
			return FALSE
		case operator == "!=":
			return TRUE
		}
	case left.Type() == object.INTEGER_OBJ:
		switch {
		case right.Type() == object.INTEGER_OBJ:
			return evalIntegerInfixExpression(operator, left, right)
		case right.Type() == object.FLOAT_OBJ:
			return evalFloatInfixExpression(operator, intToFloat(left), right)

		}
	case left.Type() == object.FLOAT_OBJ:
		switch {
		case right.Type() == object.INTEGER_OBJ:
			return evalFloatInfixExpression(operator, left, intToFloat(right))
		case right.Type() == object.FLOAT_OBJ:
			return evalFloatInfixExpression(operator, left, right)
		}
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
	return newError("type mismatch: %s %s %s",
		left.Type(), operator, right.Type())
}

func evalIntegerInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	lVal := left.(*object.Integer).Value
	rVal := right.(*object.Integer).Value
	var res object.Object

	switch operator {
	case "+":
		res = &object.Integer{Value: lVal + rVal}
	case "-":
		res = &object.Integer{Value: lVal - rVal}
	case "*":
		res = &object.Integer{Value: lVal * rVal}
		// Extra trick: If one number is actually zero we only need to depend on it!
		if lVal == 0 {
			res.SetMetadata(left.GetMetadata())
			return res
		} else if rVal == 0 {
			res.SetMetadata(right.GetMetadata())
			return res
		}
	case "/":
		res = evalFloatInfixExpression(operator, intToFloat(left), intToFloat(right))
	case "<":
		res = nativeBoolToBooleanObject(lVal < rVal)
	case ">":
		res = nativeBoolToBooleanObject(lVal > rVal)
	case "==":
		res = nativeBoolToBooleanObject(lVal == rVal)
	case "!=":
		res = nativeBoolToBooleanObject(lVal != rVal)
	case "%":
		res = &object.Integer{Value: lVal % rVal}
	default:
		return newError("unknown operator for INTEGER %v", operator)
	}
	res.SetMetadata(object.MergeDependencies(left.GetMetadata(), right.GetMetadata()))
	return res
}

func evalFloatInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	lVal := left.(*object.Float).Value
	rVal := right.(*object.Float).Value

	var res object.Object
	switch operator {
	case "+":
		res = &object.Float{Value: lVal + rVal}
	case "-":
		res = &object.Float{Value: lVal - rVal}
	case "*":
		res = &object.Float{Value: lVal * rVal}
		// Extra trick: If one number is actually zero we only need to depend on it!
		if lVal == 0 {
			res.SetMetadata(left.GetMetadata())
			return res
		} else if rVal == 0 {
			res.SetMetadata(right.GetMetadata())
			return res
		}
	case "/":
		res = &object.Float{Value: lVal / rVal}
	case "<":
		res = nativeBoolToBooleanObject(lVal < rVal)
	case ">":
		res = nativeBoolToBooleanObject(lVal > rVal)
	case "==":
		res = nativeBoolToBooleanObject(lVal == rVal)
	case "!=":
		res = nativeBoolToBooleanObject(lVal != rVal)
	case "%":
		res = &object.Float{Value: math.Mod(lVal, rVal)}
	default:
		res = newError("unknown operator for FLOAT %v", operator)
	}
	res.SetMetadata(object.MergeDependencies(left.GetMetadata(), right.GetMetadata()))
	return res
}

func intToFloat(integer object.Object) *object.Float {
	res := &object.Float{Value: float64(integer.(*object.Integer).Value)}
	res.SetMetadata(integer.GetMetadata())
	return res
}

func evalStringInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	lVal := left.(*object.String).Value
	rVal := right.(*object.String).Value

	var res object.Object
	switch operator {
	case "+":
		// dependency assignment handled inside this function
		return addStrings(left, right)
	case "==":
		res = nativeBoolToBooleanObject(lVal == rVal)
	case "!=":
		res = nativeBoolToBooleanObject(lVal != rVal)
	default:
		res = NIL.Copy()
	}
	res.SetMetadata(object.MergeDependencies(left.GetMetadata(), right.GetMetadata()))
	return res
}

func multiplyStrings(str object.Object, integer object.Object) *object.String {
	resStr := ""
	strVal := str.(*object.String).Value
	repeats := int(integer.(*object.Integer).Value)

	// a small dependecy optimization; if the integer is 0, the string is irrelevant
	if repeats <= 0 {
		res := &object.String{Value: ""}
		res.SetMetadata(integer.GetMetadata())
		return res
	}

	for i := 0; i < repeats; i++ {
		resStr += strVal
	}
	res := &object.String{Value: resStr}
	res.SetMetadata(object.MergeDependencies(str.GetMetadata(), integer.GetMetadata()))
	return res
}

func addStrings(left object.Object, right object.Object) *object.String {
	res := &object.String{Value: left.String().Value + right.String().Value}
	res.SetMetadata(object.MergeDependencies(left.GetMetadata(), right.GetMetadata()))
	return res
}

func evalArrayInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	lEls := left.(*object.Array).Elements
	rEls := right.(*object.Array).Elements

	switch operator {
	case "+":
		res := addElements(left.(*object.Array), right.(*object.Array))
		return res
	case "==":
		return nativeBoolToBooleanObject(elComparison(lEls, rEls))
	case "!=":
		return nativeBoolToBooleanObject(!elComparison(lEls, rEls))
	default:
		return NIL
	}
}

func addElements(left *object.Array, right *object.Array) *object.Array {
	elements := make([]object.Object, 0, len(left.Elements)+len(right.Elements))
	for _, el := range left.Elements {
		// we don't really need this but do it for consistency
		elCopy := el.Copy()
		elements = append(elements, elCopy)
	}
	for _, el := range right.Elements {
		elCopy := el.Copy()
		// objects on the right depend on the left array size for their index
		// if the size of the left array shifts, the objects will change index
		// they do not depend on the size of the right array
		elCopy.SetMetadata(object.MergeDependencies(elCopy.GetMetadata(), left.LengthMetadata))
		elements = append(elements, elCopy)
	}
	res := object.Array{Elements: elements}
	// TODO (Peter) really fix runtime here
	res.SetMetadata(object.MergeDependencies(left.GetMetadata(), right.GetMetadata()))
	res.LengthMetadata = object.MergeDependencies(left.LengthMetadata, right.LengthMetadata)
	return &res
}

func elComparison(left []object.Object, right []object.Object) bool {
	if len(left) != len(right) {
		return false
	}
	for i, l := range left {
		// JEM: Inspect() isn't actually going to work here, need to fix
		if l.Inspect() != right[i].Inspect() {
			return false
		}
	}
	return true
}

// JEM: This is pretty neat
func evalBangOperatorExpression(right object.Object) object.Object {
	res := nativeBoolToBooleanObject(!isTruthy(right))
	res.SetMetadata(right.GetMetadata())
	return res
}

func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	if right.Type() == object.INTEGER_OBJ {
		return &object.Integer{Value: -(right.(*object.Integer).Value)}
	} else if right.Type() == object.FLOAT_OBJ {
		return &object.Float{Value: -(right.(*object.Float).Value)}
	}
	return newError("unknown operator: -%s", right.Type())

}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func evalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		res := Eval(ie.Consequence, env)
		return deepCopyObjectAndMergeDeps(res, condition.GetMetadata())
	} else if ie.Alternative != nil {
		res := Eval(ie.Alternative, env)
		return deepCopyObjectAndMergeDeps(res, condition.GetMetadata())
	} else {
		res := NIL.Copy()
		res.SetMetadata(condition.GetMetadata())
		return res
	}
}

func isTruthy(obj object.Object) bool {
	switch obj {
	case NIL:
		return false
	case FALSE:
		return false
	case TRUE:
		return true
	case ZERO_FLOAT:
		return false
	case ZERO_INTEGER:
		return false
	case EMPTY_STRING:
		return false
	default:
		return true
	}
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}
	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}
	return newError("identifier not found: " + node.Value)
}

func evalExpressions(
	exps []ast.Expression,
	env *object.Environment,
) []object.Object {
	var result []object.Object

	for _, e := range exps {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

func evalIndexExpression(left, index object.Object) object.Object {
	if left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ {
		res := evalArrayIndexExpression(left, index)
		res.SetMetadata(object.MergeDependencies(res.GetMetadata(), index.GetMetadata()))
		return res
	}
	return newError("index operator not supported: %s", left.Type())
}

func evalArrayIndexExpression(array, index object.Object) object.Object {
	arrayObject := array.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)

	// Out of bounds
	if idx < 0 || idx > max {
		return NIL
	}

	return arrayObject.Elements[idx]
}

func addDepsToArg(arg object.Object, prefix string) object.Object {
	// (TODO) Peter THIS IS BAD BUT WORKS FOR NOW
	switch arg.(type) {
	case *object.Array:
		// BIG TODO (Peter) we really should not do this preemptively for arrays
		// possibly lazily generate these metadata pieces for when array elements are actually accessed
		// so passing an array of size n to a function doesn't incur a O(n) cost
		arrayCopy := make([]object.Object, len(arg.(*object.Array).Elements))
		for i, e := range arg.(*object.Array).Elements {
			// right now I'm setting metadata for all arrays and sub arrays
			// idk if this is really nessecary maybe we just need leaves...
			// be careful about this in object.go as well
			// (TODO) Peter string concatenation here is a performance no no remove it
			eCopy := addDepsToArg(e, prefix+"|"+strconv.Itoa(i))
			arrayCopy[i] = eCopy
		}
		copyArg := object.Array{Elements: arrayCopy}
		copyArg.SetMetadata(object.TraceMetadata{Dependencies: map[string]bool{prefix: true}})
		// length dependency
		lengthPrefix := prefix + "#"
		copyArg.LengthMetadata = object.TraceMetadata{Dependencies: map[string]bool{lengthPrefix: true}}
		return &copyArg
	default:
		// note we should probably replace this inspect stuff with a real
		// faster hash function at some point?
		copyArg := arg.Copy()
		copyArg.SetMetadata(object.TraceMetadata{Dependencies: map[string]bool{prefix: true}})
		return copyArg
	}
}

func getDepsFromArray(identifier string, pos int, arr []object.Object) object.TraceMetadata {
	// TODO (Peter) strings are not efficient at all!
	// Oh god there's a lot of cringe in this code
	num := 0
	needsLen := false
	for i := pos; i < len(identifier); i++ {
		if identifier[i] == '|' {
			break
		} else if identifier[i] == '#' {
			needsLen = true
			break
		} else {
			num *= 10
			d := int(identifier[i] - 48)
			num += d
			pos++
		}
	}
	fmt.Printf("about to search %d at pos %d\n", num, pos)
	elem := arr[num]
	if pos >= len(identifier) {
		return elem.GetMetadata()
	}
	switch elem.(type) {
	case *object.Array:
		if needsLen {
			return elem.(*object.Array).LengthMetadata
		}
		return getDepsFromArray(identifier, pos+1, elem.(*object.Array).Elements)
	default:
		return elem.GetMetadata()
	}
}

func deepCopyObjectAndMergeDeps(res object.Object, depsToMerge object.TraceMetadata) object.Object {
	// OMG things have really gone to shit
	switch res.(type) {
	case *object.Array:
		arrayCopy := make([]object.Object, len(res.(*object.Array).Elements))
		for i, e := range res.(*object.Array).Elements {
			eCopy := deepCopyObjectAndMergeDeps(e, depsToMerge)
			arrayCopy[i] = eCopy
		}
		copyRes := object.Array{Elements: arrayCopy}
		// array metadata (important in the case of empty arrays!!!)
		copyRes.SetMetadata(object.MergeDependencies(res.GetMetadata(), depsToMerge))
		// length dependency copy
		copyRes.LengthMetadata = object.MergeDependencies(res.(*object.Array).LengthMetadata, depsToMerge)
		return &copyRes
	default:
		// note we should probably replace this inspect stuff with a real
		// faster hash function at some point?
		copyRes := res.Copy()
		copyRes.SetMetadata(object.MergeDependencies(res.GetMetadata(), depsToMerge))
		return copyRes
	}
}

func deepCopyObjectAndTranslateDepsToResult(res object.Object, args []object.Object) object.Object {
	// LOL THIS IS A PERF DUMPSTER FIRE possibly n^4!!!!
	// THERE MUST BE A BETTER WAY!!!
	switch res.(type) {
	case *object.Array:
		arrayCopy := make([]object.Object, len(res.(*object.Array).Elements))
		for i, e := range res.(*object.Array).Elements {
			eCopy := deepCopyObjectAndTranslateDepsToResult(e, args)
			arrayCopy[i] = eCopy
		}
		copyRes := object.Array{Elements: arrayCopy}
		// metadata dependency translation
		// TODO (Peter immediately clean this up)
		translatedMetadataDeps := object.TraceMetadata{}
		for metadataDep, doesDepend := range res.GetMetadata().Dependencies {
			if doesDepend {
				fmt.Printf("searching for array meta: %s\n", metadataDep)
				transMetadataDep := getDepsFromArray(metadataDep, 0, args)
				fmt.Printf("found: %+v\n", transMetadataDep)
				translatedMetadataDeps = object.MergeDependencies(translatedMetadataDeps, transMetadataDep)
			}
		}
		copyRes.SetMetadata(translatedMetadataDeps)
		// length dependency translation
		translatedLenDeps := object.TraceMetadata{}
		for lenDep, doesDepend := range res.(*object.Array).LengthMetadata.Dependencies {
			if doesDepend {
				fmt.Printf("searching for array dep: %s\n", lenDep)
				transLenDep := getDepsFromArray(lenDep, 0, args)
				fmt.Printf("found: %+v\n", transLenDep)
				translatedLenDeps = object.MergeDependencies(translatedLenDeps, transLenDep)
			}
		}
		copyRes.LengthMetadata = translatedLenDeps
		return &copyRes
	default:
		// note we should probably replace this inspect stuff with a real
		// faster hash function at some point?
		copyRes := res.Copy()
		translatedDeps := object.TraceMetadata{}
		for dep, doesDepend := range res.GetMetadata().Dependencies {
			if doesDepend {
				fmt.Printf("searching for dep: %s\n", dep)
				transDep := getDepsFromArray(dep, 0, args)
				fmt.Printf("found: %+v\n", transDep)
				translatedDeps = object.MergeDependencies(translatedDeps, transDep)
			}
		}
		copyRes.SetMetadata(translatedDeps)
		return copyRes
	}
}

func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		if len(fn.Parameters) != len(args) {
			return newError("Supplied %v args, but %v are expected", len(args), len(fn.Parameters))
		}
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)
	case *object.PureFunction:
		if len(fn.Parameters) != len(args) {
			return newError("Supplied %v args, but %v are expected", len(args), len(fn.Parameters))
		}
		// strip the old metadata off of the incoming params
		traceableArgs := make([]object.Object, len(fn.Parameters))
		for i, a := range args {
			traceableArgs[i] = addDepsToArg(a, strconv.Itoa(i))
		}

		extendedEnv := extendPureFunctionEnv(fn, traceableArgs)
		var evaluated object.Object
		// TODO (Peter) should we cache errors?
		// Also this logic could be cleaned up a little
		if val, ok := fn.Get(args); ok {
			evaluated = val
		} else {
			// this code might be a little inconsistent w.r.t errors?
			evaluated = Eval(fn.Body, extendedEnv)
			fnMetadata := evaluated.GetMetadata()
			fn.Set(args, fnMetadata.Dependencies, evaluated)
		}
		// now we assign our dependencies for the function call itself
		// this code might be a little inconsistent w.r.t errors?
		res := unwrapReturnValue(evaluated)
		fmt.Printf("IN: %+v\n", res)
		fmt.Printf("Translating to fn call %s\n", fn.Inspect())
		fmt.Printf("Args:")
		for i, a := range args {
			fmt.Printf("(%d): %+v|", i, a)
		}
		fmt.Printf("\n")
		out := deepCopyObjectAndTranslateDepsToResult(res, args)
		fmt.Printf("OUT: %+v\n", out)
		return out
	case *object.Builtin:
		return fn.Fn(args...)
	default:
		return newError("not a function %s", fn.Type())
	}
}

func extendFunctionEnv(
	fn *object.Function,
	args []object.Object,
) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx])
	}
	return env
}

// JEM: Can you combine this function and the above one?
func extendPureFunctionEnv(
	fn *object.PureFunction,
	args []object.Object,
) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx])
	}
	return env
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.Return); ok {
		return returnValue.Value
	}

	return obj
}
