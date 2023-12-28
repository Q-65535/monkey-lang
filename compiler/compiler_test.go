package compiler

import (
	"fmt"
	"monkey/ast"
	"monkey/code"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"testing"
)

type compilerTestCase struct {
	input                string
	expectedConstants    []interface{}
	expectedInstructions []code.Instructions
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "1+2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.Opconst, 0),
				code.Make(code.Opconst, 1),
				code.Make(code.OpAdd),
			},
		},
	}
	runCompilerTests(t, tests)
}

func runCompilerTests(t *testing.T, tests []compilerTestCase) {
	t.Helper()

	for _, tt := range tests {
		program := parse(tt.input)
		compiler := New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error %s", err)
		}
		bytecode := compiler.Bytecode()
		err = testInstructions(tt.expectedInstructions, bytecode.Instructions)
		if err != nil {
			t.Fatalf("test instructions failed: %s", err)
		}
		err = testConstants(t, tt.expectedConstants, bytecode.Constants)
		if err != nil {
			t.Fatalf("test constants failed: %s", err)
		}
	}
}

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}

func testInstructions(expected []code.Instructions, actual code.Instructions) error {
	concatted := concatInstructions(expected)
	if len(actual) != len(concatted) {
		return fmt.Errorf("wrong instructions length: want=%q, got=%q", concatted, actual)
	}
	for i, ins := range concatted {
		if actual[i] != ins {
			return fmt.Errorf("wrong instructions at %d: want=%q, got=%q", i, concatted, actual)
		}
	}
	return nil
}

func concatInstructions(instructionsArr []code.Instructions) code.Instructions {
	var res code.Instructions
	for _, instructions := range instructionsArr {
		res = append(res, instructions...)
	}
	return res
}

func testConstants(t *testing.T, expected []interface{}, actual []object.Object) error {
	if len(actual) != len(expected) {
		return fmt.Errorf("wrong instructions length: want=%q, got=%q", expected, actual)
	}
	for i, cons := range expected {
		switch cons := cons.(type) {
		case int:
			err := testIntegerObject(int64(cons), actual[i])
			if err != nil {
				return fmt.Errorf("%dth constant testIntegerObject failed: %s", i, err)
			}
		}
	}
	return nil
}

func testIntegerObject(expected int64, actual object.Object) error {
	result, ok := actual.(*object.Integer)
	if !ok {
		return fmt.Errorf("object is not Integer. got=%T (%+v)", actual, actual)
	}
	if result.Value != expected {
		return fmt.Errorf("not expected integer number: expect=%d, got=%d", expected, result.Value)
	}
	return nil
}
