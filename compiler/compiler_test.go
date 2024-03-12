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
	expectedConstants    []any
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
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1-2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.Opconst, 0),
				code.Make(code.Opconst, 1),
				code.Make(code.OpSub),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1*2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.Opconst, 0),
				code.Make(code.Opconst, 1),
				code.Make(code.OpMul),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1/2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.Opconst, 0),
				code.Make(code.Opconst, 1),
				code.Make(code.OpDiv),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "true",
			expectedConstants: []any{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "false",
			expectedConstants: []any{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpFalse),
				code.Make(code.OpPop),
			},
		},
		// comparison operators
		{
			input:             "1 > 2",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.Opconst, 0),
				code.Make(code.Opconst, 1),
				code.Make(code.OpGreaterThan),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 < 2",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.Opconst, 0),
				code.Make(code.Opconst, 1),
				code.Make(code.OpLessThan),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 == 2",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.Opconst, 0),
				code.Make(code.Opconst, 1),
				code.Make(code.OpEqual),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 != 2",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.Opconst, 0),
				code.Make(code.Opconst, 1),
				code.Make(code.OpNotEqual),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "true != false",
			expectedConstants: []any{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpFalse),
				code.Make(code.OpNotEqual),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestConditionals(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "if (true) {10;} 3333;",
			expectedConstants: []any{10, 3333},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),             // 0000
				code.Make(code.OpJumpNotTruthy, 7), // 0001
				code.Make(code.Opconst, 0),         // 0004
				code.Make(code.OpPop),              // 0007
				code.Make(code.Opconst, 1),         // 0008
				code.Make(code.OpPop),              // 000B
			},
		},
		{
			input:             "if (true) {10} else {20}; 3333;",
			expectedConstants: []any{10, 3333},
			expectedInstructions: []code.Instructions{
				// 0000
				code.Make(code.OpTrue),
				// 0001
				code.Make(code.OpJumpNotTruthy, 10),
				// 0004
				code.Make(code.Opconst, 0),
				// 0007
				code.Make(code.OpJump, 13),
				// 0010
				code.Make(code.Opconst, 1),
				// 0013
				code.Make(code.OpPop),
				// 0014
				code.Make(code.Opconst, 2),
				// 0017
				code.Make(code.OpPop),
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
		return fmt.Errorf("wrong instructions length:\n want:\n%s \n got:\n%s", concatted, actual)
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
