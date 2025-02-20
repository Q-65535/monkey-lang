package compiler

import (
	"fmt"
	"monkey/ast"
	"monkey/code"
	"monkey/object"
)

type EmittedInstruction struct {
	Opcode   code.Opcode
	Position int
}

type Compiler struct {
	instructions        code.Instructions
	constants           []object.Object
	globals             []object.Object
	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
	symbolTable         *SymbolTable
}

var symbol_table = map[string]int{}

func New() *Compiler {
	return &Compiler{
		instructions:        code.Instructions{},
		constants:           []object.Object{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
		symbolTable:         NewSymbolTable(),
	}
}

func NewWithState(st *SymbolTable, constants []object.Object) *Compiler {
	c := New()
	// @Optimize: this copying is not efficient, we can use address instead
	c.constants = constants
	c.symbolTable = st
	return c
}

func (c *Compiler) Compile(node ast.Node, depth int) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			err := c.Compile(s, depth)
			if err != nil {
				return err
			}
		}
	case *ast.BlockStatement:
		for _, s := range node.Statements {
			err := c.Compile(s, depth)
			if err != nil {
				return err
			}
		}
	case *ast.FunctionLiteral:
		c_func := NewWithState(c.symbolTable, c.constants)
		c_func.symbolTable = NewSymbolTableWithUpper(c.symbolTable)
		err := c_func.Compile(node.Body, depth+1)
		if err != nil {
			return err
		}
		// @Problem: what if the last instruction is a let statement?
		if c_func.lastInstructionIsPop() {
			c_func.removeLastPop()
			c_func.emit(code.OpReturnValue)
		}
		if !c_func.lastInstructionIsReturnValue() {
			c_func.emit(code.OpReturn)
		}
		// constants are moved back
		// @Optimize: this copying is not efficient, we can use address instead
		c.constants = c_func.constants
		compiledFunc := &object.CompiledFunction{Instructions: c_func.instructions}
		index := c.addConstant(compiledFunc)
		c.emit(code.Opconst, index)
	case *ast.ReturnStatement:
		err := c.Compile(node.ReturnValue, depth)
		if err != nil {
			return err
		}
		c.emit(code.OpReturnValue)
	case *ast.CallExpression:
		err := c.Compile(node.Function, depth)
		if err != nil {
			return err
		}
		c.emit(code.OpCall)
	case *ast.LetStatement:
		err := c.Compile(node.Value, depth)
		if err != nil {
			return err
		}
		_, ok := c.symbolTable.Resolve(node.Name.Value)
		if ok {
			return fmt.Errorf("%s is already defined", node.Name.Value)
		}
		symbol := c.symbolTable.Define(node.Name.Value)
		var opcode code.Opcode
		if symbol.Scope == GlobalScope {
			opcode = code.OpSetGlobal
		} else {
			opcode = code.OpSetLocal
		}
		c.emit(opcode, symbol.Index)
	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		var opcode code.Opcode
		if !ok {
			return fmt.Errorf("undefined variable: %s", node.Value)
		}
		if symbol.Scope == GlobalScope {
			opcode = code.OpGetGlobal
		} else {
			opcode = code.OpGetLocal
		}
		c.emit(opcode, symbol.Index)
	// @TODO: we need an assignment statement
	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression, depth)
		if err != nil {
			return err
		}
		c.emit(code.OpPop)
	case *ast.PrefixExpression:
		err := c.Compile(node.Right, depth)
		if err != nil {
			return err
		}
		switch node.Operator {
		case "!":
			c.emit(code.OpBang)
		case "-":
			c.emit(code.OpMinus)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}
	case *ast.InfixExpression:
		err := c.Compile(node.Left, depth)
		if err != nil {
			return err
		}
		err = c.Compile(node.Right, depth)
		if err != nil {
			return err
		}
		switch node.Operator {
		case "+":
			c.emit(code.OpAdd)
		case "-":
			c.emit(code.OpSub)
		case "*":
			c.emit(code.OpMul)
		case "/":
			c.emit(code.OpDiv)
		case ">":
			c.emit(code.OpGreaterThan)
		case "<":
			c.emit(code.OpLessThan)
		case "==":
			c.emit(code.OpEqual)
		case "!=":
			c.emit(code.OpNotEqual)
		default:
			return fmt.Errorf("operator not support: %s", node.Operator)
		}
	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		integer_index := c.addConstant(integer)
		c.emit(code.Opconst, integer_index)
	case *ast.BooleanLiteral:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}
	case *ast.StringLiteral:
		str := &object.String{Value: node.Value}
		str_index := c.addConstant(str)
		c.emit(code.Opconst, str_index)
	case *ast.IfExpression:
		err := c.Compile(node.Condition, depth)
		if err != nil {
			return err
		}
		ins_jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 999)
		var ins_jumpOverAltPos int
		// consequence
		err = c.Compile(node.Consequence, depth)
		if err != nil {
			return err
		}
		if c.lastInstructionIsPop() {
			c.removeLastPop()
		}
		ins_jumpOverAltPos = c.emit(code.OpJump, 999)
		// now modify the jump position
		afterConsequencePos := len(c.instructions)
		c.changeOperand(ins_jumpNotTruthyPos, afterConsequencePos)
		if node.Altenative == nil || len(node.Altenative.Statements) == 0 {
			c.emit(code.OpNull)
		} else {
			err = c.Compile(node.Altenative, depth)
			if err != nil {
				return err
			}
			if c.lastInstructionIsPop() {
				c.removeLastPop()
			}
		}
		// now modify the jump position
		afterAltenativePos := len(c.instructions)
		c.changeOperand(ins_jumpOverAltPos, afterAltenativePos)
	case *ast.ArrayLiteral:
		for _, el := range node.Elements {
			err := c.Compile(el, depth)
			if err != nil {
				return err
			}
		}
		c.emit(code.OpArray, len(node.Elements))
	case *ast.ArrayAccessExpression:
		err := c.Compile(node.Array, depth)
		if err != nil {
			return err
		}
		err = c.Compile(node.Index, depth)
		if err != nil {
			return err
		}
		c.emit(code.OpIndex)
	}
	return nil
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)
	c.previousInstruction = c.lastInstruction
	c.lastInstruction = EmittedInstruction{Opcode: op, Position: pos}
	return pos
}

func (c *Compiler) addInstruction(ins []byte) int {
	newInstructionPos := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	return newInstructionPos
}

func (c *Compiler) lastInstructionIsPop() bool {
	return c.lastInstruction.Opcode == code.OpPop
}

func (c *Compiler) lastInstructionIsReturnValue() bool {
	return c.lastInstruction.Opcode == code.OpReturnValue
}

func (c *Compiler) removeLastPop() {
	c.instructions = c.instructions[:c.lastInstruction.Position]
	c.lastInstruction = c.previousInstruction
}

func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	for i := 0; i < len(newInstruction); i++ {
		c.instructions[pos+i] = newInstruction[i]
	}
}

func (c *Compiler) changeOperand(opPos int, operand int) {
	op := code.Opcode(c.instructions[opPos])
	newInstruction := code.Make(op, operand)
	c.replaceInstruction(opPos, newInstruction)
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}
