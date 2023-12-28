package code

import "testing"

func TestInstructionsString(t *testing.T) {
	instructions := []Instructions{
		Make(OpAdd),
		Make(Opconst, 2),
		Make(Opconst, 65534),
	}
	expected := "0x0000 OpAdd\n0x0001 OpConstant 2\n0x0004 OpConstant 65534\n"

	concatted := Instructions{}
	for _, ins := range instructions {
		concatted = append(concatted, ins...)
	}
	if concatted.String() != expected {
		t.Errorf("instructions wrongly formatted:\n wanted:\n%s got:\n%s", expected, concatted.String())
	}
}
