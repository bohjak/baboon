package parser

import (
	"fmt"
	"testing"

	"baboon/ast"
	"baboon/lexer"
)

func TestOperatorPrecedence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"-a * b", "((-a) * b)"},
		{"!-a", "(!(-a))"},
		{"a + b + c", "((a + b) + c)"},
		{"a + b * c", "(a + (b * c))"},
		{"a * b + c", "((a * b) + c)"},
		{"5 > 4 == 3 < 6", "((5 > 4) == (3 < 6))"},
		{"a + b; -c * d", "(a + b)((-c) * d)"},
		{"(a + b) * c", "((a + b) * c)"},
		{"a + (b + c) + d", "((a + (b + c)) + d)"},
		{"!(a < b)", "(!(a < b))"},
		{"!true == !(false != true)", "((!true) == (!(false != true)))"},
		{"a + add(b * c) * d", "(a + (add((b * c)) * d))"},
		{"add(a, add(b, c / d), -f < g)", "add(a, add(b, (c / d)), ((-f) < g))"},
	}

	for i, tt := range tests {
		program := testParse(t, tt.input)

		actual := program.String()

		if actual != tt.expected {
			t.Errorf("test[%d]: expected %q, got %q", i, tt.expected, actual)
		}
	}
}

func TestInfixExpressions(t *testing.T) {
	infixTests := []struct {
		input      string
		leftValue  int64
		operator   string
		rightValue int64
	}{
		{"5 + 6;", 5, "+", 6},
		{"5 - 6;", 5, "-", 6},
		{"5 * 6;", 5, "*", 6},
		{"5 / 6;", 5, "/", 6},
		{"5 < 6;", 5, "<", 6},
		{"5 > 6;", 5, ">", 6},
		{"5 == 6;", 5, "==", 6},
		{"5 != 6;", 5, "!=", 6},
		{"5 <= 6;", 5, "<=", 6},
		{"5 >= 6;", 5, ">=", 6},
	}

	for _, tt := range infixTests {
		program := testParse(t, tt.input)

		assertStatementsLen(t, program.Statements, 1)

		stmt := assertExpressionStatement(t, program.Statements[0])

		if !testInfixExpression(t, stmt.Expression, tt.leftValue, tt.operator, tt.rightValue) {
			return
		}
	}
}

func TestPrefixExpressions(t *testing.T) {
	prefixTests := []struct {
		input        string
		operator     string
		integerValue int64
	}{
		{"!5;", "!", 5},
		{"-15;", "-", 15},
	}

	for _, tt := range prefixTests {
		program := testParse(t, tt.input)

		assertStatementsLen(t, program.Statements, 1)

		stmt := assertExpressionStatement(t, program.Statements[0])

		exp, ok := stmt.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("expression is not ast.PrefixExpression, got %T", stmt.Expression)
		}

		if exp.Operator != tt.operator {
			t.Fatalf("exp.Operator is not %q, got %q", tt.operator, exp.Operator)
		}

		if !testLiteralExpression(t, exp.Right, tt.integerValue) {
			return
		}
	}
}

func TestLiteralExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`5`, 5},
		{`"Hello, World!"`, "Hello, World!"},
		{`true`, true},
		{`false`, false},
		{`foo`, "foo"},
	}

	for _, tt := range tests {
		prog := testParse(t, tt.input)
		assertStatementsLen(t, prog.Statements, 1)
		stmt := assertExpressionStatement(t, prog.Statements[0])
		testLiteralExpression(t, stmt.Expression, tt.expected)
	}
}

func TestIfExpression(t *testing.T) {
	input := `if (x < y) { x }`

	program := testParse(t, input)

	assertStatementsLen(t, program.Statements, 1)

	stmt := assertExpressionStatement(t, program.Statements[0])

	exp, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("stmt.Expression not ast.IfExpression, got %T", stmt.Expression)
	}

	if !testInfixExpression(t, exp.Condition, "x", "<", "y") {
		return
	}

	assertStatementsLen(t, exp.Consequence.Statements, 1)

	consequence := assertExpressionStatement(t, exp.Consequence.Statements[0])

	if !testIdentifier(t, consequence.Expression, "x") {
		return
	}

	if exp.Alternative != nil {
		t.Errorf("exp.Alternative was not nil, got %+v", exp.Alternative)
	}
}

func TestIfElseExpression(t *testing.T) {
	input := `if (x < y) { x } else { y }`

	program := testParse(t, input)

	assertStatementsLen(t, program.Statements, 1)

	stmt := assertExpressionStatement(t, program.Statements[0])

	exp, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("stmt.Expression not ast.IfExpression, got %T", stmt.Expression)
	}

	if !testInfixExpression(t, exp.Condition, "x", "<", "y") {
		return
	}

	assertStatementsLen(t, exp.Consequence.Statements, 1)

	consequence := assertExpressionStatement(t, exp.Consequence.Statements[0])

	if !testIdentifier(t, consequence.Expression, "x") {
		return
	}

	if exp.Alternative == nil {
		t.Errorf("exp.Alternative was nil")
	}

	assertStatementsLen(t, exp.Alternative.Statements, 1)

	alternative := assertExpressionStatement(t, exp.Alternative.Statements[0])

	if !testIdentifier(t, alternative.Expression, "y") {
		return
	}
}

func TestFunctionExpression(t *testing.T) {
	input := `fn(x, y) { x + y; }`

	program := testParse(t, input)

	assertStatementsLen(t, program.Statements, 1)

	stmt := assertExpressionStatement(t, program.Statements[0])

	function, ok := stmt.Expression.(*ast.FunctionExpression)
	if !ok {
		t.Fatalf("expression is not ast.FunctionExpression, got %T", stmt.Expression)
	}

	if len(function.Parameters) != 2 {
		t.Fatalf("expected 2 parameters, got %d", len(function.Parameters))
	}

	testLiteralExpression(t, function.Parameters[0], "x")
	testLiteralExpression(t, function.Parameters[1], "y")

	assertStatementsLen(t, function.Body.Statements, 1)

	body := assertExpressionStatement(t, function.Body.Statements[0])

	testInfixExpression(t, body.Expression, "x", "+", "y")
}

func TestFunctionParameters(t *testing.T) {
	tests := []struct {
		input          string
		expectedParams []string
	}{
		{"fn() {};", []string{}},
		{"fn(x) {};", []string{"x"}},
		{"fn(x, y, z) {};", []string{"x", "y", "z"}},
	}

	for _, tt := range tests {
		program := testParse(t, tt.input)

		assertStatementsLen(t, program.Statements, 1)
		stmt := assertExpressionStatement(t, program.Statements[0])

		function, ok := stmt.Expression.(*ast.FunctionExpression)
		if !ok {
			t.Fatalf("expression is not ast.FunctionExpression, got %T", stmt.Expression)
		}

		if len(function.Parameters) != len(tt.expectedParams) {
			t.Fatalf("incorrect number of parameters, expected %d, got %d", len(tt.expectedParams), len(function.Parameters))
		}

		for i, param := range tt.expectedParams {
			testLiteralExpression(t, function.Parameters[i], param)
		}
	}
}

func TestCallExpression(t *testing.T) {
	input := "add(1, 2 * 3, 4 + 5);"

	program := testParse(t, input)

	assertStatementsLen(t, program.Statements, 1)
	stmt := assertExpressionStatement(t, program.Statements[0])

	exp, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.CallExpression, got %T", stmt.Expression)
	}

	if !testIdentifier(t, exp.Function, "add") {
		return
	}

	if len(exp.Arguments) != 3 {
		t.Fatalf("incorrect number of arguments, expected 3, got %d", len(exp.Arguments))
	}

	testLiteralExpression(t, exp.Arguments[0], 1)
	testInfixExpression(t, exp.Arguments[1], 2, "*", 3)
	testInfixExpression(t, exp.Arguments[2], 4, "+", 5)
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input         string
		expectedValue interface{}
	}{
		{"return 3;", 3},
		{"return false;", false},
		{"return x", "x"},
	}

	for _, tt := range tests {
		program := testParse(t, tt.input)

		assertStatementsLen(t, program.Statements, 1)

		stmt := program.Statements[0]
		returnStmt, ok := stmt.(*ast.ReturnStatement)
		if !ok {
			t.Errorf("stmt not *ast.returnStatement; got %T", stmt)
			continue
		}

		if returnStmt.TokenLiteral() != "return" {
			t.Errorf("returnStmt.TokenLiteral not 'return'; got %q", returnStmt.TokenLiteral())
		}

		if !testLiteralExpression(t, returnStmt.Value, tt.expectedValue) {
			continue
		}
	}
}

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedValue      interface{}
	}{
		{"let x = 5;", "x", 5},
		{"let y = true;", "y", true},
		{"let z = a;", "z", "a"},
	}

	for _, tt := range tests {
		program := testParse(t, tt.input)

		assertStatementsLen(t, program.Statements, 1)

		stmt := program.Statements[0]
		if !testLetStatement(t, stmt, tt.expectedIdentifier) {
			return
		}

		val := stmt.(*ast.LetStatement).Value
		if !testLiteralExpression(t, val, tt.expectedValue) {
			return
		}
	}
}

func TestArrayLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`[1, 2, 3]`, `[ 1, 2, 3 ]`},
		{`["foo", 3 + 4, true]`, `[ "foo", (3 + 4), true ]`},
		{`[]`, `[  ]`},
	}

	for _, tt := range tests {
		prog := testParse(t, tt.input)

		assertStatementsLen(t, prog.Statements, 1)

		stmt := assertExpressionStatement(t, prog.Statements[0])

		arr, ok := stmt.Expression.(*ast.ArrayLiteral)
		if !ok {
			t.Errorf("expression not ast.ArrayLiteral, got %T", stmt.Expression)
			continue
		}

		if arr.String() != tt.expected {
			t.Errorf("array has wrong content; expected %s, got %s", tt.expected, arr.String())
		}
	}
}

func TestAccessExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"[][0]", 0},
		{"[0, 2, 4][1]", 1},
		{"foo[2]", 2},
	}

	for i, tt := range tests {
		prog := testParse(t, tt.input)

		assertStatementsLen(t, prog.Statements, 1)
		stmt := assertExpressionStatement(t, prog.Statements[0])

		exp, ok := stmt.Expression.(*ast.AccessExpression)
		if !ok {
			t.Errorf("[%d] expression is not ast.AccessExpression, got %T", i, stmt.Expression)
			continue
		}

		testLiteralExpression(t, exp.Key, tt.expected)
	}
}

/* HELPERS */

func testParse(t *testing.T, input string) *ast.Program {
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	return program
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()

	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %s", msg)
	}
	t.FailNow()
}

func assertStatementsLen(t *testing.T, stmts []ast.Statement, n int) {
	if len(stmts) != n {
		t.Fatalf("statements expected %d, got %d", n, len(stmts))
	}
}

func assertExpressionStatement(t *testing.T, stmt ast.Statement) *ast.ExpressionStatement {
	res, ok := stmt.(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("statement is not ast.ExpressionStatement, got %T", stmt)
	}
	return res
}

func testIntegerLiteral(t *testing.T, exp ast.Expression, value int64) bool {
	integ, ok := exp.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("exp not *ast.IntegerLiteral, got %T", exp)
		return false
	}

	if integ.Value != value {
		t.Errorf("integ.Value not %d, got %d", value, integ.Value)
		return false
	}

	if integ.Token.Literal != fmt.Sprint(value) {
		t.Errorf("integ.Token.Literal not %d, got %q", value, integ.Token.Literal)
		return false
	}

	return true
}

func testIdentifier(t *testing.T, exp ast.Expression, value string) bool {
	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Errorf("exp not ast.Identifier, got %T", exp)
		return false
	}

	if ident.Value != value {
		t.Errorf("ident.Value not %q, got %q", value, ident.Value)
		return false
	}

	if ident.Token.Literal != value {
		t.Errorf("ident.Token.Literal not %q, got %q", value, ident.Value)
		return false
	}

	return true
}

func testBoolean(t *testing.T, exp ast.Expression, value bool) bool {
	boolean, ok := exp.(*ast.Boolean)
	if !ok {
		t.Errorf("exp not ast.Boolean, got %T", exp)
		return false
	}

	if boolean.Value != value {
		t.Errorf("boolean.Value not %v, got %v", value, boolean.Value)
		return false
	}

	if boolean.Token.Literal != fmt.Sprint(value) {
		t.Errorf("boolean.Token.Literal not %q, got %q", fmt.Sprint(value), boolean.Token.Literal)
		return false
	}

	return true
}

func testStringLiteral(t *testing.T, exp ast.Expression, value string) bool {
	strLit, ok := exp.(*ast.StringLiteral)
	if !ok {
		t.Errorf("exp not ast.StringLiteral, got %T", exp)
		return false
	}

	if strLit.Value != value {
		t.Errorf("exp has wrong value; expected %q, got %q", value, strLit.Value)
		return false
	}

	return true
}

func testLiteralExpression(t *testing.T, exp ast.Expression, expected interface{}) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int64(v))
	case int64:
		return testIntegerLiteral(t, exp, v)
	case string:
		if id, ok := exp.(*ast.Identifier); ok {
			return testIdentifier(t, id, v)

		} else {
			return testStringLiteral(t, exp, v)
		}
	case bool:
		return testBoolean(t, exp, v)
	default:
		t.Errorf("type of exp not handled, got %T", exp)
		return false
	}
}

func testLetStatement(t *testing.T, s ast.Statement, name string) bool {
	if s.TokenLiteral() != "let" {
		t.Errorf("s.TokenLiteral not 'let'; got %q", s.TokenLiteral())
		return false
	}

	letStmt, ok := s.(*ast.LetStatement)
	if !ok {
		t.Errorf("s not *ast.LetStatement; got %T", s)
		return false
	}

	if !testIdentifier(t, letStmt.Name, name) {
		return false
	}

	return true
}

func testInfixExpression(t *testing.T, exp ast.Expression, left interface{}, operator string, right interface{}) bool {
	opExp, ok := exp.(*ast.InfixExpression)
	if !ok {
		t.Errorf("exp is not ast.InfixExpression, got %T", exp)
		return false
	}

	if !testLiteralExpression(t, opExp.Left, left) {
		return false
	}

	if opExp.Operator != operator {
		t.Errorf("exp.Operator is not %q, got %q", operator, opExp.Operator)
		return false
	}

	if !testLiteralExpression(t, opExp.Right, right) {
		return false
	}

	return true
}
