package object

import (
	"bytes"
	"fmt"
	"monkey/ast"
	"strconv"
	"strings"
)

type ObjectType string

const (
	BOOLEAN_OBJ  = "BOOLEAN"
	FLOAT_OBJ    = "FLOAT"
	INTEGER_OBJ  = "INTEGER"
	NIL_OBJ      = "NIL"
	RETURN_OBJ   = "RETURN"
	STRING_OBJ   = "STRING"
	ERROR_OBJ    = "ERROR"
	FUNCTION_OBJ = "FUNCTION"
	BUILTIN_OBJ  = "BUILTIN"
	ARRAY_OBJ    = "ARRAY"
)

type Object interface {
	Type() ObjectType
	Inspect() string
	String() String
}

type Integer struct {
	Value int64
}

func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }
func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) String() String   { return String{Value: i.Inspect()} }
func (i *Integer) Float() Float     { return Float{Value: float64(i.Value)} }

type Float struct {
	Value float64
}

func (f *Float) Inspect() string {
	if f.Value == float64(int64(f.Value)) {
		return fmt.Sprintf("%.1f", f.Value)
	}
	return strconv.FormatFloat(f.Value, 'f', -1, 64)
}
func (f *Float) Type() ObjectType { return FLOAT_OBJ }
func (f *Float) String() String   { return String{Value: f.Inspect()} }

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) String() String   { return String{Value: b.Inspect()} }

type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return s.Value }
func (s *String) String() String   { return *s }

type Return struct {
	Value Object
}

func (r *Return) Type() ObjectType { return RETURN_OBJ }
func (r *Return) Inspect() string  { return fmt.Sprintf("%v", r.Value.Inspect()) }
func (r *Return) String() String   { return String{Value: r.Inspect()} }

type Nil struct{}

func (n *Nil) Type() ObjectType { return NIL_OBJ }
func (n *Nil) Inspect() string  { return "nil" }
func (n *Nil) String() String   { return String{Value: n.Inspect()} }

type Error struct {
	Message string
}

// JEM: In order to print helpful error messages, need to add line and context
// data to the tokens in lexing. Maybe think about this as an extension??
// He references this vaguely in the book on page 131
func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return "ERROR: " + e.Message }
func (e *Error) String() String   { return String{Value: e.Inspect()} }

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) Inspect() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}

	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")

	return out.String()
}
func (f *Function) String() String { return String{Value: f.Inspect()} }

type BuiltinFunction func(args ...Object) Object

type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string  { return "builtin function" }
func (b *Builtin) String() String   { return String{Value: b.Inspect()} }

type Array struct {
	Elements []Object
}

func (a *Array) Type() ObjectType { return ARRAY_OBJ }
func (a *Array) Inspect() string {

	var out bytes.Buffer
	elements := []string{}
	for _, e := range a.Elements {
		elements = append(elements, e.Inspect())
	}
	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")
	return out.String()
}
func (a *Array) String() String { return String{Value: a.Inspect()} }

type PureFunction struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
	Cache      map[string]Object
}

func NewPureFunction(parameters []*ast.Identifier, env *Environment, body *ast.BlockStatement) *PureFunction {
	cache := make(map[string]Object)
	return &PureFunction{Parameters: parameters, Body: body, Env: env, Cache: cache}
}

func (f *PureFunction) Type() ObjectType { return FUNCTION_OBJ }
func (f *PureFunction) Inspect() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}

	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")

	return out.String()
}
func (f *PureFunction) String() String { return String{Value: f.Inspect()} }
func (f *PureFunction) Get(args []Object) (Object, bool) {
	obj, ok := f.Cache[objectsToString(args)]
	//println("hello", objectsToString(args), ok)
	return obj, ok
}

func (f *PureFunction) Set(args []Object, val Object) Object {
	f.Cache[objectsToString(args)] = val
	return val
}

func objectsToString(args []Object) string {
	var res string
	for _, arg := range args {
		res += "@" + arg.String().Value
	}
	return res
}
