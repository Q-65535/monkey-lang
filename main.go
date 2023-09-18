package main

import (
	"fmt"
	"monkey/lexer"
	"monkey/token"
	"monkey/repl"
	"os"
	// "os/user"
)

func main() {
	l := lexer.New("hello+world;")
	tok := l.NextToken()
	for tok.Type != token.EOF {
		fmt.Printf("%+v\n", tok)
		tok = l.NextToken()
	}
	fmt.Printf("hello world!\n")
	repl.Start(os.Stdin, os.Stdout)
}
