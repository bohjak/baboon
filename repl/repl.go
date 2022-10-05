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

	fmt.Println("[RPPL Mode]")

	for {
		fmt.Printf(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		switch line {
		case "lex":
			mode = LEXER
			fmt.Println("[RLPL Mode]")
			continue
		case "parse":
			mode = PARSER
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
