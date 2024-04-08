package vm

import (
	"fmt"
	"monkey/code"
	"monkey/compiler"
	"monkey/object"
)

const StackSize = 2048

var True = &object.Boolean{Value: true}
var False = &object.Boolean{Value: false}
var Null = &object.Null{}

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
		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan, code.OpLessThan:
			var res bool
			right := vm.pop()
			left := vm.pop()
			if left.Type() != right.Type() {
				return fmt.Errorf("unsupported types for comparison operation: %s %s", left.Type(), right.Type())
			}
			if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
				leftVal := left.(*object.Integer).Value
				rightVal := right.(*object.Integer).Value
				switch op {
				case code.OpEqual:
					res = (leftVal == rightVal)
				case code.OpNotEqual:
					res = (leftVal != rightVal)
				case code.OpGreaterThan:
					res = (leftVal > rightVal)
				case code.OpLessThan:
					res = (leftVal < rightVal)
				}
			}
			if left.Type() == object.BOOLEAN_OBJ && right.Type() == object.BOOLEAN_OBJ {
				leftVal := left.(*object.Boolean).Value
				rightVal := right.(*object.Boolean).Value
				switch op {
				case code.OpEqual:
					res = (leftVal == rightVal)
				case code.OpNotEqual:
					res = (leftVal != rightVal)
				// @Optimize: this error could be reported at compile time (instead of run time)
				case code.OpGreaterThan:
					return fmt.Errorf("unsupported types for > operator: %s %s", left.Type(), right.Type())
				case code.OpLessThan:
					return fmt.Errorf("unsupported types for < operator: %s %s", left.Type(), right.Type())
				}
			}
			vm.push(nativeBool2BooleanObject(res))

		case code.OpTrue:
			err := vm.push(True)
			if err != nil {
				return err
			}
		case code.OpFalse:
			err := vm.push(False)
			if err != nil {
				return err
			}
		case code.OpBang:
			vm.executeBangOperator()
		case code.OpNull:
			err := vm.push(Null)
			if err != nil {
				return err
			}
		case code.OpPop:
			vm.pop()
		case code.OpJump:
			// off set: -1, so the next iteration jumps to the correct position
			pos := code.ReadUint16(vm.instructions[ip+1:])
			ip = int(pos - 1)
		case code.OpJumpNotTruthy:
			pos := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2
			condition := vm.pop()
			if !isTruthy(condition) {
				ip = int(pos - 1)
			}
		default:
			return fmt.Errorf("unknown operator: %d", op)
		}
	}
	return nil
}

func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case *object.Boolean:
		return obj.Value
	case *object.Null:
		return false
	default:
		return true
	}

	c := obj.(*object.Boolean)
	return c.Value
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

func (vm *VM) executeBangOperator() error {
	obj := vm.pop()
	if obj == False || obj == Null {
		return vm.push(True)
	} else {
		return vm.push(False)
	}
}

func (vm *VM) LastPopped() object.Object {
	return vm.lastPopped
}

func nativeBool2BooleanObject(input bool) *object.Boolean {
	if input {
		return True
	} else {
		return False
	}
}
