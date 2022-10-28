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
	VOID  = &object.Void{}
)

func newBoolean(value bool) *object.Boolean {
	if value {
		return TRUE
	}
	return FALSE
}

func newError(token *token.Token, message string) *object.Error {
	return &object.Error{Message: message, Line: token.Line, Column: token.Column}
}

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
		return evalPrefixExpression(node.Operator, right, &node.Token)

	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if left.Type() == object.ERROR_OBJ {
			return left
		}
		right := Eval(node.Right, env)
		if right.Type() == object.ERROR_OBJ {
			return right
		}
		return evalInfixExpression(node.Operator, left, right, &node.Token)

	case *ast.IfExpression:
		cond := Eval(node.Condition, env)
		if cond.Type() == object.ERROR_OBJ {
			return cond
		}
		return evalIfExpression(cond, node.Consequence, node.Alternative, env, &node.Token)

	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}

	case *ast.StringLiteral:
		return &object.String{Value: node.Value}

	case *ast.Boolean:
		return newBoolean(node.Value)

	case *ast.AssignExpression:
		return evalAssignExpression(node, env)

	case *ast.Identifier:
		if val, ok := env.Get(node.Value); ok {
			return val
		}
		if builtin, ok := builtins[node.Value]; ok {
			return builtin
		}
		return newError(&node.Token, fmt.Sprintf("identifier not found: %s", node.Value))

	case *ast.FunctionExpression:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Body: body, Env: env}

	case *ast.CallExpression:
		fn := Eval(node.Function, env)
		if fn.Type() == object.ERROR_OBJ {
			return fn
		}

		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && args[0].Type() == object.ERROR_OBJ {
			return args[0]
		}

		return applyFunction(fn, args, &node.Token)

	case *ast.ArrayLiteral:
		items := evalExpressions(node.Items, env)
		if len(items) == 1 && items[0].Type() == object.ERROR_OBJ {
			return items[0]
		}

		return &object.ArrayLiteral{Items: items}

	case *ast.AccessExpression:
		arr := Eval(node.Array, env)
		if arr.Type() == object.ERROR_OBJ {
			return arr
		}

		key := Eval(node.Key, env)
		if key.Type() == object.ERROR_OBJ {
			return key
		}

		return evalAccessExpression(arr, key, &node.Token)

	default:
		// FIXME: add Token() to ast.Node interface
		return &object.Error{Message: fmt.Sprintf("unknown node: [%T] %v", node, node)}
	}
}

func evalProgram(stmts []ast.Statement, env *object.Environment) object.Object {
	var result object.Object = VOID

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
	var result object.Object = VOID

	for _, stmt := range stmts {
		result = Eval(stmt, env)
		if result.Type() == object.RETURN_OBJ || result.Type() == object.ERROR_OBJ {
			return result
		}
	}

	return result
}

func evalPrefixExpression(op string, right object.Object, token *token.Token) object.Object {
	switch op {
	case "!":
		return evalBangExpression(right, token)
	case "-":
		return evalMinusPrefixExpression(right, token)
	default:
		return newError(token, fmt.Sprintf("unknown operator: %s%s", op, right.Type()))
	}
}

func evalBangExpression(obj object.Object, token *token.Token) object.Object {
	switch obj {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case VOID:
		return TRUE
	default:
		return newError(token, fmt.Sprintf("unknown operator: !%s", obj.Type()))
	}
}

func evalMinusPrefixExpression(obj object.Object, token *token.Token) object.Object {
	if obj.Type() != object.INTEGER_OBJ {
		return newError(token, fmt.Sprintf("unknown operator: -%s", obj.Type()))
	}

	value := obj.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func evalInfixExpression(op string, left object.Object, right object.Object, token *token.Token) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerExpression(op, left, right, token)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringExpression(op, left, right, token)
	case left.Type() != right.Type():
		return newError(token, fmt.Sprintf("type mismatch: %s %s %s", left.Type(), op, right.Type()))
	case op == "==":
		return newBoolean(left == right)
	case op == "!=":
		return newBoolean(left != right)
	default:
		return newError(token, fmt.Sprintf("unknown operator: %s %s %s", left.Type(), op, right.Type()))
	}
}

func evalIntegerExpression(op string, left object.Object, right object.Object, token *token.Token) object.Object {
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
		return newBoolean(leftVal < rightVal)
	case ">":
		return newBoolean(leftVal > rightVal)
	case "<=":
		return newBoolean(leftVal <= rightVal)
	case ">=":
		return newBoolean(leftVal >= rightVal)
	case "==":
		return newBoolean(leftVal == rightVal)
	case "!=":
		return newBoolean(leftVal != rightVal)
	default:
		return newError(token, fmt.Sprintf("unknown operator: %s %s %s", left.Type(), op, right.Type()))
	}
}

func evalStringExpression(op string, left object.Object, right object.Object, token *token.Token) object.Object {
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value

	switch op {
	case "+":
		return &object.String{Value: leftVal + rightVal}
	case "==":
		return newBoolean(leftVal == rightVal)
	case "!=":
		return newBoolean(leftVal != rightVal)
	default:
		return newError(token, fmt.Sprintf("unknown operator: %s %s %s", left.Type(), op, right.Type()))
	}
}

func evalIfExpression(condition object.Object, consequence *ast.BlockStatement, alternative *ast.BlockStatement, env *object.Environment, token *token.Token) object.Object {
	if condition == TRUE {
		return Eval(consequence, env)
	} else if condition != FALSE {
		return newError(token, fmt.Sprintf("non-boolean condition in IF expression: %s", condition.Type()))

	} else if alternative != nil {
		return Eval(alternative, env)
	} else {
		return VOID
	}
}

func evalExpressions(expressions []ast.Expression, env *object.Environment) []object.Object {
	res := []object.Object{}

	for _, ex := range expressions {
		val := Eval(ex, env)
		if val.Type() == object.ERROR_OBJ {
			return []object.Object{val}
		} else {
			res = append(res, val)
		}
	}

	return res
}

func applyFunction(fn object.Object, args []object.Object, token *token.Token) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		extEnv := extendFnEnv(fn, args)
		evaluated := Eval(fn.Body, extEnv)
		if val, ok := evaluated.(*object.Return); ok {
			return val.Value
		}
		return evaluated
	case *object.Builtin:
		return fn.Fn(token, args...)
	default:
		return newError(token, fmt.Sprintf("not a function: %s", fn.Type()))
	}
}

func extendFnEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	for i, param := range fn.Parameters {
		env.Set(param.Value, args[i])
	}

	return env
}

func evalAccessExpression(arr object.Object, key object.Object, token *token.Token) object.Object {
	switch {
	case arr.Type() == object.ARRAY_OBJ && key.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(arr, key, token)
	default:
		return newError(token, fmt.Sprintf("invalid argument: %s[%s]", arr.Type(), key.Type()))
	}
}

func evalArrayIndexExpression(arr object.Object, key object.Object, token *token.Token) object.Object {
	items := arr.(*object.ArrayLiteral).Items
	idx := key.(*object.Integer).Value

	length := int64(len(items))

	switch {
	case idx < 0 && length+idx > 0:
		return items[length+idx]
	case idx >= 0 && idx < length:
		return items[idx]
	default:
		return newError(token, fmt.Sprintf("invalid argument: index %d out of bounds", idx))
	}
}

func evalAssignExpression(node *ast.AssignExpression, env *object.Environment) object.Object {
	_, declared := env.Get(node.Name.Value)

	if node.Token.Type == token.ASSIGN {
		if !declared {
			return newError(&node.Token, fmt.Sprintf("assigning to undeclared variable: %s", node.Name.Value))
		}
		if env.IsConst(node.Name.Value) {
			return newError(&node.Token, fmt.Sprintf("assigning to const: %s", node.Name.Value))
		}
	} else {
		if declared {
			return newError(&node.Token, fmt.Sprintf("identifier already declared: %s", node.Name.Value))
		}
	}

	// TODO: switch to lazy evaluation?
	val := Eval(node.Value, env)
	if val.Type() == object.ERROR_OBJ {
		return val
	}

	if node.Token.Type == token.CONST {
		return env.SetConst(node.Name.Value, val)
	} else {
		return env.Set(node.Name.Value, val)
	}
}
