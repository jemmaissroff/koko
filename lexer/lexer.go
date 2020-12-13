package lexer

import "koko/token"

type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
	lineNumber   int
	filename     string
}

func New(input string, filename string) *Lexer {
	l := &Lexer{input: input, filename: filename, lineNumber: 1}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}

	l.position = l.readPosition
	l.readPosition += 1
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	switch l.ch {
	case '=':
		tok = twoChar(l, token.ASSIGN, token.EQ, '=')
	case ',':
		tok = newToken(l, token.COMMA, l.ch)
	case '+':
		tok = newToken(l, token.PLUS, l.ch)
	case '-':
		tok = newToken(l, token.MINUS, l.ch)
	case '%':
		tok = newToken(l, token.PERCENT, l.ch)
	case '!':
		tok = twoChar(l, token.BANG, token.NOT_EQ, '=')
	case '/':
		if l.peekChar() == '/' {
			tok = l.readComment()
		} else {
			tok = newToken(l, token.SLASH, l.ch)
		}
	case '*':
		tok = newToken(l, token.ASTERISK, l.ch)
	case '<':
		tok = newToken(l, token.LT, l.ch)
	case '>':
		tok = newToken(l, token.GT, l.ch)
	case '(':
		tok = newToken(l, token.LPAREN, l.ch)
	case ')':
		tok = newToken(l, token.RPAREN, l.ch)
	case '{':
		tok = newToken(l, token.LBRACE, l.ch)
	case '}':
		tok = newToken(l, token.RBRACE, l.ch)
	case ';':
		tok = newToken(l, token.SEMICOLON, l.ch)
	case ':':
		tok = newToken(l, token.COLON, l.ch)
	case '"':
		tok.Literal = l.readString()
		tok.Type = token.STRING
	case '[':
		tok = newToken(l, token.LBRACKET, l.ch)
	case ']':
		tok = newToken(l, token.RBRACKET, l.ch)
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			return l.readNumber()
		} else {
			tok = newToken(l, token.ILLEGAL, l.ch)
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch == '?' || ch == '!'
}

func (l *Lexer) readString() string {
	l.readChar()
	position := l.position

	for l.ch != '"' {
		l.readChar()
	}

	return l.input[position:l.position]
}

func (l *Lexer) readComment() token.Token {
	l.readChar()
	l.readChar()
	position := l.position

	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}

	context := token.ContextData{LineNumber: l.lineNumber, File: l.filename}
	return token.Token{
		Type:    token.COMMENT,
		Literal: l.input[position:l.position],
		Context: context,
	}
}

func (l *Lexer) readNumber() token.Token {
	position := l.position

	var tokenType token.TokenType = token.INT

	for isDigit(l.ch) {
		l.readChar()
	}

	if isDecimal(l.ch) {
		if isDigit(l.peekChar()) {
			tokenType = token.FLOAT
			l.readChar()
			for isDigit(l.ch) {
				l.readChar()
			}
		} else {
			tokenType = token.ILLEGAL
		}
	}

	context := token.ContextData{LineNumber: l.lineNumber, File: l.filename}
	return token.Token{
		Type:    tokenType,
		Literal: l.input[position:l.position],
		Context: context,
	}
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isDecimal(ch byte) bool {
	return '.' == ch
}

func newToken(l *Lexer, tokenType token.TokenType, ch byte) token.Token {
	context := token.ContextData{LineNumber: l.lineNumber, File: l.filename}
	return token.Token{Type: tokenType, Literal: string(ch), Context: context}
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		if l.ch == '\n' {
			l.lineNumber += 1
		}
		l.readChar()
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	} else {
		return l.input[l.readPosition]
	}
}

func twoChar(l *Lexer, firstToken token.TokenType, secondToken token.TokenType, secondChar byte) token.Token {
	if l.peekChar() == secondChar {
		ch := l.ch
		l.readChar()
		literal := string(ch) + string(l.ch)
		context := token.ContextData{LineNumber: l.lineNumber, File: l.filename}
		return token.Token{Type: secondToken, Literal: literal, Context: context}
	} else {
		return newToken(l, firstToken, l.ch)
	}
}
