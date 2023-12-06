package repl

import (
	"bufio"
	"fmt"
	"io"
	"monkey/evaluator"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"os"
)

const PROMPT = "> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()
	for true {
		fmt.Printf(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}
		line := scanner.Text()
		if line == "quit" || line == "exit" {
			os.Exit(0)
		}
		l := lexer.New(line)
		p := parser.New(l)
		prog := p.ParseProgram()
		res := evaluator.Eval(prog, env)
		// print error messages
		for _, err := range p.Errors() {
			fmt.Printf(err)
		}
		io.WriteString(out, res.Inspect())
		io.WriteString(out, "\n")
	}
}
