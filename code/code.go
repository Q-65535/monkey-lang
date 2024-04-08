package code

import (
	"fmt"
	"strings"
)

type Instructions []byte
type Opcode byte

const (
	Opconst Opcode = iota
	OpAdd
	OpPop
	OpBang
	OpMinus
	OpSub
	OpMul
	OpDiv
	OpTrue
	OpFalse
	OpNull
	OpEqual
	OpNotEqual
	OpGreaterThan
	OpLessThan
	OpJumpNotTruthy
	OpJump
)

type Definition struct {
	Name          string
	OperandWidths []int
}

func (instructions Instructions) String() string {
	var out strings.Builder
	index := 0
	for index < len(instructions) {
		bt := instructions[index]
		def := definitions[Opcode(bt)]
		out.WriteString(fmt.Sprintf("0x%04x", index))
		index++
		out.WriteString(" ")
		out.WriteString(def.Name)
		for _, width := range def.OperandWidths {
			for i := 0; i < width; i++ {
				out.WriteString(" ")
				out.WriteString(fmt.Sprintf("0x%02x", instructions[index]))
				index++
			}
			// oprand separator
			out.WriteString("|")
		}
		// instruction separator
		out.WriteString("\n")
	}
	return out.String()
}

var definitions = map[Opcode]*Definition{
	Opconst:         {Name: "opConstant", OperandWidths: []int{2}},
	OpAdd:           {"OpAdd", []int{}},
	OpSub:           {"OpSub", []int{}},
	OpMul:           {"OpMul", []int{}},
	OpDiv:           {"OpDiv", []int{}},
	OpBang:          {"OpBang", []int{}},
	OpMinus:         {"OpMinus", []int{}},
	OpPop:           {"OpPop", []int{}},
	OpTrue:          {"OpTrue", []int{}},
	OpFalse:         {"OpFalse", []int{}},
	OpNull:          {"OpNull", []int{}},
	OpEqual:         {"OpEqual", []int{}},
	OpNotEqual:      {"OpNotEqual", []int{}},
	OpGreaterThan:   {"OpGreaterThan", []int{}},
	OpLessThan:      {"OpLessThan", []int{}},
	OpJumpNotTruthy: {"OpJumpNotTruthy", []int{2}},
	OpJump:          {"OpJump", []int{2}},
}

func Make(oc Opcode, oprands ...int) []byte {
	def, ok := definitions[oc]
	if !ok {
		fmt.Printf("opcode %d not defined!", oc)
		return nil
	}
	var instruction []byte
	instruction = append(instruction, byte(oc))
	for i, or := range oprands {
		width := def.OperandWidths[i]
		instruction = append(instruction, filterByte(or, width)...)
	}
	return instruction
}

// convert an integer to an array of bytes in big endian order
func filterByte(target int, byteCount int) []byte {
	var res []byte
	for i := 0; i < byteCount; i++ {
		bitShiftCount := 8 * (byteCount - 1 - i)
		res = append(res, byte(((target >> bitShiftCount) & 0xff)))
	}
	return res
}

func Lookup(oc Opcode) (*Definition, error) {
	def, ok := definitions[oc]
	if !ok {
		return nil, fmt.Errorf("opcode %d not defined", oc)
	}
	return def, nil
}

func ReadUint16(bytes []byte) uint16 {
	val := uint16(bytes[0])<<8 | uint16(bytes[1])
	return val
}
