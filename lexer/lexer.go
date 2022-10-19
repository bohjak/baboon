package lexer

import (
	"unicode"
	"unicode/utf8"

	"baboon/token"
)

type Lexer struct {
	input        string
	position     int  // current position in input (current char)
	readPosition int  // current reading position in input (after current char)
	line         int  // current line position
	column       int  // current position in line
	ch           rune // current char under examination
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
		l.position = l.readPosition
		l.readPosition += 1
	} else {
		r, size := utf8.DecodeRuneInString(l.input[l.readPosition:])

		if r == utf8.RuneError {
			l.ch = 0
		} else {
			l.ch = r
		}
		l.position = l.readPosition
		l.readPosition += size
	}
	l.column += 1
}

func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.input) {
		return 0
	} else {
		r, _ := utf8.DecodeRuneInString(l.input[l.readPosition:])
		if r == utf8.RuneError {
			return 0
		}
		return r
	}
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

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
		if l.peekChar() == '=' {
			l.readChar()
			tok.Type = token.LEQ
			tok.Literal = "<="
		} else {
			tok.Type = token.LT
			tok.Literal = string(l.ch)
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok.Type = token.GEQ
			tok.Literal = ">="
		} else {
			tok.Type = token.GT
			tok.Literal = string(l.ch)
		}
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
	for isIdentifier(l.ch) {
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
			l.line += 1
			l.column = 0
		}
		l.readChar()
	}
}

func isLetter(ch rune) bool {
	// TODO: add support for emoji?
	return unicode.IsLetter(ch) || ch == '_'
}

func isDigit(ch rune) bool {
	// TODO: expand to make work with floating point and non base-10 numbers
	// parseInt cannot parse unicode numbers
	return ch >= '0' && ch <= '9'
}

func isIdentifier(ch rune) bool {
	return isLetter(ch) || isDigit(ch) || ch == '-'
}
