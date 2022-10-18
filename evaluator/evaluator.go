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

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node.Statements, env)
	case *ast.BlockStatement:
		return evalBlockStatement(node.Statements, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.ReturnStatement:
		val := Eval(node.Value, env)
		if val.Type() == object.ERROR_OBJ {
			return val
		}
		return &object.Return{Value: val}
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if right.Type() == object.ERROR_OBJ {
			return right
		}
		return evalPrefixExpression(node.Operator, right, node.Token)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if left.Type() == object.ERROR_OBJ {
			return left
		}
		right := Eval(node.Right, env)
		if right.Type() == object.ERROR_OBJ {
			return right
		}
		return evalInfixExpression(node.Operator, left, right, node.Token)
	case *ast.IfExpression:
		cond := Eval(node.Condition, env)
		if cond.Type() == object.ERROR_OBJ {
			return cond
		}
		return evalIfExpression(cond, node.Consequence, node.Alternative, env)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.Boolean:
		return ObjectBoolean(node.Value)
	case *ast.LetStatement:
		if _, ok := env.Get(node.Name.Value); ok {
			return &object.Error{Message: fmt.Sprintf("identifier already assigned: %s", node.Name.Value), Line: node.Token.Line, Column: node.Token.Column}
		}
		// TODO: switch to lazy evaluation?
		val := Eval(node.Value, env)
		if val.Type() == object.ERROR_OBJ {
			return val
		}
		return env.Set(node.Name.Value, val)
	case *ast.Identifier:
		val, ok := env.Get(node.Value)
		if !ok {
			return &object.Error{Message: fmt.Sprintf("identifier not found: %s", node.Value), Line: node.Token.Line, Column: node.Token.Column}
		}
		return val
	default:
		return &object.Error{Message: fmt.Sprintf("unknown node: [%T] %v", node, node)}
	}
}

func ObjectBoolean(value bool) *object.Boolean {
	if value {
		return TRUE
	}
	return FALSE
}

func evalProgram(stmts []ast.Statement, env *object.Environment) object.Object {
	var result object.Object

	for _, stmt := range stmts {
		result = Eval(stmt, env)
		switch result := result.(type) {
		case *object.Error:
			return result
		case *object.Return:
			return result.Value
		}
	}

	return result
}

func evalBlockStatement(stmts []ast.Statement, env *object.Environment) object.Object {
	var result object.Object

	for _, stmt := range stmts {
		result = Eval(stmt, env)
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

func evalIfExpression(condition object.Object, consequence *ast.BlockStatement, alternative *ast.BlockStatement, env *object.Environment) object.Object {
	if condition != FALSE && condition != NULL {
		return Eval(consequence, env)
	} else if alternative != nil {
		return Eval(alternative, env)
	} else {
		return NULL
	}
}
