package evaluator

import "monkey/object"

var builtins = map[string]*object.Builtin{
	"len": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}

			return &object.Integer{Value: int64(len(args[0].String().Value))}
		},
	},
	"type": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}

			return &object.String{Value: string(args[0].Type())}
		},
	},
	"string": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}

			// JEM: WHy can't this be:
			// return args[0].String()
			return &object.String{Value: args[0].Inspect()}
		},
	},
	"bool": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}

			return &object.Boolean{Value: isTruthy(args[0])}
		},
	},
	"int": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}

			switch arg := args[0].(type) {
			case *object.Integer:
				return arg
			case *object.Float:
				return &object.Integer{Value: int64(arg.Value)}
			case *object.Boolean:
				if arg == TRUE {
					return &object.Integer{Value: 1}
				} else {
					return &object.Integer{Value: 0}
				}
			case *object.String:
				return &object.Integer{Value: int64(len(args[0].String().Value))}
			default:
				return newError("can't cast %s to an int", arg.Type())
			}
		},
	},
}
