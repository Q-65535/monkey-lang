package repl

import (
	"bufio"
	"fmt"
	"io"
	"monkey/lexer"
	"monkey/parser"
	"os"
)

const PROMPT = "> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
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
		io.WriteString(out, prog.String())
		io.WriteString(out, "\n")
	}
}
