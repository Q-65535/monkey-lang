package code

import "fmt"

type Instructions []byte
type Opcode byte

const (
	Opconst Opcode = iota
)

type Definition struct {
	Name         string
	OperandWidth []int
}

var definitions = map[Opcode]*Definition{
	Opconst: &Definition{Name: "opConstant", OperandWidth: []int{2}},
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
		width := def.OperandWidth[i]
		instruction = append(instruction, filterByte(or, width)...)
	}
	return instruction
}

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
		return nil, fmt.Errorf("opcode %d not defined!\n", oc)
	}
	return def, nil
}
