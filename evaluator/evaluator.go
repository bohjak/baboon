package evaluator

import (
	"baboon/ast"
	"baboon/object"
)

var (
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
	NULL  = &object.Null{}
)

func Eval(node ast.Node) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node.Statements)
	case *ast.BlockStatement:
		return evalBlockStatement(node.Statements)
	case *ast.ExpressionStatement:
		return Eval(node.Expression)
	case *ast.PrefixExpression:
		return evalPrefixExpression(node.Operator, Eval(node.Right))
	case *ast.InfixExpression:
		return evalInfixExpression(node.Operator, Eval(node.Left), Eval(node.Right))
	case *ast.IfExpression:
		return evalIfExpression(Eval(node.Condition), node.Consequence, node.Alternative)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.Boolean:
		return ObjectBoolean(node.Value)
	default:
		return nil
	}
}

func ObjectBoolean(value bool) *object.Boolean {
	if value == true {
		return TRUE
	}
	return FALSE
}

func evalProgram(stmts []ast.Statement) object.Object {
	var result object.Object

	for _, stmt := range stmts {
		result = Eval(stmt)
	}

	return result
}

func evalBlockStatement(stmts []ast.Statement) object.Object {
	var result object.Object

	for _, stmt := range stmts {
		result = Eval(stmt)
	}

	return result
}

func evalPrefixExpression(op string, right object.Object) object.Object {
	switch op {
	case "!":
		return evalBangExpression(right)
	case "-":
		return evalMinusPrefixExpression(right)
	default:
		// TODO: should throw an error
		return NULL
	}
}

func evalBangExpression(obj object.Object) object.Object {
	switch obj {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		// FIXME: should throw or at least handle 0 as falsy
		return FALSE
	}
}

func evalMinusPrefixExpression(obj object.Object) object.Object {
	if obj.Type() != object.INTEGER_OBJ {
		// TODO: should throw
		return NULL
	}

	value := obj.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func evalInfixExpression(op string, left object.Object, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerExpression(op, left, right)
	case op == "==":
		return ObjectBoolean(left == right)
	case op == "!=":
		return ObjectBoolean(left != right)
	default:
		return NULL
	}
}

func evalIntegerExpression(op string, left object.Object, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch op {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		return &object.Integer{Value: leftVal / rightVal}
	case "<":
		return &object.Boolean{Value: leftVal < rightVal}
	case ">":
		return &object.Boolean{Value: leftVal > rightVal}
	case "<=":
		return &object.Boolean{Value: leftVal <= rightVal}
	case ">=":
		return &object.Boolean{Value: leftVal >= rightVal}
	case "==":
		return &object.Boolean{Value: leftVal == rightVal}
	case "!=":
		return &object.Boolean{Value: leftVal != rightVal}
	default:
		// TODO: should throw
		return NULL
	}
}

func evalIfExpression(condition object.Object, consequence *ast.BlockStatement, alternative *ast.BlockStatement) object.Object {
	if condition != FALSE && condition != NULL {
		return Eval(consequence)
	} else if alternative != nil {
		return Eval(alternative)
	} else {
		return NULL
	}
}
