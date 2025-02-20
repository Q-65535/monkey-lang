package vm

import (
	"fmt"
	"monkey/code"
	"monkey/compiler"
	"monkey/object"
)

const StackSize = 2048
const GlobalSize = 65536
const MaxFrames = 1024

var True = &object.Boolean{Value: true}
var False = &object.Boolean{Value: false}
var Null = &object.Null{}

type VM struct {
	constants  []object.Object
	stack      []object.Object
	sp         int
	lastPopped object.Object
	globals    []object.Object
	frames     []*Frame
	frameIndex int
}

func New(bytecode *compiler.Bytecode) *VM {
	fn := &object.CompiledFunction{Instructions: bytecode.Instructions}
	mainFrame := NewFrame(fn)
	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame
	return &VM{
		constants: bytecode.Constants,
		stack:     make([]object.Object, StackSize),
		sp:        0,
		// @Optimization: we can determine the globalsize at compile time, and reduce the array size
		globals:    make([]object.Object, GlobalSize),
		frames:     frames,
		frameIndex: 1,
	}
}

func NewWithGlobalsStore(bytecode *compiler.Bytecode, s []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = s
	return vm
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return &object.Null{}
	} else {
		return vm.stack[vm.sp-1]
	}
}

func (vm *VM) Run() error {
	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().ip++
		ip := vm.currentFrame().ip
		ins := vm.currentFrame().Instructions()
		op := code.Opcode(ins[ip])
		switch op {
		case code.Opconst:
			constIndex := code.ReadUint16(ins[ip+1:])
			integer := vm.constants[constIndex]
			if integer == nil {
				return fmt.Errorf("const not found, problem index: %d", constIndex)
			}
			vm.currentFrame().ip += 2
			err := vm.push(integer)
			if err != nil {
				return err
			}
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			right := vm.pop()
			left := vm.pop()
			switch {
			case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
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
			case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
				leftVal := left.(*object.String).Value
				rightVal := right.(*object.String).Value
				if op != code.OpAdd {
					return fmt.Errorf("unsupported operator for string type")
				}
				vm.push(&object.String{Value: leftVal + rightVal})
			default:
				return fmt.Errorf("unsupported types for binary operation: %s %s", left.Type(), right.Type())
			}
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
			pos := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip = int(pos - 1)
		case code.OpJumpNotTruthy:
			pos := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			condition := vm.pop()
			if !isTruthy(condition) {
				vm.currentFrame().ip = int(pos - 1)
			}
		case code.OpSetGlobal:
			index := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			vm.globals[index] = vm.pop()
		case code.OpGetGlobal:
			index := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			val := vm.globals[index]
			err := vm.push(val)
			if err != nil {
				return err
			}
		case code.OpArray:
			oprand := code.ReadUint16(ins[ip+1:])
			count := int(oprand)
			vm.currentFrame().ip += 2
			elements := make([]object.Object, count)
			for i := 0; i < count; i++ {
				elements[count-1-i] = vm.pop()
			}
			array_obj := object.Array{Value: elements}
			err := vm.push(&array_obj)
			if err != nil {
				return err
			}
		case code.OpIndex:
			index := vm.pop()
			arr := vm.pop()
			if arr.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ {
				i := index.(*object.Integer).Value
				a := arr.(*object.Array).Value
				err := vm.push(a[i])
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("we currently only support array access, expect ARRAY_OBJ and INTEGER_OBJ types, but got %s, %s", arr.Type(), index.Type())
			}
		case code.OpCall:
			fn, ok := vm.pop().(*object.CompiledFunction)
			if !ok {
				return fmt.Errorf("calling non-function")
			}
			vm.pushFrame(NewFrame(fn))
		case code.OpReturnValue:
			vm.popFrame()
		case code.OpReturn:
			vm.popFrame()
			err := vm.push(Null)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown operator: %d", op)
		}
	}
	return nil
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.frameIndex-1]
}

func (vm *VM) pushFrame(frame *Frame) {
	vm.frames[vm.frameIndex] = frame
	vm.frameIndex++
}

func (vm *VM) popFrame() *Frame {
	vm.frameIndex--
	return vm.frames[vm.frameIndex]
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
