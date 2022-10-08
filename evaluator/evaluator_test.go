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
		{"!2", false},
		{"!0", false}, // FIXME: this definitely shouldn't be like this
		{"!!true", true},
		{"!!false", false},
		{"!!2", true},
		{"!!0", true},
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
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestNestedBlockStatements(t *testing.T) {
	input := `
if (true) {
	if (true) {
		return 3;
	}

	return 4;
}`

	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 3)
}

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
