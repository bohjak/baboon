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

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
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

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
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

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
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

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		if expected, ok := tt.expected.(int); ok {
			testIntegerObject(t, evaluated, int64(expected))
		} else {
			testNullObject(t, evaluated)
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
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
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
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testErrorObject(t, evaluated, tt.expected)
	}
}

/* HELPERS */

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	return Eval(program)
}

func testIntegerObject(t *testing.T, obj object.Object, expected int64) bool {
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("object is not Integer, got %T", obj)
		return false
	}

	if result.Value != expected {
		t.Errorf("object has wrong value, expected %d, got %d", expected, result.Value)
		return false
	}

	return true
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) bool {
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("object is not Boolean, got %T", obj)
		return false
	}

	if result.Value != expected {
		t.Errorf("object has wrong value, expected %T, got %T", expected, result.Value)
		return false
	}

	return true
}

func testNullObject(t *testing.T, obj object.Object) bool {
	if obj != NULL {
		t.Errorf("object is not Null, got %T", obj)
		return false
	}

	return true
}

func testErrorObject(t *testing.T, obj object.Object, expected string) bool {
	result, ok := obj.(*object.Error)
	if !ok {
		t.Errorf("object is not Error, got %T", obj)
		return false
	}

	if result.Message != expected {
		t.Errorf("object has wrong message\nexpected:\t%q\ngot:\t%q", expected, result.Message)
		return false
	}

	return true
}
