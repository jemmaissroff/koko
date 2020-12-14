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
	BOOLEAN_OBJ              = "BOOLEAN"
	FLOAT_OBJ                = "FLOAT"
	INTEGER_OBJ              = "INTEGER"
	NIL_OBJ                  = "NIL"
	RETURN_OBJ               = "RETURN"
	STRING_OBJ               = "STRING"
	ERROR_OBJ                = "ERROR"
	FUNCTION_OBJ             = "FUNCTION"
	BUILTIN_OBJ              = "BUILTIN"
	ARRAY_OBJ                = "ARRAY"
	TRACE_OBJ                = "TRACE"
	HASH_OBJ                 = "HASH"
	DEBUG_TRACE_METADATA_OBJ = "DEBUG_TRACE_METADATA"
)

var (
	NIL = &Nil{ASTCreator: &ast.BuiltinValue{}}

	TRUE  = &Boolean{Value: true, ASTCreator: &ast.BuiltinValue{}}
	FALSE = &Boolean{Value: false, ASTCreator: &ast.BuiltinValue{}}

	EMPTY_STRING = &String{Value: "", ASTCreator: &ast.BuiltinValue{}}
	ZERO_INTEGER = &Integer{Value: 0, ASTCreator: &ast.BuiltinValue{}}
	ZERO_FLOAT   = &Float{Value: 0, ASTCreator: &ast.BuiltinValue{}}
	EMPTY_ARRAY  = &Array{Elements: []Object{}, ASTCreator: &ast.BuiltinValue{}, Length: Integer{ASTCreator: &ast.BuiltinValue{}}, Offset: Integer{ASTCreator: &ast.BuiltinValue{}}}
	EMPTY_HASH   = &Hash{Pairs: make(map[HashKey]HashPair), ASTCreator: &ast.BuiltinValue{}, Length: Integer{ASTCreator: &ast.BuiltinValue{}}, Offset: Integer{ASTCreator: &ast.BuiltinValue{}}}
)

func copyDependencies(deps []Object) []Object {
	res := make([]Object, len(deps))
	copy(res, deps)
	return res
}

func GetAllDependencies(result Object) map[Object]bool {
	out := make(map[Object]bool)
	queue := []Object{}
	queue = append(queue, result)
	for len(queue) > 0 {
		head := queue[0]
		if out[head] {
			if len(queue) > 1 {
				queue = queue[1:]
			} else {
				queue = []Object{}
			}
			continue
		}
		out[head] = true
		if len(queue) > 1 {
			queue = queue[1:]
			for link := range head.GetDependencyLinks() {
				queue = append(queue, link)
			}
		} else {
			for link := range head.GetDependencyLinks() {
				queue = append(queue, link)
			}
		}
	}
	return out
}

type Object interface {
	Type() ObjectType
	Inspect() string
	String() String
	Copy() Object
	CopyWithoutDependency() Object
	Equal(o Object) bool
	Falsey() Object
	AddDependency(dep Object)
	GetDependencyLinks() map[Object]bool
	GetCreatorNode() ast.Node
	SetCreatorNode(node ast.Node)
}

func Bool(o Object) bool { return !o.Equal(o.Falsey()) }

type Integer struct {
	Value        int64
	Dependencies map[Object]bool
	ASTCreator   ast.Node
}

func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }
func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) String() String   { return String{Value: i.Inspect()} }
func (i *Integer) Float() Float     { return Float{Value: float64(i.Value)} }
func (i *Integer) Copy() Object {
	return &Integer{Value: i.Value, Dependencies: map[Object]bool{i: true}, ASTCreator: i.ASTCreator}
}
func (i *Integer) CopyWithoutDependency() Object {
	return &Integer{Value: i.Value, ASTCreator: i.ASTCreator}
}
func (i *Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: float64(i.Value)}
}
func (i *Integer) Equal(o Object) bool {
	comp, ok := o.(*Integer)
	return ok && comp.Value == i.Value
}
func (i *Integer) Falsey() Object { return ZERO_INTEGER.Copy() }
func (i *Integer) AddDependency(dep Object) {
	if i.Dependencies == nil {
		i.Dependencies = make(map[Object]bool)
	}
	i.Dependencies[dep] = true
}
func (i *Integer) GetDependencyLinks() map[Object]bool { return i.Dependencies }
func (i *Integer) GetCreatorNode() ast.Node            { return i.ASTCreator }
func (i *Integer) SetCreatorNode(node ast.Node)        { i.ASTCreator = node }

type Float struct {
	Value        float64
	Dependencies map[Object]bool
	ASTCreator   ast.Node
}

func (f *Float) Inspect() string {
	if f.Value == float64(int64(f.Value)) {
		return fmt.Sprintf("%.1f", f.Value)
	}
	return strconv.FormatFloat(f.Value, 'f', -1, 64)
}
func (f *Float) Type() ObjectType { return FLOAT_OBJ }
func (f *Float) String() String   { return String{Value: f.Inspect()} }
func (f *Float) Copy() Object {
	return &Float{Value: f.Value, Dependencies: map[Object]bool{f: true}, ASTCreator: f.ASTCreator}
}
func (f *Float) CopyWithoutDependency() Object {
	return &Float{Value: f.Value, ASTCreator: f.ASTCreator}
}
func (f *Float) HashKey() HashKey {
	return HashKey{Type: f.Type(), Value: f.Value}
}
func (f *Float) Equal(o Object) bool {
	comp, ok := o.(*Float)
	return ok && comp.Value == f.Value
}
func (f *Float) Falsey() Object { return ZERO_FLOAT.Copy() }

func (f *Float) AddDependency(dep Object) {
	if f.Dependencies == nil {
		f.Dependencies = make(map[Object]bool)
	}
	f.Dependencies[dep] = true
}
func (f *Float) GetDependencyLinks() map[Object]bool { return f.Dependencies }
func (f *Float) GetCreatorNode() ast.Node            { return f.ASTCreator }
func (f *Float) SetCreatorNode(node ast.Node)        { f.ASTCreator = node }

type Boolean struct {
	Value        bool
	Dependencies map[Object]bool
	ASTCreator   ast.Node
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) String() String   { return String{Value: b.Inspect()} }
func (b *Boolean) Copy() Object {
	return &Boolean{Value: b.Value, Dependencies: map[Object]bool{b: true}, ASTCreator: b.ASTCreator}
}
func (b *Boolean) CopyWithoutDependency() Object {
	return &Boolean{Value: b.Value, ASTCreator: b.ASTCreator}
}
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
func (b *Boolean) Falsey() Object { return FALSE.Copy() }

func (b *Boolean) AddDependency(dep Object) {
	if b.Dependencies == nil {
		b.Dependencies = make(map[Object]bool)
	}
	b.Dependencies[dep] = true
}
func (b *Boolean) GetDependencyLinks() map[Object]bool { return b.Dependencies }
func (b *Boolean) GetCreatorNode() ast.Node            { return b.ASTCreator }
func (b *Boolean) SetCreatorNode(node ast.Node)        { b.ASTCreator = node }

type String struct {
	Value        string
	Dependencies map[Object]bool
	ASTCreator   ast.Node
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return s.Value }
func (s *String) String() String   { return *s }
func (s *String) Copy() Object {
	return &String{Value: s.Value, Dependencies: map[Object]bool{s: true}, ASTCreator: s.ASTCreator}
}
func (s *String) CopyWithoutDependency() Object {
	return &String{Value: s.Value, ASTCreator: s.ASTCreator}
}
func (s *String) Equal(o Object) bool {
	comp, ok := o.(*String)
	return ok && comp.Value == s.Value
}
func (s *String) Falsey() Object { return EMPTY_STRING.Copy() }

// JEM: Could cache these values to optimize for performance
func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))
	return HashKey{Type: s.Type(), Value: float64(h.Sum64())}
}

func (s *String) AddDependency(dep Object) {
	if s.Dependencies == nil {
		s.Dependencies = make(map[Object]bool)
	}
	s.Dependencies[dep] = true
}
func (s *String) GetDependencyLinks() map[Object]bool { return s.Dependencies }
func (s *String) GetCreatorNode() ast.Node            { return s.ASTCreator }
func (s *String) SetCreatorNode(node ast.Node)        { s.ASTCreator = node }

type Return struct {
	Value        Object
	Dependencies map[Object]bool
	ASTCreator   ast.Node
}

func (r *Return) Type() ObjectType { return RETURN_OBJ }
func (r *Return) Inspect() string  { return fmt.Sprintf("%v", r.Value.Inspect()) }
func (r *Return) String() String   { return String{Value: r.Inspect()} }
func (r *Return) Copy() Object {
	return &Return{Value: r.Value, Dependencies: map[Object]bool{r: true}, ASTCreator: r.ASTCreator}
}
func (r *Return) CopyWithoutDependency() Object {
	return &Return{Value: r.Value, ASTCreator: r.ASTCreator}
}
func (r *Return) Equal(o Object) bool {
	comp, ok := o.(*Return)
	return ok && comp.Value == r.Value
}
func (r *Return) Falsey() Object { return NIL.Copy() }

func (r *Return) AddDependency(dep Object) {
	if r.Dependencies == nil {
		r.Dependencies = make(map[Object]bool)
	}
	r.Dependencies[dep] = true
}
func (r *Return) GetDependencyLinks() map[Object]bool { return r.Dependencies }
func (r *Return) GetCreatorNode() ast.Node            { return r.ASTCreator }
func (r *Return) SetCreatorNode(node ast.Node)        { r.ASTCreator = node }

type Nil struct {
	Dependencies map[Object]bool
	ASTCreator   ast.Node
}

func (n *Nil) Type() ObjectType { return NIL_OBJ }
func (n *Nil) Inspect() string  { return "nil" }
func (n *Nil) String() String   { return String{Value: n.Inspect()} }
func (n *Nil) Copy() Object {
	return &Nil{Dependencies: map[Object]bool{n: true}, ASTCreator: n.ASTCreator}
}
func (n *Nil) CopyWithoutDependency() Object {
	return &Nil{ASTCreator: n.ASTCreator}
}
func (n *Nil) Equal(o Object) bool {
	_, ok := o.(*Nil)
	return ok
}
func (n *Nil) Falsey() Object { return NIL.Copy() }

func (n *Nil) AddDependency(dep Object) {
	if n.Dependencies == nil {
		n.Dependencies = make(map[Object]bool)
	}
	n.Dependencies[dep] = true
}
func (n *Nil) GetDependencyLinks() map[Object]bool { return n.Dependencies }
func (n *Nil) GetCreatorNode() ast.Node            { return n.ASTCreator }
func (n *Nil) SetCreatorNode(node ast.Node)        { n.ASTCreator = node }

type Error struct {
	Message      string
	Dependencies map[Object]bool
	ASTCreator   ast.Node
}

// JEM: In order to print helpful error messages, need to add line and context
// data to the tokens in lexing. Maybe think about this as an extension??
// He references this vaguely in the book on page 131
func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return "ERROR: " + e.Message }
func (e *Error) String() String   { return String{Value: e.Inspect()} }
func (e *Error) Copy() Object {
	return &Error{Message: e.Message, Dependencies: map[Object]bool{e: true}, ASTCreator: e.ASTCreator}
}
func (e *Error) CopyWithoutDependency() Object {
	return &Error{Message: e.Message, ASTCreator: e.ASTCreator}
}
func (e *Error) Equal(o Object) bool {
	comp, ok := o.(*Error)
	return ok && comp.Message == e.Message
}
func (e *Error) Falsey() Object { return NIL.Copy() }

func (e *Error) AddDependency(dep Object) {
	if e.Dependencies == nil {
		e.Dependencies = make(map[Object]bool)
	}
	e.Dependencies[dep] = true
}
func (e *Error) GetDependencyLinks() map[Object]bool { return e.Dependencies }
func (e *Error) GetCreatorNode() ast.Node            { return e.ASTCreator }
func (e *Error) SetCreatorNode(node ast.Node)        { e.ASTCreator = node }

type Function struct {
	Parameters   []*ast.Identifier
	Body         *ast.BlockStatement
	Env          *Environment
	Dependencies map[Object]bool
	ASTCreator   ast.Node
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
func (f *Function) Copy() Object {
	return &Function{Parameters: f.Parameters, Body: f.Body, Env: f.Env, Dependencies: map[Object]bool{f: true}, ASTCreator: f.ASTCreator}
}
func (f *Function) CopyWithoutDependency() Object {
	return &Function{Parameters: f.Parameters, Body: f.Body, Env: f.Env, ASTCreator: f.ASTCreator}
}

// JEM: Could properly implement function comparison
func (f *Function) Equal(o Object) bool {
	_, ok := o.(*Function)
	return ok && false
}
func (f *Function) Falsey() Object { return NIL.Copy() }

func (f *Function) AddDependency(dep Object) {
	if f.Dependencies == nil {
		f.Dependencies = make(map[Object]bool)
	}
	f.Dependencies[dep] = true
}
func (f *Function) GetDependencyLinks() map[Object]bool { return f.Dependencies }
func (f *Function) GetCreatorNode() ast.Node            { return f.ASTCreator }
func (f *Function) SetCreatorNode(node ast.Node)        { f.ASTCreator = node }

type BuiltinFunction func(args ...Object) Object

type Builtin struct {
	Fn           BuiltinFunction
	Dependencies map[Object]bool
	ASTCreator   ast.Node
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string  { return "builtin function" }
func (b *Builtin) String() String   { return String{Value: b.Inspect()} }
func (b *Builtin) Copy() Object {
	return &Builtin{Fn: b.Fn, Dependencies: map[Object]bool{b: true}, ASTCreator: b.ASTCreator}
}
func (b *Builtin) CopyWithoutDependency() Object {
	return &Builtin{Fn: b.Fn, ASTCreator: b.ASTCreator}
}

// JEM: Could properly implement builtin comparison
func (b *Builtin) Equal(o Object) bool {
	_, ok := o.(*Builtin)
	return ok && false
}
func (b *Builtin) Falsey() Object { return NIL.Copy() }

func (b *Builtin) AddDependency(dep Object) {
	if b.Dependencies == nil {
		b.Dependencies = make(map[Object]bool)
	}
	b.Dependencies[dep] = true
}
func (b *Builtin) GetDependencyLinks() map[Object]bool { return b.Dependencies }
func (b *Builtin) GetCreatorNode() ast.Node            { return b.ASTCreator }
func (b *Builtin) SetCreatorNode(node ast.Node)        { b.ASTCreator = node }

type Array struct {
	Elements     []Object
	Dependencies map[Object]bool
	Length       Integer
	Offset       Integer
	ASTCreator   ast.Node
}

func CreateArray(elements []Object) *Array {
	res := Array{Elements: elements, Length: Integer{Value: int64(len(elements))}}
	res.Offset.ASTCreator = &ast.StringLiteral{Value: "OFFSET"}
	res.Length.ASTCreator = &ast.StringLiteral{Value: "LENGTHA"}
	for _, e := range elements {
		res.AddDependency(e)
	}
	res.AddDependency(&res.Length)
	return &res
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
func (a *Array) Falsey() Object { return EMPTY_ARRAY.Copy() }
func (a *Array) Copy() Object {
	return &Array{Elements: a.Elements, Dependencies: map[Object]bool{a: true}, Length: *a.Length.Copy().(*Integer), Offset: *a.Offset.Copy().(*Integer), ASTCreator: a.ASTCreator}
}
func (a *Array) CopyWithoutDependency() Object {
	return &Array{Elements: a.Elements, Length: *a.Length.Copy().(*Integer), Offset: *a.Offset.Copy().(*Integer), ASTCreator: a.ASTCreator}
}

func (a *Array) AddDependency(dep Object) {
	if a.Dependencies == nil {
		a.Dependencies = make(map[Object]bool)
	}
	a.Dependencies[dep] = true
}
func (a *Array) AddLengthDependency(dep Object) {
	a.Length.AddDependency(dep)
}
func (a *Array) AddOffsetDependency(dep Object) {
	a.Offset.AddDependency(dep)
}

func (a *Array) GetDependencyLinks() map[Object]bool {
	out := make(map[Object]bool)
	for k, v := range a.Dependencies {
		out[k] = v
	}
	out[&a.Length] = true
	return out
}

func (a *Array) GetCreatorNode() ast.Node { return a.ASTCreator }
func (a *Array) SetCreatorNode(node ast.Node) {
	a.ASTCreator = node
	a.Length.ASTCreator = &ast.LengthNode{Child: node}
	a.Offset.ASTCreator = &ast.OffsetNode{Child: node}
}

type PureFunction struct {
	Parameters   []*ast.Identifier
	Body         *ast.BlockStatement
	Env          *Environment
	Cache        map[string]Object
	Dependencies map[Object]bool
	ASTCreator   ast.Node
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
func (f *PureFunction) Falsey() Object { return NIL.Copy() }

func (f *PureFunction) Set(args []Object, val Object) Object {
	f.Cache[objectsToString(args)] = val
	return val
}
func (f *PureFunction) Copy() Object {
	return &PureFunction{Parameters: f.Parameters, Body: f.Body, Cache: f.Cache, Env: f.Env, Dependencies: map[Object]bool{f: true}, ASTCreator: f.ASTCreator}
}

func (f *PureFunction) CopyWithoutDependency() Object {
	return &PureFunction{Parameters: f.Parameters, Body: f.Body, Cache: f.Cache, Env: f.Env, ASTCreator: f.ASTCreator}
}

func (f *PureFunction) GetCreatorNode() ast.Node     { return f.ASTCreator }
func (f *PureFunction) SetCreatorNode(node ast.Node) { f.ASTCreator = node }

func (f *PureFunction) AddDependency(dep Object) {
	if f.Dependencies == nil {
		f.Dependencies = make(map[Object]bool)
	}
	f.Dependencies[dep] = true
}

// JEM: Could properly implement function comparison
func (f *PureFunction) Equal(o Object) bool {
	_, ok := o.(*PureFunction)
	return ok && false
}

func (f *PureFunction) GetDependencyLinks() map[Object]bool { return f.Dependencies }

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
	Pairs        map[HashKey]HashPair
	Length       Integer
	Offset       Integer
	Dependencies map[Object]bool
	ASTCreator   ast.Node
}

func CreateHash(pairs map[HashKey]HashPair) *Hash {
	res := Hash{Pairs: pairs, Length: Integer{Value: int64(len(pairs))}}
	for _, v := range res.Pairs {
		res.AddDependency(v.Key)
		res.AddDependency(v.Value)
	}
	res.AddDependency(&res.Length)
	return &res
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
func (h *Hash) Falsey() Object { return EMPTY_HASH.Copy() }

func (h *Hash) Copy() Object {
	return &Hash{Pairs: h.Pairs, Length: *h.Length.Copy().(*Integer), Offset: *h.Offset.Copy().(*Integer), Dependencies: map[Object]bool{h: true}, ASTCreator: h.ASTCreator}
}

func (h *Hash) CopyWithoutDependency() Object {
	return &Hash{Pairs: h.Pairs, Length: *h.Length.Copy().(*Integer), Offset: *h.Offset.Copy().(*Integer), ASTCreator: h.ASTCreator}
}

func (h *Hash) AddDependency(dep Object) {
	if h.Dependencies == nil {
		h.Dependencies = make(map[Object]bool)
	}
	h.Dependencies[dep] = true
}
func (h *Hash) AddLengthDependency(dep Object) { h.Length.AddDependency(dep) }
func (h *Hash) AddOffsetDependency(dep Object) { h.Offset.AddDependency(dep) }

func (h *Hash) GetDependencyLinks() map[Object]bool {
	out := make(map[Object]bool)
	for k, v := range h.Dependencies {
		out[k] = v
	}
	out[&h.Length] = true
	return out
}

func (h *Hash) GetCreatorNode() ast.Node     { return h.ASTCreator }
func (h *Hash) SetCreatorNode(node ast.Node) { h.ASTCreator = node }

type HashKey struct {
	Type  ObjectType
	Value float64
}

type Hashable interface {
	HashKey() HashKey
}

type DebugTraceMetadata struct {
	DebugMetadata map[string]bool
	Dependencies  map[Object]bool
	ASTCreator    ast.Node
}

// NOTE this object is techincal debt
// TODO (Peter) remove this gracefully and replace with a better version
func (d *DebugTraceMetadata) Type() ObjectType { return DEBUG_TRACE_METADATA_OBJ }
func (d *DebugTraceMetadata) Inspect() string  { return fmt.Sprintf("%+v\n", d.DebugMetadata) }
func (d *DebugTraceMetadata) String() String   { return String{Value: d.Inspect()} }
func (d *DebugTraceMetadata) Copy() Object {
	return &DebugTraceMetadata{DebugMetadata: d.DebugMetadata, Dependencies: map[Object]bool{d: true}, ASTCreator: d.ASTCreator}
}
func (d *DebugTraceMetadata) CopyWithoutDependency() Object {
	return &DebugTraceMetadata{DebugMetadata: d.DebugMetadata, ASTCreator: d.ASTCreator}
}

// JEM: Could properly implement builtin comparison
func (d *DebugTraceMetadata) Equal(o Object) bool {
	_, ok := o.(*Builtin)
	return ok && false
}
func (d *DebugTraceMetadata) Falsey() Object { return NIL.Copy() }

func (d *DebugTraceMetadata) AddDependency(dep Object) {
	if d.Dependencies == nil {
		d.Dependencies = make(map[Object]bool)
	}
	d.Dependencies[dep] = true
}
func (d *DebugTraceMetadata) GetDependencyLinks() map[Object]bool { return d.Dependencies }

func (d *DebugTraceMetadata) GetCreatorNode() ast.Node     { return d.ASTCreator }
func (d *DebugTraceMetadata) SetCreatorNode(node ast.Node) { d.ASTCreator = node }
