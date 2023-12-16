package code

import "testing"

func TestInstructionsString(t *testing.T) {
	instructions := []Instructions{
		Make(Opconst, 1),
		Make(Opconst, 2),
		Make(Opconst, 65535),
	}
	expected := "0000 OpConstant 1\n 0003 OpConstant 2\n 0006 OpConstant 65535\n"

	concatted := Instructions{}
	for _, ins := range instructions {
		concatted = append(concatted, ins...)
	}
	if concatted.String() != expected {
		t.Errorf("instructions wrongly formatted:\n wanted:\n%s got:\n%s", expected, concatted.String())
	}
}
