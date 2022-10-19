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
	target := args[1]
	data, err := os.ReadFile(target)
	if err != nil {
		log.Fatalf("Couldn't read file: %v", err)
	}

	l := lexer.New(string(data))
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
