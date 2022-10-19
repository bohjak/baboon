package cli

import (
	"fmt"
	"log"
	"os"

	"baboon/evaluator"
	"baboon/lexer"
	"baboon/object"
	"baboon/parser"
)

func Run(args []string) {
	var input string
	// TODO: create a simple parser to allow for execution flags (lex, parse, eval)
	if args[1] == "-s" {
		// TODO: allow for multiple execution units
		if len(args) != 3 {
			log.Fatalf("%s -s [PROGRAM]", args[0])
		}

		input = args[2]
	} else {
		target := args[1]
		data, err := os.ReadFile(target)
		if err != nil {
			log.Fatalf("Couldn't read file: %v", err)
		}
		input = string(data)
	}

	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if len(p.Errors()) > 0 {
		for _, err := range p.Errors() {
			fmt.Fprintln(os.Stderr, err)
		}
		panic(2)
	}
	env := object.NewEnvironment()
	fmt.Println(evaluator.Eval(prog, env).Inspect())
}
