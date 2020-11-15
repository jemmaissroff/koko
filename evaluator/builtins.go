package evaluator

import (
	"math/rand"
	"monkey/object"
)

var builtins = map[string]*object.Builtin{
	"len": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if err := validateNumberOfArgs(1, args); err != NIL {
				return err
			}

			var value int64
			switch args[0].(type) {
			case *object.Array:
				value = int64(len(args[0].(*object.Array).Elements))
				res := object.Integer{Value: value}
				res.SetMetadata(args[0].(*object.Array).LengthMetadata)
				return &res
			default:
				// TODO (Peter) add this case to dep tracking
				value = int64(len(args[0].String().Value))
				return &object.Integer{Value: value}
			}
		},
	},
	"type": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if err := validateNumberOfArgs(1, args); err != NIL {
				return err
			}

			return &object.String{Value: string(args[0].Type())}
		},
	},
	"string": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if err := validateNumberOfArgs(1, args); err != NIL {
				return err
			}

			// JEM: WHy can't this be:
			// return args[0].String()
			return &object.String{Value: args[0].Inspect()}
		},
	},
	"bool": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if err := validateNumberOfArgs(1, args); err != NIL {
				return err
			}

			return &object.Boolean{Value: isTruthy(args[0])}
		},
	},
	"int": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if err := validateNumberOfArgs(1, args); err != NIL {
				return err
			}

			switch arg := args[0].(type) {
			case *object.Integer:
				return arg
			case *object.Float:
				res := &object.Integer{Value: int64(arg.Value)}
				// (TODO) Peter: possibly expand this pattern
				res.SetMetadata(arg.GetMetadata())
				return res
			case *object.Boolean:
				if arg == TRUE {
					return &object.Integer{Value: 1}
				} else {
					return &object.Integer{Value: 0}
				}
			case *object.String:
				return &object.Integer{Value: int64(len(arg.String().Value))}
			default:
				return newError("can't cast %s to an int", arg.Type())
			}
		},
	},
	"first": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if err := validateNumberOfArgs(1, args); err != NIL {
				return err
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `first` must be ARRAY, got %s", args[0].Type())
			}

			arr := args[0].(*object.Array)
			if len(arr.Elements) > 0 {
				return arr.Elements[0]
			}
			return NIL
		},
	},
	"last": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if err := validateNumberOfArgs(1, args); err != NIL {
				return err
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `first` must be ARRAY, got %s", args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)
			if length > 0 {
				return arr.Elements[length-1]
			}
			return NIL
		},
	},
	"rest": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if err := validateNumberOfArgs(1, args); err != NIL {
				return err
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `first` must be ARRAY, got %s", args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)
			if length > 0 {
				newElements := make([]object.Object, length-1, length-1)
				copy(newElements, arr.Elements[1:length])
				res := &object.Array{Elements: newElements}
				res.LengthMetadata = arr.LengthMetadata
				return res
				// let f = pfn(a, b) { if (a > 0) { if (a > 0) { rest(b) + [a] } } }
			}
			return NIL
		},
	},
	"push": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if err := validateNumberOfArgs(2, args); err != NIL {
				return err
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `first` must be ARRAY, got %s", args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)
			if length > 0 {
				newElements := make([]object.Object, length+1, length+1)
				copy(newElements, arr.Elements)
				newElements[length] = args[1]
				return &object.Array{Elements: newElements}
			}
			return NIL
		},
	},
	"take": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if err := validateNumberOfArgs(2, args); err != NIL {
				return err
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `take` must be ARRAY, got %s", args[0].Type())
			}
			if args[1].Type() != object.INTEGER_OBJ {
				return newError("argument to `take` must be INTEGER, got %s", args[1].Type())
			}

			arr := args[0].(*object.Array)
			takeNum := args[1].(*object.Integer).Value
			if takeNum < 0 || takeNum >= int64(len(arr.Elements)) {
				return newError("Invalid index %d for %v", takeNum, arr.Inspect())
			}
			return &object.Array{Elements: arr.Elements[0:takeNum]}
		},
	},
	"drop": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if err := validateNumberOfArgs(2, args); err != NIL {
				return err
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `take` must be ARRAY, got %s", args[0].Type())
			}
			if args[1].Type() != object.INTEGER_OBJ {
				return newError("argument to `take` must be INTEGER, got %s", args[1].Type())
			}

			arr := args[0].(*object.Array)
			takeNum := args[1].(*object.Integer).Value
			if takeNum < 0 || takeNum > int64(len(arr.Elements)) {
				return newError("Invalid index %d for %v", takeNum, arr.Inspect())
			}
			return &object.Array{Elements: arr.Elements[takeNum:]}
		},
	},
	"rando": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if err := validateNumberOfArgs(1, args); err != NIL {
				return err
			}

			arg := args[0].(*object.Integer)
			return &object.Integer{Value: int64(rand.Intn(int(arg.Value)))}
		},
	},
}

func validateNumberOfArgs(length int, args []object.Object) object.Object {
	if len(args) != length {
		return newError("wrong number of arguments. got=%d, want=%d",
			len(args), length)
	}

	return NIL
}
