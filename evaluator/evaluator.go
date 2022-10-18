package evaluator

import (
	"baboon/ast"
	"baboon/object"
	"baboon/token"
	"fmt"
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
	case *ast.ReturnStatement:
		return &object.Return{Value: Eval(node.Value)}
	case *ast.PrefixExpression:
		return evalPrefixExpression(node.Operator, Eval(node.Right), node.Token)
	case *ast.InfixExpression:
		return evalInfixExpression(node.Operator, Eval(node.Left), Eval(node.Right), node.Token)
	case *ast.IfExpression:
		return evalIfExpression(Eval(node.Condition), node.Consequence, node.Alternative)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.Boolean:
		return ObjectBoolean(node.Value)
	default:
		return &object.Error{Message: fmt.Sprintf("unknown node: %T", node)}
	}
}

func ObjectBoolean(value bool) *object.Boolean {
	if value {
		return TRUE
	}
	return FALSE
}

func evalProgram(stmts []ast.Statement) object.Object {
	var result object.Object

	for _, stmt := range stmts {
		result = Eval(stmt)
		if result.Type() == object.ERROR_OBJ {
			return result
		}
		if result, ok := result.(*object.Return); ok {
			return result.Value
		}
	}

	return result
}

func evalBlockStatement(stmts []ast.Statement) object.Object {
	var result object.Object

	for _, stmt := range stmts {
		result = Eval(stmt)
		if result.Type() == object.RETURN_OBJ || result.Type() == object.ERROR_OBJ {
			return result
		}
	}

	return result
}

func evalPrefixExpression(op string, right object.Object, token token.Token) object.Object {
	switch op {
	case "!":
		return evalBangExpression(right, token)
	case "-":
		return evalMinusPrefixExpression(right, token)
	default:
		return &object.Error{Message: fmt.Sprintf("unknown operator: %s%s", op, right.Type()), Line: token.Line, Column: token.Column}
	}
}

func evalBangExpression(obj object.Object, token token.Token) object.Object {
	switch obj {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return &object.Error{Message: fmt.Sprintf("unknown operator: !%s", obj.Type()), Line: token.Line, Column: token.Column}
	}
}

func evalMinusPrefixExpression(obj object.Object, token token.Token) object.Object {
	if obj.Type() != object.INTEGER_OBJ {
		return &object.Error{Message: fmt.Sprintf("unknown operator: -%s", obj.Type()), Line: token.Line, Column: token.Column}
	}

	value := obj.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func evalInfixExpression(op string, left object.Object, right object.Object, token token.Token) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerExpression(op, left, right, token)
	case left.Type() != right.Type():
		return &object.Error{Message: fmt.Sprintf("type mismatch: %s %s %s", left.Type(), op, right.Type()), Line: token.Line, Column: token.Column}
	case op == "==":
		return ObjectBoolean(left == right)
	case op == "!=":
		return ObjectBoolean(left != right)
	default:
		return &object.Error{Message: fmt.Sprintf("unknown operator: %s %s %s", left.Type(), op, right.Type()), Line: token.Line, Column: token.Column}
	}
}

func evalIntegerExpression(op string, left object.Object, right object.Object, token token.Token) object.Object {
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
		return &object.Error{Message: fmt.Sprintf("unknown operator: %s %s %s", left.Type(), op, right.Type()), Line: token.Line, Column: token.Column}
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
