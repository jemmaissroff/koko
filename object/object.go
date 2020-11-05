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

type TraceMetadata struct {
	Dependencies map[int]bool
}

func MergeDependencies(a TraceMetadata, b TraceMetadata) TraceMetadata {
	res := TraceMetadata{Dependencies: make(map[int]bool)}
	for k, v := range a.Dependencies {
		if v == true {
			res.Dependencies[k] = true
		}
	}
	for k, v := range b.Dependencies {
		if v == true {
			res.Dependencies[k] = true
		}
	}
	return res
}

type Object interface {
	Type() ObjectType
	Inspect() string
	String() String
	GetMetadata() TraceMetadata
	SetMetadata(metadata TraceMetadata)
	Copy() Object
}

type Integer struct {
	Value    int64
	metadata TraceMetadata
}

func (i *Integer) Inspect() string                    { return fmt.Sprintf("%d", i.Value) }
func (i *Integer) Type() ObjectType                   { return INTEGER_OBJ }
func (i *Integer) String() String                     { return String{Value: i.Inspect()} }
func (i *Integer) Float() Float                       { return Float{Value: float64(i.Value)} }
func (i *Integer) GetMetadata() TraceMetadata         { return i.metadata }
func (i *Integer) SetMetadata(metadata TraceMetadata) { i.metadata = metadata }
func (i *Integer) Copy() Object                       { return &Integer{Value: i.Value, metadata: i.metadata} }

type Float struct {
	Value    float64
	metadata TraceMetadata
}

func (f *Float) Inspect() string {
	if f.Value == float64(int64(f.Value)) {
		return fmt.Sprintf("%.1f", f.Value)
	}
	return strconv.FormatFloat(f.Value, 'f', -1, 64)
}
func (f *Float) Type() ObjectType                   { return FLOAT_OBJ }
func (f *Float) String() String                     { return String{Value: f.Inspect()} }
func (f *Float) GetMetadata() TraceMetadata         { return f.metadata }
func (f *Float) SetMetadata(metadata TraceMetadata) { f.metadata = metadata }
func (f *Float) Copy() Object                       { return &Float{Value: f.Value, metadata: f.metadata} }

type Boolean struct {
	Value    bool
	metadata TraceMetadata
}

func (b *Boolean) Type() ObjectType                   { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string                    { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) String() String                     { return String{Value: b.Inspect()} }
func (b *Boolean) GetMetadata() TraceMetadata         { return b.metadata }
func (b *Boolean) SetMetadata(metadata TraceMetadata) { b.metadata = metadata }
func (b *Boolean) Copy() Object                       { return &Boolean{Value: b.Value, metadata: b.metadata} }

type String struct {
	Value    string
	metadata TraceMetadata
}

func (s *String) Type() ObjectType                   { return STRING_OBJ }
func (s *String) Inspect() string                    { return s.Value }
func (s *String) String() String                     { return *s }
func (s *String) GetMetadata() TraceMetadata         { return s.metadata }
func (s *String) SetMetadata(metadata TraceMetadata) { s.metadata = metadata }
func (s *String) Copy() Object                       { return &String{Value: s.Value, metadata: s.metadata} }

type Return struct {
	Value    Object
	metadata TraceMetadata
}

func (r *Return) Type() ObjectType                   { return RETURN_OBJ }
func (r *Return) Inspect() string                    { return fmt.Sprintf("%v", r.Value.Inspect()) }
func (r *Return) String() String                     { return String{Value: r.Inspect()} }
func (r *Return) GetMetadata() TraceMetadata         { return r.metadata }
func (r *Return) SetMetadata(metadata TraceMetadata) { r.metadata = metadata }
func (r *Return) Copy() Object                       { return &Return{Value: r.Value, metadata: r.metadata} }

type Nil struct {
	metadata TraceMetadata
}

func (n *Nil) Type() ObjectType                   { return NIL_OBJ }
func (n *Nil) Inspect() string                    { return "nil" }
func (n *Nil) String() String                     { return String{Value: n.Inspect()} }
func (n *Nil) GetMetadata() TraceMetadata         { return n.metadata }
func (n *Nil) SetMetadata(metadata TraceMetadata) { n.metadata = metadata }
func (n *Nil) Copy() Object                       { return &Nil{metadata: n.metadata} }

type Error struct {
	Message  string
	metadata TraceMetadata
}

// JEM: In order to print helpful error messages, need to add line and context
// data to the tokens in lexing. Maybe think about this as an extension??
// He references this vaguely in the book on page 131
func (e *Error) Type() ObjectType                   { return ERROR_OBJ }
func (e *Error) Inspect() string                    { return "ERROR: " + e.Message }
func (e *Error) String() String                     { return String{Value: e.Inspect()} }
func (e *Error) GetMetadata() TraceMetadata         { return e.metadata }
func (e *Error) SetMetadata(metadata TraceMetadata) { e.metadata = metadata }
func (e *Error) Copy() Object                       { return &Error{Message: e.Message, metadata: e.metadata} }

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
	metadata   TraceMetadata
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
func (f *Function) String() String                     { return String{Value: f.Inspect()} }
func (f *Function) GetMetadata() TraceMetadata         { return f.metadata }
func (f *Function) SetMetadata(metadata TraceMetadata) { f.metadata = metadata }
func (f *Function) Copy() Object {
	return &Function{Parameters: f.Parameters, Body: f.Body, Env: f.Env, metadata: f.metadata}
}

type BuiltinFunction func(args ...Object) Object

type Builtin struct {
	Fn       BuiltinFunction
	metadata TraceMetadata
}

func (b *Builtin) Type() ObjectType                   { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string                    { return "builtin function" }
func (b *Builtin) String() String                     { return String{Value: b.Inspect()} }
func (b *Builtin) GetMetadata() TraceMetadata         { return b.metadata }
func (b *Builtin) SetMetadata(metadata TraceMetadata) { b.metadata = metadata }
func (b *Builtin) Copy() Object                       { return &Builtin{Fn: b.Fn, metadata: b.metadata} }

type Array struct {
	Elements []Object
	metadata TraceMetadata
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
func (a *Array) String() String                     { return String{Value: a.Inspect()} }
func (a *Array) GetMetadata() TraceMetadata         { return a.metadata }
func (a *Array) SetMetadata(metadata TraceMetadata) { a.metadata = metadata }
func (a *Array) Copy() Object                       { return &Array{Elements: a.Elements, metadata: a.metadata} }

type PureFunction struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
	Cache      map[string]Object
	metadata   TraceMetadata
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
	return obj, ok
}

func (f *PureFunction) Set(args []Object, val Object) Object {
	f.Cache[objectsToString(args)] = val
	return val
}

func (f *PureFunction) GetMetadata() TraceMetadata         { return f.metadata }
func (f *PureFunction) SetMetadata(metadata TraceMetadata) { f.metadata = metadata }
func (f *PureFunction) Copy() Object {
	return &PureFunction{Parameters: f.Parameters, Body: f.Body, Cache: f.Cache, Env: f.Env, metadata: f.metadata}
}

func objectsToString(args []Object) string {
	var res string
	for _, arg := range args {
		res += "@" + arg.String().Value
	}
	return res
}
