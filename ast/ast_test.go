package ast

import (
	"baboon/token"
	"testing"
)

func TestString(t *testing.T) {
	program := &Program{
		Statements: []Statement{
			&ExpressionStatement{
				Expression: &IfExpression{
					Token: token.Token{Type: token.IF, Literal: "if"},
					Condition: &Identifier{
						Token: token.Token{Type: token.IDENT, Literal: "myVar"},
						Value: "myVar",
					},
					Consequence: &BlockStatement{
						Statements: []Statement{
							&ExpressionStatement{
								Expression: &Identifier{
									Token: token.Token{Type: token.IDENT, Literal: "anotherVar"},
									Value: "anotherVar",
								},
							},
						},
					},
				},
			},
		},
	}

	if program.String() != "if myVar { anotherVar }" {
		t.Errorf("program.String() wrong; got %q", program.String())
	}
}
