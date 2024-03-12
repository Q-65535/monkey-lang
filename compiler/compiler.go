package compiler

import (
	"fmt"
	"monkey/ast"
	"monkey/code"
	"monkey/object"
)

type Compiler struct {
	instructions code.Instructions
	constants    []object.Object
}

func New() *Compiler {
	return &Compiler{
		instructions: code.Instructions{},
		constants:    []object.Object{},
	}
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
	case *ast.BlockStatement:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
		// @Note: this implementation is different from what the book says.
		_, ok := node.Expression.(*ast.IfExpression)
		if !ok {
			c.emit(code.OpPop)
		}
	case *ast.InfixExpression:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}
		err = c.Compile(node.Right)
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
	case *ast.IfExpression:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}
		c.emit(code.OpJumpNotTruthy, 999)
		oprandPos := len(c.instructions) - 2 // oprand width is 2
		// consequence
		c.Compile(node.Consequence)
		// jump to execute OpPop
		// @Problem: what if there is no pop instruction emitted when compiling the consequence?
		popPos := len(c.instructions) - 1
		// now modify the jump position
		c.instructions[oprandPos] = byte((popPos >> 8) & 0xff)
		c.instructions[oprandPos+1] = byte(popPos & 0xff)

		if node.Altenative != nil {
			// jump over altenative
			c.emit(code.OpJump, 999)
			oprandPos = len(c.instructions) - 2 // oprand width is 2
			// altenative
			c.Compile(node.Altenative)
			popPos = len(c.instructions)
			// now modify the jump position
			c.instructions[oprandPos] = byte((popPos >> 8) & 0xff)
			c.instructions[oprandPos+1] = byte(popPos & 0xff)
		}
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
	return pos
}

func (c *Compiler) addInstruction(ins []byte) int {
	newInstructionPos := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	return newInstructionPos
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
