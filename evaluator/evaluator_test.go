package evaluator

import (
	"baboon/lexer"
	"baboon/object"
	"baboon/parser"
	"testing"
)

func TestEvalVoid(t *testing.T) {
	tests := []struct {
		input string
	}{
		{""},
		{"fn(){}()"},
		{"if true {}"},
		{"()"},
	}

	for i, tt := range tests {
		eval := testEval(tt.input)
		testVoidObject(t, i, eval)
	}
}

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

func TestEvalStringExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"Hello, world!"`, "Hello, world!"},
		{`let name = "Zalgo"; "My name is" + " " + name`, "My name is Zalgo"},
		{`if ("a" == "a") { "true" }`, "true"},
		{`if ("a" != "b") { "true" }`, "true"},
	}

	for i, tt := range tests {
		evaluated := testEval(tt.input)
		testStringObject(t, i, evaluated, tt.expected)
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
		{"if true {5}", 5},
		{"if false {5}", nil},
		{"if true {1} else {2}", 1},
		{"if false {1} else {2}", 2},
		{"if false {true == 5} else {2}", 2},
		{"fn(ns){if len(ns) == 0 {return 1} else {return 2}}([1])", 2},
		{"if 1 == 2 {true}", nil},
		{"if true {}", nil},
	}

	for i, tt := range tests {
		evaluated := testEval(tt.input)

		if expected, ok := tt.expected.(int); ok {
			testIntegerObject(t, i, evaluated, int64(expected))
		} else {
			testVoidObject(t, i, evaluated)
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

func TestError(t *testing.T) {
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
		{`"hello" + 42`, "type mismatch: STRING + INTEGER"},
		{`"a" < "b"`, "unknown operator: STRING < STRING"},
		{"[][3]", "invalid argument: index 3 out of bounds"},
		{"[1, 2][2]", "invalid argument: index 2 out of bounds"},
		{"[1, 2][-3]", "invalid argument: index -3 out of bounds"},
		{"[1, 2][true]", "invalid argument: ARRAY[BOOLEAN]"},
		{"if 1 {2}", "non-boolean condition in IF expression: INTEGER"},
		{"if if false {} {}", "non-boolean condition in IF expression: VOID"},
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

func TestFunctionExpression(t *testing.T) {
	tests := []struct {
		input          string
		expectedParams []string
		expectedBody   string
	}{
		{"fn(a){a * 2}", []string{"a"}, "(a * 2)"},
		{"fn(foo, bar, baz) { foo + bar + baz }", []string{"foo", "bar", "baz"}, "((foo + bar) + baz)"},
	}

	for i, tt := range tests {
		evaluated := testEval(tt.input)
		testFunctionObject(t, i, evaluated, tt.expectedParams, tt.expectedBody)
	}
}

func TestCallExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let id = fn(x) { x }; id(42)", 42},
		{"let id = fn(x) { return x }; id(42)", 42},
		{"let mul = fn(a, b) { a * b }; mul(3, 4)", 12},
		{"let apply = fn(x, cb) { cb(x) }; let sqr = fn(n) { n * n }; apply(8, sqr)", 64},
		{"let add = fn(a, b) { a + b }; add(add(1, 2), fn(n) {n / 2}(10))", 8},
		{"let foo = 1; let bar = fn(foo) { foo + 2 }; bar(3)", 5},
		{"let newAdder = fn(x) {fn(y) {x + y}}; let addTwo = newAdder(2); addTwo(8)", 10},
	}

	for i, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, i, evaluated, tt.expected)
	}
}

func TestBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`len("123456")`, 6},
		{`len("৩")`, 3},
		{`len("")`, 0},
		{`len("šíleně žluťoučký ৩æ")`, 29},
		{`len(42)`, "invalid argument: len(INTEGER)"},
		{`len("abc", "def")`, "too many arguments for len: expected 1, found 2"},
		{`len()`, "not enough arguments for len: expected 1, found 0"},
		{`len([1, 2, 3, 4, 5])`, 5},
		{`len([])`, 0},
		{`len(tail([]))`, 0},
		{`
let map = fn(arr, f) {
	let aux = fn(acc, arr) {
		if len(arr) == 0 {
			return acc;
		}
		aux(append(acc, f(first(arr))), tail(arr));
	};
	aux([], arr)
};

let sum = fn(arr) {
	let aux = fn(acc, arr) {
		if len(arr) == 0 {
			return acc
		}
		aux(acc + first(arr), tail(arr))
	};
	aux(0, arr)
};

let ns = [1, 2, 3, 4, 5]
sum(map(ns, fn(n) {n * 2}))
		`, 30},
	}

	for i, tt := range tests {
		eval := testEval(tt.input)

		switch e := tt.expected.(type) {
		case int:
			testIntegerObject(t, i, eval, int64(e))
		case string:
			testErrorObject(t, i, eval, e)
		}
	}
}

func TestArrayLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected []int
	}{
		{`[0, 1, 2]`, []int{0, 1, 2}},
		{`[]`, []int{}},
	}

	for i, tt := range tests {
		eval := testEval(tt.input)

		array, ok := eval.(*object.ArrayLiteral)
		if !ok {
			t.Errorf("[%d] object is not Array, got %T", i, eval)
			continue
		}

		for j, expected := range tt.expected {
			item, ok := array.Items[j].(*object.Integer)
			if !ok {
				t.Errorf("[%d] array item %d is not Integer, got %s", i, j, array.Items[j].Type())
				continue
			}

			if item.Value != int64(expected) {
				t.Errorf("[%d] array item %d has wrong value; expected %d, got %d", i, j, expected, item.Value)
			}
		}
	}
}

func TestArrayAccess(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`[0, 1, 2, 3][2]`, 2},
		{`[42, "abc", true][1]`, "abc"},
		{`let arr = [true, false]; arr[0]`, true},
		{`let foo = [fn(bar) { len(bar) }]; foo[0]("12345")`, 5},
		{`[3, 2, 1][-2]`, 2},
	}

	for i, tt := range tests {
		eval := testEval(tt.input)

		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, i, eval, int64(expected))
		case string:
			testStringObject(t, i, eval, expected)
		case bool:
			testBooleanObject(t, i, eval, expected)
		}
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

func testStringObject(t *testing.T, i int, obj object.Object, expected string) bool {
	result, ok := obj.(*object.String)
	if !ok {
		t.Errorf("[%d] object is not String, got %T", i, obj)
		return false
	}

	if result.Value != expected {
		t.Errorf("[%d] object has wrong value, expected %q, got %q", i, expected, result.Value)
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

func testVoidObject(t *testing.T, i int, obj object.Object) bool {
	if obj != VOID {
		t.Errorf("[%d] object is not Void, got %T", i, obj)
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

func testFunctionObject(t *testing.T, i int, obj object.Object, expectedParams []string, expectedBody string) bool {
	result, ok := obj.(*object.Function)
	if !ok {
		t.Errorf("[%d] object is not Function, got %T", i, obj)
		return false
	}

	if len(result.Parameters) != len(expectedParams) {
		t.Errorf("[%d] function has wrong number of parameters; expected %d, got %d", i, len(expectedParams), len(result.Parameters))
		return false
	}

	for j, p := range expectedParams {
		if p != result.Parameters[j].String() {
			t.Errorf("[%d] function has wrong parameter number %d; expected %q, got %q", i, j, p, result.Parameters[j].String())
			return false
		}
	}

	if expectedBody != result.Body.String() {
		t.Errorf("[%d] function has wrong body; expected %q, got %q", i, expectedBody, result.Body.String())
		return false
	}

	return true
}
