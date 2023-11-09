package main

import (
	"fmt"
	"monkey/lexer"
	"monkey/repl"
	"monkey/token"
	"os"
	// "os/user"
)

func main() {
	l := lexer.New("let a = 7;")
	tok := l.NextToken()
	for tok.Type != token.EOF {
		fmt.Printf("%+v\n", tok)
		tok = l.NextToken()
	}
	fmt.Printf("start repl...!\n")
	repl.Start(os.Stdin, os.Stdout)
}
