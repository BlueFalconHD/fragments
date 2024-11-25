package main

type TokenType string

const (
	TOKEN_EOF                  = "EOF"
	TOKEN_TEXT                 = "TEXT"
	TOKEN_ESCAPED_CHAR         = "ESCAPED_CHAR"
	TOKEN_META_REF             = "META_REF"
	TOKEN_BUILDER_REF          = "BUILDER_REF"
	TOKEN_FRAGMENT_REF         = "FRAGMENT_REF"
	TOKEN_OPEN_BRACE           = "{"
	TOKEN_CLOSE_BRACE          = "}"
	TOKEN_DOUBLE_OPEN_BRACKET  = "[["
	TOKEN_DOUBLE_CLOSE_BRACKET = "]]"
)

type Token struct {
	Type    TokenType
	Literal string
}

type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
}

func NewLexer(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // ASCII code for NUL, signifies EOF
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace()

	switch l.ch {
	case '\\':
		ch := l.peekChar()
		if ch == '@' || ch == '*' || ch == '$' || ch == '\\' {
			l.readChar()
			tok = Token{Type: TOKEN_ESCAPED_CHAR, Literal: string(ch)}
			l.readChar()
		} else {
			tok = Token{Type: TOKEN_TEXT, Literal: string(l.ch)}
			l.readChar()
		}
	case '@':
		if l.peekChar() == '{' {
			l.readChar()
			l.readChar()
			tok = Token{Type: TOKEN_FRAGMENT_REF, Literal: "@{"}
		} else {
			tok = Token{Type: TOKEN_TEXT, Literal: string(l.ch)}
			l.readChar()
		}
	case '*':
		if l.peekChar() == '{' {
			l.readChar()
			l.readChar()
			tok = Token{Type: TOKEN_BUILDER_REF, Literal: "*{"}
		} else {
			tok = Token{Type: TOKEN_TEXT, Literal: string(l.ch)}
			l.readChar()
		}
	case '$':
		if l.peekChar() == '{' {
			l.readChar()
			l.readChar()
			tok = Token{Type: TOKEN_META_REF, Literal: "${"}
		} else {
			tok = Token{Type: TOKEN_TEXT, Literal: string(l.ch)}
			l.readChar()
		}
	case '{':
		tok = Token{Type: TOKEN_OPEN_BRACE, Literal: "{"}
		l.readChar()
	case '}':
		tok = Token{Type: TOKEN_CLOSE_BRACE, Literal: "}"}
		l.readChar()
	case '[':
		if l.peekChar() == '[' {
			l.readChar()
			l.readChar()
			tok = Token{Type: TOKEN_DOUBLE_OPEN_BRACKET, Literal: "[["}
		} else {
			tok = Token{Type: TOKEN_TEXT, Literal: string(l.ch)}
			l.readChar()
		}
	case ']':
		if l.peekChar() == ']' {
			l.readChar()
			l.readChar()
			tok = Token{Type: TOKEN_DOUBLE_CLOSE_BRACKET, Literal: "]]"}
		} else {
			tok = Token{Type: TOKEN_TEXT, Literal: string(l.ch)}
			l.readChar()
		}
	case 0:
		tok = Token{Type: TOKEN_EOF, Literal: ""}
	default:
		literal := l.readText()
		tok = Token{Type: TOKEN_TEXT, Literal: literal}
	}

	return tok
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) readText() string {
	position := l.position
	for l.ch != '\\' && l.ch != '@' && l.ch != '*' && l.ch != '$' && l.ch != '{' && l.ch != '}' && l.ch != '[' && l.ch != ']' && l.ch != 0 {
		l.readChar()
	}
	return l.input[position:l.position]
}
