package evaluator

import (
	"koko/ast"
	"koko/object"
	"strconv"

	"fmt"
	"math"
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
		return &object.Integer{Value: node.Value}
	case *ast.FloatLiteral:
		return &object.Float{Value: node.Value}
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
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

		res := applyFunction(function, args)
		//fmt.Printf("in m: %+v\n", res.GetMetadata())
		out := deepCopyObjectAndTranslateDepsToResult(res, args)
		//fmt.Printf("out m: %+v\n", out.GetMetadata())
		return out
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
	case *ast.HashLiteral:
		return evalHashLiteral(node, env)
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
	if operator == "==" {
		// TODO (Peter) in the future make obejct comparisons more granular
		res := nativeBoolToBooleanObject(left.Equal(right)).Copy()
		res.SetMetadata(object.MergeDependencies(left.GetMetadata(), right.GetMetadata()))
		return res
	} else if operator == "!=" {
		res := nativeBoolToBooleanObject(!left.Equal(right))
		res.SetMetadata(object.MergeDependencies(left.GetMetadata(), right.GetMetadata()))
		return res
	}
	switch {
	case left.Type() == object.ARRAY_OBJ:
		switch {
		case right.Type() == object.ARRAY_OBJ:
			return evalArrayInfixExpression(operator, left, right)
		}
	case left.Type() == object.HASH_OBJ:
		switch {
		case right.Type() == object.HASH_OBJ:
			return evalHashInfixExpression(operator, left, right)
		}
	case left.Type() == object.STRING_OBJ:
		switch {
		case right.Type() == object.STRING_OBJ:
			return evalStringInfixExpression(operator, left, right)
		case operator == "*" && right.Type() == object.INTEGER_OBJ:
			return multiplyStrings(left, right)
		case operator == "+":
			return addStrings(left, right)
		}
	case right.Type() == object.STRING_OBJ:
		switch {
		case operator == "*" && left.Type() == object.INTEGER_OBJ:
			return multiplyStrings(right, left)
		case operator == "+":
			return addStrings(left, right)
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
	if operator == "+" {
		// dependency assignment handled inside this function
		return addStrings(left, right)
	}
	return newError("Unsupported Operator %s for strings", operator)
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
	switch operator {
	case "+":
		res := addElements(left.(*object.Array), right.(*object.Array))
		return res
	default:
		return newError("Unsupported Operator %s for arrays", operator)
	}
}

func addElements(left *object.Array, right *object.Array) *object.Array {
	elements := make([]object.Object, 0, len(left.Elements)+len(right.Elements))
	for _, el := range left.Elements {
		// TODO (Peter) is this a big problem?
		elCopy := el
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

func evalHashInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	lPairs := left.(*object.Hash).Pairs
	rPairs := right.(*object.Hash).Pairs

	switch operator {
	case "+":
		return addPairs(lPairs, rPairs)
	case "-":
		return subtractPairs(lPairs, rPairs)
	default:
		return newError("Unsupported Operator %s for hashes", operator)
	}
}

func addPairs(left map[object.HashKey]object.HashPair, right map[object.HashKey]object.HashPair) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)

	for k, v := range left {
		pairs[k] = v
	}
	for k, v := range right {
		pairs[k] = v
	}

	return &object.Hash{Pairs: pairs}
}

func subtractPairs(left map[object.HashKey]object.HashPair, right map[object.HashKey]object.HashPair) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)

	for k, v := range left {
		pairs[k] = v
	}
	for k, v := range right {
		if pairs[k].Value.Equal(v.Value) {
			delete(pairs, k)
		}
	}

	return &object.Hash{Pairs: pairs}
}

// JEM: This is pretty neat
func evalBangOperatorExpression(right object.Object) object.Object {
	res := nativeBoolToBooleanObject(!object.Bool(right))
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
		return (object.TRUE.Copy()).(*object.Boolean)
	}
	return (object.FALSE.Copy()).(*object.Boolean)
}

func evalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(ie.Condition, env)
	conditionMetadataCopy := object.MergeDependencies(object.TraceMetadata{}, condition.GetMetadata())
	if isError(condition) {
		return condition
	}
	if object.Bool(condition) {
		res := Eval(ie.Consequence, env)
		return deepCopyObjectAndMergeDeps(res, conditionMetadataCopy)
	} else if ie.Alternative != nil {
		res := Eval(ie.Alternative, env)
		return deepCopyObjectAndMergeDeps(res, conditionMetadataCopy)
	} else {
		res := object.NIL.Copy()
		res.SetMetadata(conditionMetadataCopy)
		return res
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
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		res := evalArrayIndexExpression(left, index)
		res.SetMetadata(object.MergeDependencies(res.GetMetadata(), index.GetMetadata()))
		return res
	case left.Type() == object.HASH_OBJ:
		return evalHashIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
	return newError("index operator not supported: %s", left.Type())
}

func evalArrayIndexExpression(array, index object.Object) object.Object {
	arrayObject := array.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)

	// Out of bounds
	if idx < 0 || idx > max {
		return object.NIL
	}

	return arrayObject.Elements[idx]
}

func deepCopyAndAddDepsToArg(arg object.Object, prefix string) object.Object {
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
			eCopy := deepCopyAndAddDepsToArg(e, prefix+"|"+strconv.Itoa(i))
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
		// TODO (Peter) could probably replace this with a throw...
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
				//fmt.Printf("searching for array meta: %s\n", metadataDep)
				transMetadataDep := getDepsFromArray(metadataDep, 0, args)
				//fmt.Printf("found: %+v\n", transMetadataDep)
				translatedMetadataDeps = object.MergeDependencies(translatedMetadataDeps, transMetadataDep)
			}
		}
		copyRes.SetMetadata(translatedMetadataDeps)
		// length dependency translation
		translatedLenDeps := object.TraceMetadata{}
		for lenDep, doesDepend := range res.(*object.Array).LengthMetadata.Dependencies {
			if doesDepend {
				//fmt.Printf("searching for array dep: %s\n", lenDep)
				transLenDep := getDepsFromArray(lenDep, 0, args)
				//fmt.Printf("found: %+v\n", transLenDep)
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
				//fmt.Printf("searching for dep: %s\n", dep)
				transDep := getDepsFromArray(dep, 0, args)
				//fmt.Printf("found: %+v\n", transDep)
				translatedDeps = object.MergeDependencies(translatedDeps, transDep)
			}
		}
		copyRes.SetMetadata(translatedDeps)
		return copyRes
	}
}

func applyPureFunction(fn *object.PureFunction, args []object.Object) object.Object {
	if len(fn.Parameters) != len(args) {
		return newError("Supplied %v args, but %v are expected", len(args), len(fn.Parameters))
	}

	extendedEnv := extendPureFunctionEnv(fn, args)
	var res object.Object
	// TODO (Peter) should we cache errors?
	if val, ok := fn.Get(args); ok {
		res = val
	} else {
		// this code might be a little inconsistent w.r.t errors?
		res = unwrapReturnValue(Eval(fn.Body, extendedEnv))
		fn.Set(args, res)
	}
	return res
}

func evalHashIndexExpression(hash, index object.Object) object.Object {
	hashObject := hash.(*object.Hash)
	key, ok := index.(object.Hashable)
	if !ok {
		return newError("unusable as hash key: %s", index.Type())
	}
	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return object.NIL
	}
	return pair.Value
}

func applyFunction(fn object.Object, args []object.Object) object.Object {
	// strip the old metadata off of the incoming params, for tracing
	traceableArgs := make([]object.Object, len(args))
	for i, a := range args {
		traceableArgs[i] = deepCopyAndAddDepsToArg(a, strconv.Itoa(i))
	}
	switch fn := fn.(type) {
	case *object.Function:
		if len(fn.Parameters) != len(args) {
			return newError("Supplied %v args, but %v are expected", len(args), len(fn.Parameters))
		}
		extendedEnv := extendFunctionEnv(fn, traceableArgs)
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)
	case *object.PureFunction:
		res := applyPureFunction(fn, traceableArgs)
		return res
		//return deepCopyObjectAndTranslateDepsToResult(res, args)
	case *object.Builtin:
		return fn.Fn(traceableArgs...)
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

func evalHashLiteral(
	node *ast.HashLiteral,
	env *object.Environment,
) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)
	for keyNode, valueNode := range node.Pairs {
		key := Eval(keyNode, env)

		if isError(key) {
			return key
		}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %s", key.Type())
		}
		value := Eval(valueNode, env)
		if isError(value) {
			return value
		}
		hashed := hashKey.HashKey()
		pairs[hashed] = object.HashPair{Key: key, Value: value}
	}
	return &object.Hash{Pairs: pairs}
}
