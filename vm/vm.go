package vm

import (
	"fmt"
	"monkey/code"
	"monkey/compiler"
	"monkey/object"
)

const StackSize = 2048

type VM struct {
	constants    []object.Object
	instructions code.Instructions
	stack        []object.Object
	sp           int
	lastPopped   object.Object
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		constants:    bytecode.Constants,
		instructions: bytecode.Instructions,
		stack:        make([]object.Object, StackSize),
		sp:           0,
	}
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return &object.Null{}
	} else {
		return vm.stack[vm.sp-1]
	}
}

func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		op := code.Opcode(vm.instructions[ip])
		switch op {
		case code.Opconst:
			constIndex := code.ReadUint16(vm.instructions[ip+1:])
			integer := vm.constants[constIndex]
			if integer == nil {
				return fmt.Errorf("const not found, problem index: %d", constIndex)
			}
			ip += 2
			err := vm.push(integer)
			if err != nil {
				return err
			}
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			right := vm.pop()
			left := vm.pop()
			if left.Type() != object.INTEGER_OBJ || right.Type() != object.INTEGER_OBJ {
				return fmt.Errorf("unsupported types for binary operation: %s %s", left.Type(), right.Type())
			}
			leftVal := left.(*object.Integer).Value
			rightVal := right.(*object.Integer).Value
			var res int64
			switch op {
			case code.OpAdd:
				res = leftVal + rightVal
			case code.OpSub:
				res = leftVal - rightVal
			case code.OpMul:
				res = leftVal * rightVal
			case code.OpDiv:
				res = leftVal / rightVal
			}
			vm.push(&object.Integer{Value: res})
		case code.OpPop:
			vm.pop()
		default:
			return fmt.Errorf("unknown operator: %d", op)
		}
	}
	return nil
}

func (vm *VM) push(obj object.Object) error {
	if obj == nil {
		return fmt.Errorf("vm push error: the pushed object is nil")
	}
	vm.stack[vm.sp] = obj
	vm.sp++
	return nil
}

func (vm *VM) pop() object.Object {
	obj := vm.StackTop()
	vm.stack[vm.sp-1] = nil
	vm.lastPopped = obj
	vm.sp--
	return obj
}

func (vm *VM) LastPopped() object.Object {
	return vm.lastPopped
}
