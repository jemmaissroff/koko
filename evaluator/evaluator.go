package evaluator

import (
	"monkey/ast"
	"monkey/object"

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
		return &object.Return{Value: val}

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
		// TODO (Peter) if one number is actually zero we only need to depend on it!
		res = &object.Integer{Value: lVal * rVal}
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

	switch operator {
	case "+":
		return &object.Float{Value: lVal + rVal}
	case "-":
		return &object.Float{Value: lVal - rVal}
	case "*":
		return &object.Float{Value: lVal * rVal}
	case "/":
		return &object.Float{Value: lVal / rVal}
	case "<":
		return nativeBoolToBooleanObject(lVal < rVal)
	case ">":
		return nativeBoolToBooleanObject(lVal > rVal)
	case "==":
		return nativeBoolToBooleanObject(lVal == rVal)
	case "!=":
		return nativeBoolToBooleanObject(lVal != rVal)
	case "%":
		return &object.Float{Value: math.Mod(lVal, rVal)}
	default:
		return newError("unknown operator for FLOAT %v", operator)
	}
}

func intToFloat(integer object.Object) *object.Float {
	return &object.Float{Value: float64(integer.(*object.Integer).Value)}
}

func evalStringInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	lVal := left.(*object.String).Value
	rVal := right.(*object.String).Value

	switch operator {
	case "+":
		return addStrings(left, right)
	case "==":
		return nativeBoolToBooleanObject(lVal == rVal)
	case "!=":
		return nativeBoolToBooleanObject(lVal != rVal)
	default:
		return NIL
	}
}

func multiplyStrings(str object.Object, integer object.Object) *object.String {
	res := ""
	strVal := str.(*object.String).Value
	for i := 0; i < int(integer.(*object.Integer).Value); i++ {
		res += strVal
	}
	return &object.String{Value: res}
}

func addStrings(left object.Object, right object.Object) *object.String {
	return &object.String{Value: left.String().Value + right.String().Value}
}

func evalArrayInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	lEls := left.(*object.Array).Elements
	rEls := right.(*object.Array).Elements

	switch operator {
	case "+":
		return addElements(lEls, rEls)
	case "==":
		return nativeBoolToBooleanObject(elComparison(lEls, rEls))
	case "!=":
		return nativeBoolToBooleanObject(!elComparison(lEls, rEls))
	default:
		return NIL
	}
}

func addElements(left []object.Object, right []object.Object) *object.Array {
	elements := make([]object.Object, 0, len(left)+len(right))
	for _, el := range left {
		elements = append(elements, el)
	}
	for _, el := range right {
		elements = append(elements, el)
	}
	return &object.Array{Elements: elements}
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
	return nativeBoolToBooleanObject(!isTruthy(right))
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
		res.SetMetadata(object.MergeDependencies(res.GetMetadata(), condition.GetMetadata()))
		return res
	} else if ie.Alternative != nil {
		res := Eval(ie.Alternative, env)
		res.SetMetadata(object.MergeDependencies(res.GetMetadata(), condition.GetMetadata()))
		return res
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
		return evalArrayIndexExpression(left, index)
	} else {
		return newError("index operator not supported: %s", left.Type())
	}
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
			deps := make(map[int]bool)
			deps[i] = true
			traceableArgs[i] = a.Copy()
			traceableArgs[i].SetMetadata(object.TraceMetadata{Dependencies: deps})
		}

		extendedEnv := extendPureFunctionEnv(fn, traceableArgs)
		var evaluated object.Object
		if val, ok := fn.Get(args); ok {
			evaluated = val
		} else {
			evaluated = Eval(fn.Body, extendedEnv)
			fn.Set(args, evaluated)
		}
		fmt.Printf("deps: %+v\n", evaluated.GetMetadata())
		res := unwrapReturnValue(evaluated)
		// now we assign our dependencies for the function call itself
		// let b = pfn(x, y) { if (x % 2 == 0) { a(x, 2, y) } else { 1 } }
		fnMetadata := res.GetMetadata()
		callMetadata := object.TraceMetadata{}
		for i, a := range args {
			if obj, ok := fnMetadata.Dependencies[i]; ok && obj == true {
				fmt.Printf("arg: %+v\n", a)
				callMetadata = object.MergeDependencies(callMetadata, a.GetMetadata())
			}
		}
		res.SetMetadata(callMetadata)
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
		return returnValue.Value
	}

	return obj
}
