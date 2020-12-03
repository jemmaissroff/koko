package evaluator

import (
	"koko/object"
	"math/rand"
	"sort"
	"strconv"
)

var builtins map[string]*object.Builtin

// Builtins is in an init so that its creation is deferred and we can define
// the builtin function builtins without an initialization loop
func init() {
	builtins = map[string]*object.Builtin{
		"builtins": &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if err := validateNumberOfArgs(0, args); err != object.NIL {
					return err
				}
				builtinStringKeys := make([]string, 0, len(builtins))
				for builtinStringKey := range builtins {
					builtinStringKeys = append(builtinStringKeys, builtinStringKey)
				}
				sort.Strings(builtinStringKeys)
				builtinKeys := make([]object.Object, 0, len(builtins))
				for _, builtinFunction := range builtinStringKeys {
					builtinKeys = append(
						builtinKeys,
						&object.String{Value: builtinFunction},
					)
				}
				return &object.Array{Elements: builtinKeys}
			},
		},
		"deps": &object.Builtin{
			// this is an internal function which allows us to debug dependency tracing
			// TODO (Peter) possibly add an interface which pretty prints deps for the user
			// directly returns an object with untranslated dependency information
			// this object should actually never be used in the program because it can poison the dependency
			// tracing system
			// this should probably be in buildins but cannot be b/c of circular dependencies
			Fn: func(args ...object.Object) object.Object {
				if len(args) < 1 {
					return newError("wrong number of arguments. got=%d, need at least=%d",
						len(args), 1)
				}

				fn := args[0]

				fnRes := applyFunction(fn, args[1:])
				res := object.DebugTraceMetadata{}
				res.SetDebugMetadata(fnRes.GetMetadata())
				return &res
			},
		},
		"len": &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if err := validateNumberOfArgs(1, args); err != object.NIL {
					return err
				}

				var value int64
				switch args[0].(type) {
				case *object.Array:
					value = int64(len(args[0].(*object.Array).Elements))
					res := object.Integer{Value: value}
					res.SetMetadata(args[0].(*object.Array).LengthMetadata)
					return &res
				case *object.Hash:
					value = int64(len(args[0].(*object.Hash).Pairs))
				default:
					value = int64(len(args[0].String().Value))
				}
				return &object.Integer{Value: value}
			},
		},
		"type": &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if err := validateNumberOfArgs(1, args); err != object.NIL {
					return err
				}

				return &object.String{Value: string(args[0].Type())}
			},
		},
		"string": &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if err := validateNumberOfArgs(1, args); err != object.NIL {
					return err
				}

				// JEM: WHy can't this be:
				// return args[0].String()
				return &object.String{Value: args[0].Inspect()}
			},
		},
		"array": &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if err := validateNumberOfArgs(1, args); err != object.NIL {
					return err
				}

				elements := make([]object.Object, 0, 1)

				switch arg := args[0].(type) {
				case *object.String:
					for _, val := range arg.Value {
						elements = append(elements, &object.String{Value: string(val)})
					}
				case *object.Array:
					return arg
				default:
					elements = append(elements, arg)
				}

				return &object.Array{Elements: elements}
			},
		},
		"bool": &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if err := validateNumberOfArgs(1, args); err != object.NIL {
					return err
				}

				return &object.Boolean{Value: object.Bool(args[0])}
			},
		},
		"int": &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if err := validateNumberOfArgs(1, args); err != object.NIL {
					return err
				}

				switch arg := args[0].(type) {
				case *object.Integer:
					return arg
				case *object.Float:
					return &object.Integer{Value: int64(arg.Value)}
				case *object.Boolean:
					if arg == object.TRUE {
						return &object.Integer{Value: 1}
					} else {
						return &object.Integer{Value: 0}
					}
				case *object.String:
					i, err := strconv.ParseInt(arg.String().Value, 10, 64)
					if err != nil {
						return object.NIL
					} else {
						return &object.Integer{Value: i}
					}
				default:
					return newError("can't cast %s to an int", arg.Type())
				}
			},
		},
		"float": &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if err := validateNumberOfArgs(1, args); err != object.NIL {
					return err
				}

				switch arg := args[0].(type) {
				case *object.Integer:
					return &object.Float{Value: float64(arg.Value)}
				case *object.Float:
					return arg
				case *object.Boolean:
					if arg == object.TRUE {
						return &object.Float{Value: 1}
					} else {
						return &object.Float{Value: 0}
					}
				case *object.String:
					f, err := strconv.ParseFloat(arg.String().Value, 64)
					if err != nil {
						return object.NIL
					} else {
						return &object.Float{Value: f}
					}
				default:
					return newError("can't cast %s to a float", arg.Type())
				}
			},
		},
		"first": &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if err := validateNumberOfArgs(1, args); err != object.NIL {
					return err
				}

				if args[0].Type() != object.ARRAY_OBJ {
					return newError("argument to `first` must be ARRAY, got %s", args[0].Type())
				}

				arr := args[0].(*object.Array)
				if len(arr.Elements) > 0 {
					return arr.Elements[0]
				}
				return object.NIL
			},
		},
		"last": &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if err := validateNumberOfArgs(1, args); err != object.NIL {
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
				return object.NIL
			},
		},
		"rest": &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if err := validateNumberOfArgs(1, args); err != object.NIL {
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
					return &object.Array{Elements: newElements}
				}
				return object.NIL
			},
		},
		"push": &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if err := validateNumberOfArgs(2, args); err != object.NIL {
					return err
				}

				if args[0].Type() != object.ARRAY_OBJ {
					return newError("argument to `first` must be ARRAY, got %s", args[0].Type())
				}

				arr := args[0].(*object.Array)
				length := len(arr.Elements)
				newElements := make([]object.Object, length+1, length+1)
				copy(newElements, arr.Elements)
				newElements[length] = args[1]
				return &object.Array{Elements: newElements}
			},
		},
		"take": &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if err := validateNumberOfArgs(2, args); err != object.NIL {
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
				if err := validateNumberOfArgs(2, args); err != object.NIL {
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
		"keys": &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if err := validateNumberOfArgs(1, args); err != object.NIL {
					return err
				}

				if args[0].Type() != object.HASH_OBJ {
					return newError("argument to `keys` must be HASH, got %s", args[0].Type())
				}

				hash := args[0].(*object.Hash)
				elements := make([]object.Object, 0, len(hash.Pairs))
				for _, hashPair := range hash.Pairs {
					elements = append(elements, hashPair.Key)
				}
				return &object.Array{Elements: elements}
			},
		},
		"values": &object.Builtin{
			// JEM: Possible refactor to pull these out
			Fn: func(args ...object.Object) object.Object {
				if err := validateNumberOfArgs(1, args); err != object.NIL {
					return err
				}

				if args[0].Type() != object.HASH_OBJ {
					return newError("argument to `keys` must be HASH, got %s", args[0].Type())
				}

				hash := args[0].(*object.Hash)
				elements := make([]object.Object, 0, len(hash.Pairs))
				for _, hashPair := range hash.Pairs {
					elements = append(elements, hashPair.Value)
				}
				return &object.Array{Elements: elements}
			},
		},
		"rando": &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if err := validateNumberOfArgs(1, args); err != object.NIL {
					return err
				}

				if args[0].Type() != object.INTEGER_OBJ {
					return newError("argument to `rando` must be INTEGER, got %s", args[0].Type())
				}

				arg := args[0].(*object.Integer)
				if arg.Value < 1 {
					return newError("argument to `rando` must be at least 1, got %v", arg.Value)
				}
				return &object.Integer{Value: int64(rand.Intn(int(arg.Value)))}
			},
		},
	}
}

func validateNumberOfArgs(length int, args []object.Object) object.Object {
	if len(args) != length {
		return newError("wrong number of arguments. got=%d, want=%d",
			len(args), length)
	}

	return object.NIL
}
