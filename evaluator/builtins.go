package evaluator

import (
	"fmt"

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

			s, ok := args[0].(*object.String)
			if !ok {
				return &object.Error{Message: fmt.Sprintf("invalid argument: len(%s)", args[0].Type()), Line: token.Line, Column: token.Column}
			}

			return &object.Integer{Value: int64(len(s.Value))}
		},
	},
}
