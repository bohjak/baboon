package cli

import (
	"fmt"
	"os"

	"baboon/evaluator"
	"baboon/lexer"
	"baboon/object"
	"baboon/parser"
	"baboon/token"
)

type opts struct {
	name  string
	file  string
	str   string
	eval  bool
	parse bool
	lex   bool
}

func (o *opts) printHelp(code int) {
	if code != 0 {
		fmt.Fprintln(os.Stderr, "incorrect usage")
	}
	fmt.Printf("%s [-e | -p | -l] [-s <PROGRAM> | <FILE>]\n", o.name)
	fmt.Println()
	fmt.Println("FLAGS:")
	fmt.Println("\t-e: evaluate input program and print result")
	fmt.Println("\t-p: parse input program and print prettified AST")
	fmt.Println("\t-l: lex input program and print tokens")
	fmt.Println()
	fmt.Println("ARGUMENTS:")
	fmt.Println("\t-s <PROGRAM>:\tread program from argument")
	fmt.Println("\t<FILE>:\t\tread program from file")
	os.Exit(code)
}

func parseArgs(args []string) *opts {
	opts := opts{name: args[0]}

	argc := len(args)
	for i := 1; i < argc; i++ {
		arg := args[i]
		switch arg {
		case "-h":
			opts.printHelp(0)
		case "-e":
			opts.eval = true
		case "-p":
			opts.parse = true
		case "-l":
			opts.lex = true
		case "-s":
			if i+1 >= argc || args[i+1][0] == '-' {
				opts.printHelp(1)
			}
			// TODO: allow for multiple execution units
			opts.str = args[i+1]
			i++
		default:
			opts.file = arg
		}
	}

	if len(opts.str) == len(opts.file) {
		opts.printHelp(1)
	}

	return &opts
}

func Run(args []string) {
	var input string
	opts := parseArgs(args)
	if len(opts.str) != 0 {
		input = opts.str
	} else {
		data, err := os.ReadFile(opts.file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "couldn't read file %q\n", opts.file)
			panic(err)
		}
		input = string(data)
	}

	l := lexer.New(input)
	if opts.lex {
		for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
			fmt.Printf("[%d:%d]\t%8s: %q\n", tok.Line, tok.Column, tok.Type, tok.Literal)
		}
	} else if opts.parse {

		p := parser.New(l)
		prog := p.ParseProgram()
		printErrors(p.Errors())
		fmt.Println(prog)
	} else {
		p := parser.New(l)
		prog := p.ParseProgram()
		env := object.NewEnvironment()
		fmt.Println(evaluator.Eval(prog, env).Inspect())
	}
}

func printErrors(errors []string) {
	if len(errors) > 0 {
		for _, err := range errors {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}
