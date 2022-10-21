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
			return NULL
		},
	},
}
