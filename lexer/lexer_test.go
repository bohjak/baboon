package lexer

import (
	"testing"

	"baboon/token"
)

func TestNextTokenBasic(t *testing.T) {
	input := `=+(){},;!-/*5<>[]==<=>=:=::`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.ASSIGN, "="},
		{token.PLUS, "+"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.COMMA, ","},
		{token.SEMICOLON, ";"},
		{token.BANG, "!"},
		{token.MINUS, "-"},
		{token.SLASH, "/"},
		{token.ASTERISK, "*"},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.GT, ">"},
		{token.LBRACKET, "["},
		{token.RBRACKET, "]"},
		{token.EQ, "=="},
		{token.LEQ, "<="},
		{token.GEQ, ">="},
		{token.DEFINE, ":="},
		{token.CONST, "::"},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected %q, got %q", i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong, expected %q, got %q", i, tt.expectedLiteral, tok.Literal)

		}
	}
}

func TestNextTokenComplex(t *testing.T) {
	input := `
five = 5;
ten := 10;

add :: fn(x, y) {
	x + y;
};

result :: add("five", ten);

if (5 < 10) {
	return true;
} else {
	return false;
}

10 == 10;
10 != 9;
10 <= 11;
12 >= 11;
	`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.IDENT, "five"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},

		{token.IDENT, "ten"},
		{token.DEFINE, ":="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},

		{token.IDENT, "add"},
		{token.CONST, "::"},
		{token.FUNCTION, "fn"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.IDENT, "y"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"},

		{token.IDENT, "result"},
		{token.CONST, "::"},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.STRING, "five"},
		{token.COMMA, ","},
		{token.IDENT, "ten"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},

		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.TRUE, "true"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.ELSE, "else"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.FALSE, "false"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},

		{token.INT, "10"},
		{token.EQ, "=="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.INT, "10"},
		{token.NEQ, "!="},
		{token.INT, "9"},
		{token.SEMICOLON, ";"},

		{token.INT, "10"},
		{token.LEQ, "<="},
		{token.INT, "11"},
		{token.SEMICOLON, ";"},
		{token.INT, "12"},
		{token.GEQ, ">="},
		{token.INT, "11"},
		{token.SEMICOLON, ";"},

		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected %q, got %q:%q", i, tt.expectedType, tok.Type, tok.Literal)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong, expected %q, got %q", i, tt.expectedLiteral, tok.Literal)

		}
	}
}

func TestTokenPosition(t *testing.T) {
	input := `
identifier = 321;
return!`

	l := New(input)

	tests := []struct {
		expectedType string
		expectedLine int
		expectedCol  int
	}{
		{token.IDENT, 2, 1},
		{token.ASSIGN, 2, 12},
		{token.INT, 2, 14},
		{token.SEMICOLON, 2, 17},
		{token.RETURN, 3, 1},
		{token.BANG, 3, 7},
	}

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Line != tt.expectedLine {
			t.Fatalf("tests[%d] - line wrong, expected %d, got %d", i, tt.expectedLine, tok.Line)
		}

		if tok.Column != tt.expectedCol {
			t.Fatalf("tests[%d] - column wrong, expected %d, got %d", i, tt.expectedCol, tok.Column)
		}
	}
}

func TestUnicode(t *testing.T) {
	tests := []struct {
		input        string
		expectedType token.TokenType
		expectedLit  string
	}{
		{"???", token.ILLEGAL, "???"},
		{"foo", token.IDENT, "foo"},
		{"??e-r??", token.IDENT, "??e-r??"},
		{"_abc123-", token.IDENT, "_abc123-"},
	}

	for i, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		testToken(t, i, tok, tt.expectedType, tt.expectedLit)
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"\"hello\"", "hello"},
		{"\"what a == wonderfully!, spacious: sentence\"", "what a == wonderfully!, spacious: sentence"},
	}

	for i, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		testToken(t, i, tok, token.STRING, tt.expected)
	}
}

/* HELPERS */

func testToken(t *testing.T, i int, tok token.Token, expType token.TokenType, expLit string) bool {
	if tok.Type != expType {
		t.Errorf("token is of wrong type; expected %s, got %s", expType, tok.Type)
		return false
	}

	if tok.Literal != expLit {
		t.Errorf("token has wrong literal; expected %q, got %q", expLit, tok.Literal)
		return false
	}

	return true
}
