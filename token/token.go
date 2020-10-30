package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
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
)

// Jem: Would be cool to make this default lookup the token type in all caps??
var keywords = map[string]TokenType{
	"fn":     FUNCTION,
	"pfn":    PURE_FUNCTION,
	"let":    LET,
	"true":   TRUE,
	"false":  FALSE,
	"if":     IF,
	"else":   ELSE,
	"elsif":  ELSIF,
	"return": RETURN,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
