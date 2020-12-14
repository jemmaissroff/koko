package evaluator

import (
	"fmt"
	"io/ioutil"
	"koko/ast"
	"koko/object"
	"math/rand"
	"sort"
	"strconv"
)

var builtins map[string]*object.Builtin

// NOTE (Peter) this function is techincal debt
func getObjRelatedDependenciesInTrace(obj object.Object, trace map[object.Object]bool, prefix string, out map[string]bool) {
	if _, ok := trace[obj]; ok {
		out[prefix] = true
	}
	if arrObj, ok := obj.(*object.Array); ok {
		if _, ok := trace[&arrObj.Length]; ok {
			out[prefix+"#"] = true
		}
		for i, subObj := range arrObj.Elements {
			getObjRelatedDependenciesInTrace(subObj, trace, prefix+"|"+fmt.Sprint(i), out)
		}
	}
	if hashObj, ok := obj.(*object.Hash); ok {
		if _, ok := trace[&hashObj.Length]; ok {
			out[prefix+"#"] = true
		}
		for _, pair := range hashObj.Pairs {
			getObjRelatedDependenciesInTrace(pair.Value, trace, prefix+"|@"+fmt.Sprint(pair.Key.Inspect()), out)
		}
	}
}

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
				return object.CreateArray(builtinKeys)
			},
		},
		"print": &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return newError("wrong number of arguments. got=%d, need 1",
						len(args))
				}

				res := args[0]
				fmt.Println(res.Inspect())
				return res
			},
		},
		"deps": &object.Builtin{
			// NOTE this function is legacy to support tests from the old version
			// TODO (Peter) update this to a better version later
			Fn: func(args ...object.Object) object.Object {
				if len(args) < 1 {
					return newError("wrong number of arguments. got=%d, need at least=%d",
						len(args), 1)
				}

				fn := args[0]
				fnRes := applyFunction(fn, args[1:])

				traceDeps := object.GetAllDependencies(fnRes)
				argDeps := make(map[string]bool)
				for i, arg := range args[1:] {
					getObjRelatedDependenciesInTrace(arg, traceDeps, fmt.Sprint(i), argDeps)
				}

				res := object.DebugTraceMetadata{DebugMetadata: argDeps}
				res.AddDependency(fnRes)
				return &res
			},
		},
		"dep_diagraph": {
			// NOTE this function is legacy to support tests from the old version
			// TODO (Peter) update this to a better version later
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return newError("wrong number of arguments. got=%d, need 1",
						len(args))
				}

				arg := args[0]

				res := object.String{Value: object.GetAllDependenciesToDotLang(arg)}
				res.AddDependency(arg)

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
					res.AddDependency(&args[0].(*object.Array).Length)
					return &res
				case *object.Hash:
					value = int64(len(args[0].(*object.Hash).Pairs))
					res := object.Integer{Value: value}
					res.AddDependency(&args[0].(*object.Hash).Length)
					return &res
				default:
					value = int64(len(args[0].String().Value))
					res := object.Integer{Value: value}
					res.AddDependency(args[0])
					return &res
				}
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
						e := &object.String{Value: string(val)}
						e.AddDependency(arg)
						elements = append(elements, e)
					}
				case *object.Array:
					return arg
				default:
					elements = append(elements, arg)
				}

				// NOTE (Peter) this should be okay instead of calling object.CreateArray
				// But be very careful when changing this for dependency reasons
				res := object.Array{Elements: elements}
				res.Offset.ASTCreator = &ast.StringLiteral{Value: "OFFSET"}
				res.Length.ASTCreator = &ast.StringLiteral{Value: "LENGTHA"}
				res.AddDependency(args[0])
				res.AddLengthDependency(args[0])
				return &res
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
					res := &object.Integer{Value: int64(arg.Value)}
					res.AddDependency(arg)
					return res
				case *object.Boolean:
					if arg == object.TRUE {
						res := &object.Integer{Value: 1}
						res.AddDependency(arg)
						return res
					} else {
						res := &object.Integer{Value: 0}
						res.AddDependency(arg)
						return res
					}
				case *object.String:
					i, err := strconv.ParseInt(arg.String().Value, 10, 64)
					if err != nil {
						res := object.NIL.Copy()
						res.AddDependency(arg)
						return res
					} else {
						res := &object.Integer{Value: i}
						res.AddDependency(arg)
						return res
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
					res := &object.Float{Value: float64(arg.Value)}
					res.AddDependency(arg)
					return res
				case *object.Float:
					return arg
				case *object.Boolean:
					if arg == object.TRUE {
						res := &object.Float{Value: 1}
						res.AddDependency(arg)
						return res
					} else {
						res := &object.Float{Value: 0}
						res.AddDependency(arg)
						return res
					}
				case *object.String:
					f, err := strconv.ParseFloat(arg.String().Value, 64)
					if err != nil {
						res := object.NIL.Copy()
						res.AddDependency(arg)
						return res
					} else {
						return &object.Float{Value: f}
					}
				default:
					return newError("can't cast %s to a float", arg.Type())
				}
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
				return object.CreateArray(elements)
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
				return object.CreateArray(elements)
			},
		},
		"read": &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if err := validateNumberOfArgs(1, args); err != object.NIL {
					return err
				}

				if args[0].Type() != object.STRING_OBJ {
					return newError("argument to `read` must be STRING, got %s", args[0].Type())
				}

				fileLocation := args[0].(*object.String).Value
				data, err := ioutil.ReadFile(fileLocation)

				if err != nil {
					return newError("File reading error %v", fileLocation)
				}
				res := &object.String{Value: string(data)}
				res.AddDependency(args[0])
				return res
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
				res := &object.Integer{Value: int64(rand.Intn(int(arg.Value)))}
				res.AddDependency(arg)
				return res
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
