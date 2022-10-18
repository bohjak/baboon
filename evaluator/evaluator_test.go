package evaluator

import (
	"baboon/lexer"
	"baboon/object"
	"baboon/parser"
	"testing"
)

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"42", 42},
		{"-5", -5},
		{"-42", -42},
		{"1 + 2 + 3", 6},
		{"2 * 2 * 2 * 2", 16},
		{"50 / 2 * 3 + 10 - 3 * 2", 79},
	}

	for i, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, i, evaluated, tt.expected)
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1 == 0", false},
		{"1 != 0", true},
		{"30 < 92", true},
		{"53 > 55", false},
		{"true == true", true},
		{"false != true", true},
	}

	for i, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, i, evaluated, tt.expected)
	}
}

func TestEvalBangExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!!true", true},
		{"!!false", false},
	}

	for i, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, i, evaluated, tt.expected)
	}
}

func TestEvalIfExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if (true) {5}", 5},
		{"if (false) {5}", nil},
		{"if (true) {1} else {2}", 1},
		{"if (false) {1} else {2}", 2},
	}

	for i, tt := range tests {
		evaluated := testEval(tt.input)

		if expected, ok := tt.expected.(int); ok {
			testIntegerObject(t, i, evaluated, int64(expected))
		} else {
			testNullObject(t, i, evaluated)
		}
	}
}

func TestReturnStatement(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"return 3", 3},
		{"return 8 + 4", 12},
		{"return 5; 9", 5},
		{`
if (true) {
	if (true) {
		return 3;
	}

	return 4;
}
		`, 3},
		{"if (false) {!1} else {1}", 1},
	}

	for i, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, i, evaluated, tt.expected)
	}
}

func TestErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"!0", "unknown operator: !INTEGER"},
		{"-true", "unknown operator: -BOOLEAN"},
		{"true + false", "unknown operator: BOOLEAN + BOOLEAN"},
		{"0 < true", "type mismatch: INTEGER < BOOLEAN"},
		{"true == 2; 4", "type mismatch: BOOLEAN == INTEGER"},
		{"3 - true + 4", "type mismatch: INTEGER - BOOLEAN"},
		{"if (!!1) {true}", "unknown operator: !INTEGER"},
		{"true >= false", "unknown operator: BOOLEAN >= BOOLEAN"},
		{"let a = if (true) {return -false}", "unknown operator: -BOOLEAN"},
		{"foo", "identifier not found: foo"},
		{"let a = 1; let a = 2", "identifier already assigned: a"},
	}

	for i, tt := range tests {
		evaluated := testEval(tt.input)
		testErrorObject(t, i, evaluated, tt.expected)
	}
}

func TestBinding(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let a = 5;", 5},
		{"let a = 5; a", 5},
		{"let a = 6; let b = 6 * 6; b", 36},
		{"let a = 3; let b = 4; a + b", 7},
	}

	for i, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, i, evaluated, tt.expected)
	}
}

/* HELPERS */

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		return &object.Error{Message: "program could not be parsed"}
	}
	env := object.NewEnvironment()
	return Eval(program, env)
}

func testIntegerObject(t *testing.T, i int, obj object.Object, expected int64) bool {
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("[%d] object is not Integer, got %T", i, obj)
		return false
	}

	if result.Value != expected {
		t.Errorf("[%d] object has wrong value, expected %d, got %d", i, expected, result.Value)
		return false
	}

	return true
}

func testBooleanObject(t *testing.T, i int, obj object.Object, expected bool) bool {
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("[%d] object is not Boolean, got %T", i, obj)
		return false
	}

	if result.Value != expected {
		t.Errorf("[%d] object has wrong value, expected %T, got %T", i, expected, result.Value)
		return false
	}

	return true
}

func testNullObject(t *testing.T, i int, obj object.Object) bool {
	if obj != NULL {
		t.Errorf("[%d] object is not Null, got %T", i, obj)
		return false
	}

	return true
}

func testErrorObject(t *testing.T, i int, obj object.Object, expected string) bool {
	result, ok := obj.(*object.Error)
	if !ok {
		t.Errorf("[%d] object is not Error, got %T", i, obj)
		return false
	}

	if result.Message != expected {
		t.Errorf("[%d] object has wrong message\nexpected:\t%q\ngot:\t\t%q", i, expected, result.Message)
		return false
	}

	return true
}
