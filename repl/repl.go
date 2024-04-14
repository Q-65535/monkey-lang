package repl

import (
	"bufio"
	"fmt"
	"io"
	"monkey/code"
	"monkey/compiler"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"monkey/vm"
	"os"
)

const PROMPT = "> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	constants := []object.Object{}
	globals := make([]object.Object, vm.GlobalSize)
	symbolTable := compiler.NewSymbolTable()

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
		comp := compiler.NewWithState(symbolTable, constants)
		comp.Compile(prog)
		constants = comp.Bytecode().Constants
		v := vm.NewWithGlobalsStore(comp.Bytecode(), globals)
		v.Run()

		// print error messages
		for _, err := range p.Errors() {
			fmt.Printf(err)
		}
		io.WriteString(out, v.LastPopped().Inspect())
		io.WriteString(out, "\n")
	}
}
