package repl

import (
	"bufio"
	"fmt"
	"io"
	"monkey/code"
	"monkey/compiler"
	"monkey/lexer"
	"monkey/parser"
	"monkey/vm"
	"os"
)

const PROMPT = "> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	instruction := code.Make(code.Opconst, 65534)
	for i, b := range instruction {
		fmt.Printf("%dth byte: %x\n", i, b)
	}
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
		comp := compiler.New()
		comp.Compile(prog)
		v := vm.New(comp.Bytecode())
		v.Run()

		// print error messages
		for _, err := range p.Errors() {
			fmt.Printf(err)
		}
		io.WriteString(out, v.StackTop().Inspect())
		io.WriteString(out, "\n")
	}
}
