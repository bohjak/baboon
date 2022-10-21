package evaluator

import (
	"fmt"
	"strings"

	"baboon/object"
	"baboon/token"
)

var builtins = map[string]*object.Builtin{
	"len": {
		Fn: func(token *token.Token, args ...object.Object) object.Object {
			if len(args) == 0 {
				return newError(token, "not enough arguments for len: expected 1, found 0")
			}
			if len(args) > 1 {
				return newError(token, fmt.Sprintf("too many arguments for len: expected 1, found %d", len(args)))
			}

			switch arg := args[0].(type) {
			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}
			case *object.ArrayLiteral:
				// TODO: use length field
				return &object.Integer{Value: int64(len(arg.Items))}
			default:
				return newError(token, fmt.Sprintf("invalid argument: len(%s)", arg.Type()))
			}
		},
	},

	"print": {
		Fn: func(_ *token.Token, args ...object.Object) object.Object {
			out := []string{}
			for _, arg := range args {
				out = append(out, arg.Inspect())
			}
			fmt.Println(strings.Join(out, " "))
			// TODO: add void?
			return VOID
		},
	},

	"append": {
		Fn: func(token *token.Token, args ...object.Object) object.Object {
			if len(args) < 2 {
				return newError(token, fmt.Sprintf("not enough arguments for append: expected 2, found %d", len(args)))
			}

			items := args[1:]
			switch arr := args[0].(type) {
			case *object.ArrayLiteral:
				return &object.ArrayLiteral{Items: append(arr.Items, items...)}
			default:
				return newError(token, fmt.Sprintf("invalid argument: append(%s, ...items)", arr.Type()))
			}
		},
	},

	"first": {
		Fn: func(token *token.Token, args ...object.Object) object.Object {
			if len(args) == 0 {
				return newError(token, "not enough arguments for first: expected 1, found 0")
			}
			if len(args) > 1 {
				return newError(token, fmt.Sprintf("too many arguments for first: expected 1, found %d", len(args)))
			}

			switch arg := args[0].(type) {
			case *object.String:
				if len(arg.Value) == 0 {
					return newError(token, "invalid argument for first: index 0 out of bounds")
				}
				return &object.String{Value: string(arg.Value[0])}
			case *object.ArrayLiteral:
				if len(arg.Items) == 0 {
					return newError(token, "invalid argument for first: index 0 out of bounds")
				}
				return arg.Items[0]
			default:
				return newError(token, fmt.Sprintf("invalid argument: first(%s)", arg.Type()))
			}
		},
	},

	"tail": {
		Fn: func(token *token.Token, args ...object.Object) object.Object {
			if len(args) == 0 {
				return newError(token, "not enough arguments for tail: expected 1, found 0")
			}
			if len(args) > 1 {
				return newError(token, fmt.Sprintf("too many arguments for tail: expected 1, found %d", len(args)))
			}

			switch arg := args[0].(type) {
			case *object.String:
				if len(arg.Value) == 0 {
					return arg
				}
				return &object.String{Value: arg.Value[1:]}
			case *object.ArrayLiteral:
				if len(arg.Items) == 0 {
					return arg
				}
				return &object.ArrayLiteral{Items: arg.Items[1:]}
			default:
				return newError(token, fmt.Sprintf("invalid argument: tail(%s)", arg.Type()))
			}
		},
	},
}
