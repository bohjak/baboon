package repl

import (
	"bufio"
	"fmt"
	"io"

	"baboon/evaluator"
	"baboon/lexer"
	"baboon/object"
	"baboon/parser"
	"baboon/token"
)

const PROMPT = ">> "

const (
	_ int = iota
	LEX
	PARSE
	EVAL
)

var mode int = EVAL

// TODO: implement history, persistent state, real-time eval, pretty printing?
func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()

	fmt.Println("[REPL Mode]")

	for {
		fmt.Printf("%s", PROMPT)

		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		switch line {
		case "lex":
			mode = LEX
			fmt.Println("[RLPL Mode]")
			continue
		case "parse":
			mode = PARSE
			fmt.Println("[RPPL Mode]")
			continue
		case "eval":
			mode = EVAL
			fmt.Println("[REPL Mode]")
			continue
		case "exit":
			fmt.Println("[Exit]")
			return
		}

		l := lexer.New(line)
		switch mode {
		case LEX:
			for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
				fmt.Printf("%+v\n", tok)
			}
		case PARSE:
			p := parser.New(l)
			program := p.ParseProgram()
			printErrors(p.Errors())
			fmt.Println(program.String())
		case EVAL:
			p := parser.New(l)
			program := p.ParseProgram()
			printErrors(p.Errors())
			result := evaluator.Eval(program, env)
			if result == nil {
				fmt.Println("could not evaluate")
			} else {
				fmt.Println(result.Inspect())
			}
		}
	}
}

func printErrors(errors []string) {
	if len(errors) > 0 {
		for _, error := range errors {
			fmt.Println(error)
		}
	}
}
