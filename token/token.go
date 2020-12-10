package token

import (
	"bytes"
	"strconv"
)

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
	Context ContextData
}

type ContextData struct {
	LineNumber int
	File       string
}

func (c ContextData) String() string {
	var out bytes.Buffer

	out.WriteString(c.File + " line " + strconv.Itoa(c.LineNumber))

	return out.String()
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers + literals
	IDENT  = "IDENT" // add, foobar, x, y
	INT    = "INT"   // 123456
	FLOAT  = "FLOAT" // 1.232
	STRING = "STRING"

	// Operators
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	PERCENT  = "%"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"

	// Comparisons
	EQ     = "=="
	NOT_EQ = "!="
	LT     = "<"
	GT     = ">"

	// Delimeters
	COLON     = ":"
	COMMA     = ","
	SEMICOLON = ";"

	LPAREN   = "("
	RPAREN   = ")"
	LBRACE   = "{"
	RBRACE   = "}"
	LBRACKET = "["
	RBRACKET = "]"

	COMMENT = "//"

	// Keywords
	FUNCTION      = "FUNCTION"
	PURE_FUNCTION = "PURE_FUNCTION"
	LET           = "LET"
	TRUE          = "TRUE"
	FALSE         = "FALSE"
	IF            = "IF"
	ELSE          = "ELSE"
	ELSIF         = "ELSIF"
	RETURN        = "RETURN"
	IMPORT        = "IMPORT"
)

// Jem: Would be cool to make this default lookup the token type in all caps??
var keywords = map[string]TokenType{
	"else":   ELSE,
	"elsif":  ELSIF,
	"false":  FALSE,
	"fn":     FUNCTION,
	"if":     IF,
	"import": IMPORT,
	"let":    LET,
	"pfn":    PURE_FUNCTION,
	"return": RETURN,
	"true":   TRUE,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
