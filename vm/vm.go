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
