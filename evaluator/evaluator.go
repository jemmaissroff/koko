package evaluator

import (
	"monkey/ast"
	"monkey/object"

	"fmt"
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
	case left.Type() == object.STRING_OBJ:
		switch {
		case right.Type() == object.STRING_OBJ:
			return evalStringInfixExpression(operator, left, right)
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

	switch operator {
	case "+":
		return &object.Integer{Value: lVal + rVal}
	case "-":
		return &object.Integer{Value: lVal - rVal}
	case "*":
		return &object.Integer{Value: lVal * rVal}
	case "/":
		return evalFloatInfixExpression(operator, intToFloat(left), intToFloat(right))
	case "<":
		return nativeBoolToBooleanObject(lVal < rVal)
	case ">":
		return nativeBoolToBooleanObject(lVal > rVal)
	case "==":
		return nativeBoolToBooleanObject(lVal == rVal)
	case "!=":
		return nativeBoolToBooleanObject(lVal != rVal)
	default:
		return NIL
	}
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
	default:
		return NIL
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

func addStrings(left object.Object, right object.Object) *object.String {
	return &object.String{Value: left.String().Value + right.String().Value}
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

	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	} else {
		return NIL
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
