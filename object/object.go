package object

import (
	"fmt"
	"strconv"
)

type ObjectType string

const (
	BOOLEAN_OBJ = "BOOLEAN"
	FLOAT_OBJ   = "FLOAT"
	INTEGER_OBJ = "INTEGER"
	NIL_OBJ     = "NIL"
	RETURN_OBJ  = "RETURN"
	STRING_OBJ  = "STRING"
	ERROR_OBJ   = "ERROR"
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
