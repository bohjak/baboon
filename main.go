package main

import (
	"baboon/cli"
	"baboon/repl"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		cli.Run(os.Args)
	} else {
		fmt.Println("Baboon Interactive Environment")
		repl.Start(os.Stdin, os.Stdout)
	}
}
