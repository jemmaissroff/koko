package object

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"koko/ast"
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
	HASH_OBJ     = "HASH"
)

var (
	NIL = &Nil{}

	TRUE  = &Boolean{Value: true}
	FALSE = &Boolean{Value: false}

	EMPTY_STRING = &String{Value: ""}
	ZERO_INTEGER = &Integer{Value: 0}
	ZERO_FLOAT   = &Float{Value: 0}
	EMPTY_ARRAY  = &Array{Elements: []Object{}}
	EMPTY_HASH   = &Hash{Pairs: make(map[HashKey]HashPair)}
)

type Object interface {
	Type() ObjectType
	Inspect() string
	String() String
	Equal(o Object) bool
	Falsey() Object
}

func Bool(o Object) bool { return !o.Equal(o.Falsey()) }

type Integer struct {
	Value int64
}

func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }
func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) String() String   { return String{Value: i.Inspect()} }
func (i *Integer) Float() Float     { return Float{Value: float64(i.Value)} }
func (i *Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: float64(i.Value)}
}
func (i *Integer) Equal(o Object) bool {
	comp, ok := o.(*Integer)
	return ok && comp.Value == i.Value
}
func (i *Integer) Falsey() Object { return ZERO_INTEGER }

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
func (f *Float) HashKey() HashKey {
	return HashKey{Type: f.Type(), Value: f.Value}
}
func (f *Float) Equal(o Object) bool {
	comp, ok := o.(*Float)
	return ok && comp.Value == f.Value
}
func (f *Float) Falsey() Object { return ZERO_FLOAT }

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) String() String   { return String{Value: b.Inspect()} }
func (b *Boolean) HashKey() HashKey {
	var value float64
	if b.Value {
		value = 1
	} else {
		value = 0
	}
	return HashKey{Type: b.Type(), Value: value}
}
func (b *Boolean) Equal(o Object) bool {
	comp, ok := o.(*Boolean)
	return ok && comp.Value == b.Value
}
func (b *Boolean) Falsey() Object { return FALSE }

type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return s.Value }
func (s *String) String() String   { return *s }
func (s *String) Equal(o Object) bool {
	comp, ok := o.(*String)
	return ok && comp.Value == s.Value
}
func (s *String) Falsey() Object { return EMPTY_STRING }

// JEM: Could cache these values to optimize for performance
func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))
	return HashKey{Type: s.Type(), Value: float64(h.Sum64())}
}

type Return struct {
	Value Object
}

func (r *Return) Type() ObjectType { return RETURN_OBJ }
func (r *Return) Inspect() string  { return fmt.Sprintf("%v", r.Value.Inspect()) }
func (r *Return) String() String   { return String{Value: r.Inspect()} }
func (r *Return) Equal(o Object) bool {
	comp, ok := o.(*Return)
	return ok && comp.Value == r.Value
}
func (r *Return) Falsey() Object { return NIL }

type Nil struct{}

func (n *Nil) Type() ObjectType { return NIL_OBJ }
func (n *Nil) Inspect() string  { return "nil" }
func (n *Nil) String() String   { return String{Value: n.Inspect()} }
func (n *Nil) Equal(o Object) bool {
	_, ok := o.(*Nil)
	return ok
}
func (n *Nil) Falsey() Object { return NIL }

type Error struct {
	Message string
}

// JEM: In order to print helpful error messages, need to add line and context
// data to the tokens in lexing. Maybe think about this as an extension??
// He references this vaguely in the book on page 131
func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return "ERROR: " + e.Message }
func (e *Error) String() String   { return String{Value: e.Inspect()} }
func (r *Error) Equal(o Object) bool {
	comp, ok := o.(*Error)
	return ok && comp.Message == r.Message
}
func (e *Error) Falsey() Object { return NIL }

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

// JEM: Could properly implement function comparison
func (f *Function) Equal(o Object) bool {
	_, ok := o.(*Function)
	return ok && false
}
func (f *Function) Falsey() Object { return NIL }

type BuiltinFunction func(args ...Object) Object

type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string  { return "builtin function" }
func (b *Builtin) String() String   { return String{Value: b.Inspect()} }

// JEM: Could properly implement builtin comparison
func (b *Builtin) Equal(o Object) bool {
	_, ok := o.(*Builtin)
	return ok && false
}
func (b *Builtin) Falsey() Object { return NIL }

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
func (a *Array) Equal(o Object) bool {
	comp, ok := o.(*Array)
	if !ok || len(comp.Elements) != len(a.Elements) {
		return false
	}
	for i, el := range a.Elements {
		if !el.Equal(comp.Elements[i]) {
			return false
		}
	}
	return true
}
func (a *Array) Falsey() Object { return EMPTY_ARRAY }

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
	return obj, ok
}
func (f *PureFunction) Falsey() Object { return NIL }

func (f *PureFunction) Set(args []Object, val Object) Object {
	f.Cache[objectsToString(args)] = val
	return val
}

// JEM: Could properly implement function comparison
func (f *PureFunction) Equal(o Object) bool {
	_, ok := o.(*PureFunction)
	return ok && false
}

func objectsToString(args []Object) string {
	var res string
	for _, arg := range args {
		res += "@" + arg.String().Value
	}
	return res
}

type HashPair struct {
	Key   Object
	Value Object
}

type Hash struct {
	Pairs map[HashKey]HashPair
}

func (h *Hash) Type() ObjectType { return HASH_OBJ }

func (h *Hash) Inspect() string {
	var out bytes.Buffer
	pairs := []string{}
	for _, pair := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s", pair.Key.Inspect(), pair.Value.Inspect()))
	}
	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}
func (h *Hash) String() String { return String{Value: h.Inspect()} }
func (h *Hash) Equal(o Object) bool {
	comp, ok := o.(*Hash)
	if !ok || len(h.Pairs) != len(comp.Pairs) {
		return false
	}

	for k, v := range h.Pairs {
		if !(v.Value.Equal(comp.Pairs[k].Value) &&
			v.Key.Equal(comp.Pairs[k].Key)) {
			return false
		}
	}
	return true
}
func (h *Hash) Falsey() Object { return EMPTY_HASH }

type HashKey struct {
	Type  ObjectType
	Value float64
}

type Hashable interface {
	HashKey() HashKey
}
