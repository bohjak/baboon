package lexer

import (
	"baboon/token"
	"fmt"
)

type Lexer struct {
	input        string
	position     int  // current position in input (current char)
	readPosition int  // current reading position in input (after current char)
	line         uint // current line position
	column       uint // current position in line
	ch           byte // current char under examination
}

// TODO: make work with a reader instead
func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		// TODO: replace this with a Unicode aware parser
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
	l.column += 1
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	fmt.Println("nexttoken", l.column)

	l.skipWhitespace()

	tok.Line = l.line
	tok.Column = l.column

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			tok.Type = token.EQ
			tok.Literal = "=="
		} else {
			tok.Type = token.ASSIGN
			tok.Literal = string(l.ch)
		}
	case ';':
		tok.Type = token.SEMICOLON
		tok.Literal = string(l.ch)
	case '{':
		tok.Type = token.LBRACE
		tok.Literal = string(l.ch)
	case '}':
		tok.Type = token.RBRACE
		tok.Literal = string(l.ch)
	case '(':
		tok.Type = token.LPAREN
		tok.Literal = string(l.ch)
	case ')':
		tok.Type = token.RPAREN
		tok.Literal = string(l.ch)
	case ',':
		tok.Type = token.COMMA
		tok.Literal = string(l.ch)
	case '+':
		tok.Type = token.PLUS
		tok.Literal = string(l.ch)
	case '-':
		tok.Type = token.MINUS
		tok.Literal = string(l.ch)
	case '/':
		tok.Type = token.SLASH
		tok.Literal = string(l.ch)
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			tok.Type = token.NEQ
			tok.Literal = "!="
		} else {
			tok.Type = token.BANG
			tok.Literal = string(l.ch)
		}
	case '*':
		tok.Type = token.ASTERISK
		tok.Literal = string(l.ch)
	case '<':
		tok.Type = token.LT
		tok.Literal = string(l.ch)
	case '>':
		tok.Type = token.GT
		tok.Literal = string(l.ch)
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdentifier(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok.Literal = l.readNumber()
			tok.Type = token.INT
			return tok
		} else {
			tok.Type = token.ILLEGAL
			tok.Literal = string(l.ch)
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) readIdentifier() string {
	start := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[start:l.position]
}

func (l *Lexer) readNumber() string {
	start := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[start:l.position]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		if l.ch == '\n' {
			fmt.Println("newline")
			l.line += 1
			l.column = 0
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

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	// TODO: expand to make work with floating point and non base-10 numbers
	return '0' <= ch && ch <= '9'
}
