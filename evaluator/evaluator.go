package evaluator

import (
	"koko/ast"
	"koko/object"

	"fmt"
	"math"
)

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {

	case *ast.Program:
		res := evalProgram(node, env)
		res.SetCreatorNode(node)
		return res

	case *ast.BlockStatement:
		res := evalBlockStatement(node, env)
		res.SetCreatorNode(node)
		return res

	case *ast.ExpressionStatement:
		res := Eval(node.Expression, env)
		res.SetCreatorNode(node)
		return res

	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		res := &object.Return{Value: val}
		res.AddDependency(val)
		return res

	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value, ASTCreator: node}
	case *ast.FloatLiteral:
		return &object.Float{Value: node.Value, ASTCreator: node}
	case *ast.StringLiteral:
		return &object.String{Value: node.Value, ASTCreator: node}
	case *ast.Boolean:
		res := nativeBoolToBooleanObject(node.Value)
		res.SetCreatorNode(node)
		return res
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		res := evalPrefixExpression(node.Operator, right)
		res.SetCreatorNode(node)
		return res
	case *ast.InfixExpression:
		// TODO (Peter) add short circuiting
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		res := evalInfixExpression(node.Operator, left, right)
		res.SetCreatorNode(node)
		return res
	case *ast.IfExpression:
		res := evalIfExpression(node, env)
		res.SetCreatorNode(node)
		return res
	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		res := env.Set(node.Name.Value, val)
		res.SetCreatorNode(node)
		return res
	case *ast.ImportStatement:
		// TODO: Decide on whether to output the result of the file?
		// Right now this return nil means that the last line of the
		// file won't appear in the repl
		LoadProgramFromFile(node.Value, env)
		return nil
	case *ast.Identifier:
		res := evalIdentifier(node, env)
		res.SetCreatorNode(node)
		return res
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		res := &object.Function{Parameters: params, Env: env, Body: body}
		res.SetCreatorNode(node)
		return res
	case *ast.PureFunctionLiteral:
		params := node.Parameters
		body := node.Body
		res := object.NewPureFunction(params, env, body)
		res.SetCreatorNode(node)
		return res
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
		res.SetCreatorNode(node)
		return res
	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		res := object.CreateArray(elements)
		res.SetCreatorNode(node)
		return res
	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		res := evalIndexExpression(left, index)
		res.SetCreatorNode(node)
		return res
	case *ast.HashLiteral:
		res := evalHashLiteral(node, env)
		res.SetCreatorNode(node)
		return res
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
		res.AddDependency(left)
		res.AddDependency(right)
		return res
	} else if operator == "!=" {
		res := nativeBoolToBooleanObject(!left.Equal(right))
		res.AddDependency(left)
		res.AddDependency(right)
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
		// This is a short circuit dependency
		if lVal == 0 {
			res.AddDependency(left)
			return res
		} else if rVal == 0 {
			res.AddDependency(right)
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
	res.AddDependency(left)
	res.AddDependency(right)
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
		// This is a short circuit dependency
		if lVal == 0 {
			res.AddDependency(left)
			return res
		} else if rVal == 0 {
			res.AddDependency(right)
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
	res.AddDependency(left)
	res.AddDependency(right)
	return res
}

func intToFloat(integer object.Object) *object.Float {
	res := &object.Float{Value: float64(integer.(*object.Integer).Value)}
	res.AddDependency(integer)
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
		res.AddDependency(integer)
		return res
	}

	// a small dependecy optimization; if the str is "", then the integer is irrelevant
	if strVal == "" {
		res := &object.String{Value: ""}
		res.AddDependency(str)
		return res
	}

	for i := 0; i < repeats; i++ {
		resStr += strVal
	}
	res := &object.String{Value: resStr}
	res.AddDependency(str)
	res.AddDependency(integer)
	return res
}

func addStrings(left object.Object, right object.Object) *object.String {
	res := &object.String{Value: left.String().Value + right.String().Value}
	res.AddDependency(left)
	res.AddDependency(right)
	return res
}

func evalArrayInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	switch operator {
	case "+":
		return addElements(left.(*object.Array), right.(*object.Array))
	default:
		return newError("Unsupported Operator %s for arrays", operator)
	}
}

func addElements(left *object.Array, right *object.Array) *object.Array {
	elements := make([]object.Object, 0, len(left.Elements)+len(right.Elements))
	// NOTE (Peter) this should be okay instead of calling object.CreateArray
	// But be very careful when changing this for dependency reasons
	res := object.Array{}
	for _, el := range left.Elements {
		elCopy := el.Copy()
		res.AddDependency(elCopy)
		elements = append(elements, elCopy)
	}
	for _, el := range right.Elements {
		elCopy := el.Copy()
		// objects on the right depend on the left array size for their index
		// if the size of the left array shifts, the objects will change index
		// they do not depend on the size of the right array
		elCopy.AddDependency(&left.Length)
		if elArr, ok := elCopy.(*object.Array); ok {
			elArr.AddOffsetDependency(&left.Length)
		}
		if elHash, ok := elCopy.(*object.Hash); ok {
			elHash.AddOffsetDependency(&left.Length)
		}
		res.AddDependency(elCopy)
		elements = append(elements, elCopy)
	}
	// TODO (Peter) do we need to do anything with the offsets here???
	res.Elements = elements
	res.Length.Value = int64(len(elements))
	res.Length.AddDependency(&left.Length)
	res.Length.AddDependency(&right.Length)
	return &res
}

func evalHashInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	lHash := left.(*object.Hash)
	rHash := right.(*object.Hash)

	switch operator {
	case "+":
		return addPairs(lHash, rHash)
	case "-":
		return subtractPairs(lHash, rHash)
	default:
		return newError("Unsupported Operator %s for hashes", operator)
	}
}

func addPairs(left *object.Hash, right *object.Hash) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)

	for k, v := range left.Pairs {
		pairs[k] = v
	}
	for k, v := range right.Pairs {
		pairs[k] = v
	}

	res := object.CreateHash(pairs)
	res.Length.AddDependency(&left.Length)
	res.Length.AddDependency(&right.Length)
	return res
}

func subtractPairs(left *object.Hash, right *object.Hash) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)

	for k, v := range left.Pairs {
		pairs[k] = v
	}
	for k, v := range right.Pairs {
		if pairs[k].Value.Equal(v.Value) {
			delete(pairs, k)
		}
	}

	res := object.CreateHash(pairs)
	res.Length.AddDependency(&left.Length)
	res.Length.AddDependency(&right.Length)
	return res
}

// JEM: This is pretty neat
func evalBangOperatorExpression(right object.Object) object.Object {
	res := nativeBoolToBooleanObject(!object.Bool(right))
	res.AddDependency(right)
	return res
}

func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	if right.Type() == object.INTEGER_OBJ {
		res := &object.Integer{Value: -(right.(*object.Integer).Value)}
		res.AddDependency(right)
		return res
	} else if right.Type() == object.FLOAT_OBJ {
		res := &object.Float{Value: -(right.(*object.Float).Value)}
		res.AddDependency(right)
		return res
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
	condition := Eval(ie.Condition, env).Copy()
	if isError(condition) {
		return condition
	}
	if object.Bool(condition) {
		res := Eval(ie.Consequence, env).Copy()
		res.AddDependency(condition)
		return res
	} else if ie.Alternative != nil {
		res := Eval(ie.Alternative, env).Copy()
		res.AddDependency(condition)
		return res
	} else {
		res := object.NIL.Copy()
		res.AddDependency(condition)
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
		res.AddDependency(index)
		// propegate dependencies
		res.AddDependency(&left.(*object.Array).Offset)
		if arrRes, ok := res.(*object.Array); ok {
			arrRes.AddOffsetDependency(&left.(*object.Array).Offset)
		}
		if hashRes, ok := res.(*object.Hash); ok {
			hashRes.AddOffsetDependency(&left.(*object.Array).Offset)
		}
		return res
	case left.Type() == object.HASH_OBJ:
		return evalHashIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalArrayIndexExpression(array, index object.Object) object.Object {
	arrayObject := array.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)

	// Out of bounds
	if idx < 0 || idx > max {
		res := object.NIL.Copy()
		res.AddDependency(index)
		return res
	}

	res := arrayObject.Elements[idx].Copy()
	res.AddDependency(index)
	return res
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
		res := object.NIL.Copy()
		res.AddDependency(index)
		res.AddDependency(&hash.(*object.Hash).Offset)
		return res
	}
	res := pair.Value.Copy()
	res.AddDependency(index)
	// propegate offset dependencies
	if arrRes, ok := res.(*object.Array); ok {
		arrRes.AddOffsetDependency(&hash.(*object.Hash).Offset)
	}
	if hashRes, ok := res.(*object.Hash); ok {
		hashRes.AddOffsetDependency(&hash.(*object.Hash).Offset)
	}
	res.AddDependency(&hash.(*object.Hash).Offset)
	return res
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
		res := applyPureFunction(fn, args)
		return res
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
		// TODO Peter this graph is a little over complex
		res := returnValue.Value.Copy()
		res.AddDependency(obj)
		return res
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
	return object.CreateHash(pairs)
}
