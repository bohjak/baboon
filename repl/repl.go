package repl

import (
	"bufio"
	"fmt"
	"io"

	"baboon/lexer"
	"baboon/parser"
	"baboon/token"
)

const PROMPT = ">> "

const (
	_ int = iota
	LEXER
	PARSER
	EVAL
)

var mode int = PARSER

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Printf(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		switch line {
		case "lexer":
			mode = LEXER
			fmt.Println("mode: LEXER")
			continue
		case "parser":
			mode = PARSER
			fmt.Println("mode: PARSER")
			continue
		case "exit":
			return
		}

		l := lexer.New(line)
		switch mode {
		case LEXER:
			for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
				fmt.Printf("%+v\n", tok)
			}
		case PARSER:
			p := parser.New(l)
			program := p.ParseProgram()
			if len(p.Errors()) > 0 {
				for _, error := range p.Errors() {
					fmt.Println(error)
				}
			}
			fmt.Println(program.String())
		}
	}
}
