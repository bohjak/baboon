package evaluator

import (
	"fmt"
	"strings"

	"baboon/object"
	"baboon/token"
)

var builtins = map[string]*object.Builtin{
	"len": {
		Fn: func(token token.Token, args ...object.Object) object.Object {
			if len(args) == 0 {
				return &object.Error{Message: "not enough arguments for len: expected 1, found 0", Line: token.Line, Column: token.Column}
			}
			if len(args) > 1 {
				return &object.Error{Message: fmt.Sprintf("too many arguments for len: expected 1, found %d", len(args)), Line: token.Line, Column: token.Column}
			}

			switch arg := args[0].(type) {
			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}
			case *object.ArrayLiteral:
				// TODO: use length field
				return &object.Integer{Value: int64(len(arg.Items))}
			default:
				return &object.Error{Message: fmt.Sprintf("invalid argument: len(%s)", arg.Type()), Line: token.Line, Column: token.Column}
			}
		},
	},

	"print": {
		Fn: func(_ token.Token, args ...object.Object) object.Object {
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
