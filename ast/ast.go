package ast

import (
	"bytes"
	"koko/token"
	"strings"
)

var EMPTY_SPAN = Span{empty: true}

type Span struct {
	empty     bool
	BeginLine int
	BeginPos  int
}

func spanFromToken(t token.Token) Span {
	return Span{BeginLine: t.Context.LineNumber}
}

func (self Span) merge(other Span) Span {
	out := Span{BeginLine: self.BeginLine, BeginPos: self.BeginPos}
	if self.empty {
		out = Span{BeginLine: other.BeginLine, BeginPos: other.BeginPos, empty: other.empty}
		return out
	} else if other.empty {
		out = Span{BeginLine: self.BeginLine, BeginPos: self.BeginPos, empty: self.empty}
		return out
	}
	if other.BeginLine < self.BeginLine || (other.BeginLine == self.BeginLine && other.BeginPos < self.BeginPos) {
		out.BeginLine = other.BeginLine
		out.BeginPos = other.BeginPos
	}
	return out
}

// The base Node interface
type Node interface {
	TokenLiteral() string
	String() string
	Span() Span
}

// All statement nodes implement this
type Statement interface {
	Node
	statementNode()
}

// All expression nodes implement this
type Expression interface {
	Node
	expressionNode()
}

type BuiltinValue struct {
}

func (b *BuiltinValue) TokenLiteral() string {
	return ""
}

func (b *BuiltinValue) String() string {
	return ""
}

func (b *BuiltinValue) Span() Span {
	return EMPTY_SPAN
}

type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

func (p *Program) String() string {
	var out bytes.Buffer

	for _, s := range p.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

func (p *Program) Span() Span {
	out := EMPTY_SPAN

	for _, s := range p.Statements {
		out = out.merge(s.Span())
	}

	return out
}

// Statements
type LetStatement struct {
	Token token.Token // the token.LET token
	Name  *Identifier
	Value Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")

	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}

	out.WriteString(";")

	return out.String()
}

func (ls *LetStatement) Span() Span {
	out := spanFromToken(ls.Token)
	if ls.Value != nil {
		out = out.merge(ls.Value.Span())
	}
	return out
}

type ReturnStatement struct {
	Token       token.Token // the 'return' token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.TokenLiteral() + " ")

	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}

	out.WriteString(";")

	return out.String()
}

func (rs *ReturnStatement) Span() Span {
	out := spanFromToken(rs.Token)
	if rs.ReturnValue != nil {
		out = out.merge(rs.ReturnValue.Span())
	}
	return out
}

type ExpressionStatement struct {
	Token      token.Token // the first token of the expression
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

func (es *ExpressionStatement) Span() Span {
	out := spanFromToken(es.Token)
	if es.Expression != nil {
		out = out.merge(es.Expression.Span())
	}
	return out
}

type BlockStatement struct {
	Token      token.Token // the { token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer

	out.WriteString(token.LBRACE + " ")
	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}
	out.WriteString(" " + token.RBRACE)

	return out.String()
}

func (bs *BlockStatement) Span() Span {
	out := spanFromToken(bs.Token)
	for _, s := range bs.Statements {
		out = out.merge(s.Span())
	}
	return out
}

type ImportStatement struct {
	Token token.Token // the IMPORT token
	Value string
}

func (is *ImportStatement) statementNode()       {}
func (is *ImportStatement) TokenLiteral() string { return is.Token.Literal }
func (is *ImportStatement) String() string {
	return is.Value
}
func (is *ImportStatement) Span() Span {
	return spanFromToken(is.Token)
}

// Expressions
type Identifier struct {
	Token token.Token // the token.IDENT token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }
func (i *Identifier) Span() Span {
	return spanFromToken(i.Token)
}

type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) expressionNode()      {}
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }
func (b *Boolean) String() string       { return b.Token.Literal }
func (b *Boolean) Span() Span {
	return spanFromToken(b.Token)
}

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }
func (il *IntegerLiteral) Span() Span {
	return spanFromToken(il.Token)
}

type CommentLiteral struct {
	Token token.Token
	Value string
}

func (com *CommentLiteral) expressionNode()      {}
func (com *CommentLiteral) TokenLiteral() string { return com.Token.Literal }
func (com *CommentLiteral) String() string {
	return "//" + com.Token.Literal
}
func (com *CommentLiteral) Span() Span {
	return spanFromToken(com.Token)
}

type StringLiteral struct {
	Token token.Token
	Value string
}

func (str *StringLiteral) expressionNode()      {}
func (str *StringLiteral) TokenLiteral() string { return str.Token.Literal }
func (str *StringLiteral) String() string {
	return "\"" + str.Token.Literal + "\""
}
func (str *StringLiteral) Span() Span {
	return spanFromToken(str.Token)
}

type FloatLiteral struct {
	Token token.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }
func (fl *FloatLiteral) Span() Span {
	return spanFromToken(fl.Token)
}

type PrefixExpression struct {
	Token    token.Token // The prefix token, e.g. !
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")

	return out.String()
}
func (pe *PrefixExpression) Span() Span {
	out := spanFromToken(pe.Token)
	out = out.merge(pe.Right.Span())
	return out
}

type InfixExpression struct {
	Token    token.Token // The operator token, e.g. +
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")

	return out.String()
}

func (ie *InfixExpression) Span() Span {
	out := spanFromToken(ie.Token)
	out = out.merge(ie.Left.Span())
	out = out.merge(ie.Right.Span())
	return out
}

type IfExpression struct {
	Token       token.Token // The 'if' token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IfExpression) String() string {
	var out bytes.Buffer

	out.WriteString("if ")
	out.WriteString(ie.Condition.String())
	out.WriteString(" { ")
	out.WriteString(ie.Consequence.String())
	out.WriteString(" } ")

	if ie.Alternative != nil {
		out.WriteString("else { ")
		out.WriteString(ie.Alternative.String())
		out.WriteString(" }")
	}

	return out.String()
}

func (ie *IfExpression) Span() Span {
	out := spanFromToken(ie.Token)
	out = out.merge(ie.Condition.Span())
	out = out.merge(ie.Consequence.Span())
	if ie.Alternative != nil {
		out = out.merge(ie.Alternative.Span())
	}
	return out
}

type FunctionLiteral struct {
	Token      token.Token // The 'fn' token
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}

	out.WriteString(fl.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(fl.Body.String())

	return out.String()
}

func (fl *FunctionLiteral) Span() Span {
	out := spanFromToken(fl.Token)
	for _, p := range fl.Parameters {
		out = out.merge(p.Span())
	}
	out.merge(fl.Body.Span())
	return out
}

type PureFunctionLiteral struct {
	Token      token.Token // The 'pfn' token
	Parameters []*Identifier
	Body       *BlockStatement
}

func (pfl *PureFunctionLiteral) expressionNode()      {}
func (pfl *PureFunctionLiteral) TokenLiteral() string { return pfl.Token.Literal }
func (pfl *PureFunctionLiteral) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range pfl.Parameters {
		params = append(params, p.String())
	}

	out.WriteString(pfl.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(pfl.Body.String())

	return out.String()
}

func (pfl *PureFunctionLiteral) Span() Span {
	out := spanFromToken(pfl.Token)
	for _, p := range pfl.Parameters {
		out = out.merge(p.Span())
	}
	out.merge(pfl.Body.Span())
	return out
}

type CallExpression struct {
	Token     token.Token // The '(' token
	Function  Expression  // Identifier or FunctionLiteral
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer

	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}

	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
}

func (ce *CallExpression) Span() Span {
	out := spanFromToken(ce.Token)
	for _, a := range ce.Arguments {
		out = out.merge(a.Span())
	}
	return out
}

type ArrayLiteral struct {
	Token    token.Token
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }
func (al *ArrayLiteral) String() string {
	var out bytes.Buffer

	elements := []string{}
	for _, el := range al.Elements {
		elements = append(elements, el.String())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

func (al *ArrayLiteral) Span() Span {
	out := spanFromToken(al.Token)
	for _, el := range al.Elements {
		out = out.merge(el.Span())
	}
	return out
}

type IndexExpression struct {
	Token token.Token
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("])")
	return out.String()
}

func (ie *IndexExpression) Span() Span {
	out := spanFromToken(ie.Token)
	out = out.merge(ie.Left.Span())
	out = out.merge(ie.Index.Span())
	return out
}

type HashLiteral struct {
	Token token.Token // { token
	Pairs map[Expression]Expression
}

func (hl *HashLiteral) expressionNode()      {}
func (hl *HashLiteral) TokenLiteral() string { return hl.Token.Literal }
func (hl *HashLiteral) String() string {
	var out bytes.Buffer
	pairs := []string{}
	for key, value := range hl.Pairs {
		pairs = append(pairs, key.String()+":"+value.String())
	}
	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}

func (hl *HashLiteral) Span() Span {
	out := spanFromToken(hl.Token)
	for key, value := range hl.Pairs {
		out = out.merge(key.Span())
		out = out.merge(value.Span())
	}
	return out
}
