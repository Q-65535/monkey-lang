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

func TestArrayLiterals(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "[]",
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpArray, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "[1, 2, 3]",
			expectedConstants: []interface{}{1, 2, 3},
			expectedInstructions: []code.Instructions{
				code.Make(code.Opconst, 0),
				code.Make(code.Opconst, 1),
				code.Make(code.Opconst, 2),
				code.Make(code.OpArray, 3),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "[1 + 2, 3 - 4, 5 * 6]",
			expectedConstants: []interface{}{1, 2, 3, 4, 5, 6},
			expectedInstructions: []code.Instructions{
				code.Make(code.Opconst, 0),
				code.Make(code.Opconst, 1),
				code.Make(code.OpAdd),
				code.Make(code.Opconst, 2),
				code.Make(code.Opconst, 3),
				code.Make(code.OpSub),
				code.Make(code.Opconst, 4),
				code.Make(code.Opconst, 5),
				code.Make(code.OpMul),
				code.Make(code.OpArray, 3),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestStringExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             `"monkey"`,
			expectedConstants: []interface{}{"monkey"},
			expectedInstructions: []code.Instructions{
				code.Make(code.Opconst, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             `"mon" + "key"`,
			expectedConstants: []interface{}{"mon", "key"},
			expectedInstructions: []code.Instructions{
				code.Make(code.Opconst, 0),
				code.Make(code.Opconst, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
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
				code.Make(code.OpTrue),              // 0000
				code.Make(code.OpJumpNotTruthy, 10), // 0001
				code.Make(code.Opconst, 0),          // 0004
				code.Make(code.OpJump, 11),          // 0007
				code.Make(code.OpNull),              // 000A
				code.Make(code.OpPop),               // 000B
				code.Make(code.Opconst, 1),          // 000C
				code.Make(code.OpPop),               // 000F
			},
		},
		{
			input:             "if (true) {10} else {20}; 3333;",
			expectedConstants: []any{10, 20, 3333},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),              // 0000
				code.Make(code.OpJumpNotTruthy, 10), // 0001
				code.Make(code.Opconst, 0),          // 0004
				code.Make(code.OpJump, 13),          // 0007
				code.Make(code.Opconst, 1),          // 0010
				code.Make(code.OpPop),               // 0013
				code.Make(code.Opconst, 2),          // 0014
				code.Make(code.OpPop),               // 0017
			},
		},
		// test for null case
		{
			input:             "if (false) {10;};",
			expectedConstants: []any{10},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpFalse),             // 0000
				code.Make(code.OpJumpNotTruthy, 10), // 0001
				code.Make(code.Opconst, 0),          // 0004
				code.Make(code.OpJump, 11),          // 0007
				code.Make(code.OpNull),              // 000A
				code.Make(code.OpPop),               // 000B
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestGlobalLetStatements(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "let a = 1; let b = 2;",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.Opconst, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.Opconst, 1),
				code.Make(code.OpSetGlobal, 1),
			},
		},
		{
			input:             "let a = 1; a;",
			expectedConstants: []any{1},
			expectedInstructions: []code.Instructions{
				code.Make(code.Opconst, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `let a = 1;
					let b = a;
					b;`,
			expectedConstants: []any{1},
			expectedInstructions: []code.Instructions{
				code.Make(code.Opconst, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpGetGlobal, 1),
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
			return fmt.Errorf("wrong instructions at %d\n: want:\n%s \n got:\n%s", i, concatted, actual)
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
		return fmt.Errorf("wrong instructions length: want=%s\n, got=%s\n", expected, actual)
	}
	for i, cons := range expected {
		switch cons := cons.(type) {
		case int:
			err := testIntegerObject(int64(cons), actual[i])
			if err != nil {
				return fmt.Errorf("%dth constant testIntegerObject failed: %s", i, err)
			}
		case string:
			err := testStringObject(string(cons), actual[i])
			if err != nil {
				return fmt.Errorf("%dth constant testStringObject failed: %s", i, err)
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

func testStringObject(expected string, actual object.Object) error {
	result, ok := actual.(*object.String)
	if !ok {
		return fmt.Errorf("object is not String. got=%T (%+v)",
			actual, actual)
	}
	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%q, want=%q",
			result.Value, expected)
	}
	return nil
}
