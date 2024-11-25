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
	Line    int
	Column  int
}

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
	line         int
	column       int
}

func NewLexer(input string) *Lexer {
	l := &Lexer{input: input, line: 1, column: 0}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}

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

	line := l.line
	column := l.column

	switch l.ch {
	case '\\':
		ch := l.peekChar()
		if ch == '@' || ch == '*' || ch == '$' || ch == '\\' {
			l.readChar()
			tok = Token{Type: TOKEN_ESCAPED_CHAR, Literal: string(ch), Line: line, Column: column}
			l.readChar()
		} else {
			tok = Token{Type: TOKEN_TEXT, Literal: string(l.ch), Line: line, Column: column}
			l.readChar()
		}
	case '@':
		if l.peekChar() == '{' {
			l.readChar()
			l.readChar()
			tok = Token{Type: TOKEN_FRAGMENT_REF, Literal: "@{", Line: line, Column: column}
		} else {
			tok = Token{Type: TOKEN_TEXT, Literal: string(l.ch), Line: line, Column: column}
			l.readChar()
		}
	case '*':
		if l.peekChar() == '{' {
			l.readChar()
			l.readChar()
			tok = Token{Type: TOKEN_BUILDER_REF, Literal: "*{", Line: line, Column: column}
		} else {
			tok = Token{Type: TOKEN_TEXT, Literal: string(l.ch), Line: line, Column: column}
			l.readChar()
		}
	case '$':
		if l.peekChar() == '{' {
			l.readChar()
			l.readChar()
			tok = Token{Type: TOKEN_META_REF, Literal: "${", Line: line, Column: column}
		} else {
			tok = Token{Type: TOKEN_TEXT, Literal: string(l.ch), Line: line, Column: column}
			l.readChar()
		}
	case '{':
		tok = Token{Type: TOKEN_OPEN_BRACE, Literal: "{", Line: line, Column: column}
		l.readChar()
	case '}':
		tok = Token{Type: TOKEN_CLOSE_BRACE, Literal: "}", Line: line, Column: column}
		l.readChar()
	case '[':
		if l.peekChar() == '[' {
			l.readChar()
			l.readChar()
			tok = Token{Type: TOKEN_DOUBLE_OPEN_BRACKET, Literal: "[[", Line: line, Column: column}
		} else {
			tok = Token{Type: TOKEN_TEXT, Literal: string(l.ch), Line: line, Column: column}
			l.readChar()
		}
	case ']':
		if l.peekChar() == ']' {
			l.readChar()
			l.readChar()
			tok = Token{Type: TOKEN_DOUBLE_CLOSE_BRACKET, Literal: "]]", Line: line, Column: column}
		} else {
			tok = Token{Type: TOKEN_TEXT, Literal: string(l.ch), Line: line, Column: column}
			l.readChar()
		}
	case 0:
		tok = Token{Type: TOKEN_EOF, Literal: "", Line: line, Column: column}
	default:
		literal := l.readText()
		tok = Token{Type: TOKEN_TEXT, Literal: literal, Line: line, Column: column}
	}

	return tok
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' {
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
	for l.ch != '\\' && l.ch != '@' && l.ch != '*' && l.ch != '$' &&
		l.ch != '{' && l.ch != '}' && l.ch != '[' && l.ch != ']' && l.ch != 0 {
		l.readChar()
	}
	return l.input[position:l.position]
}
