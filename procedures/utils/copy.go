package utils

import (
	"fmt"

	"github.com/spectrevert/doze"
)

func init() {
	doze.RegisterProcedure(Copy{})
}

type Copy struct {
}

func (Copy) GetProcedureInfo() doze.ProcedureInfo {
	return doze.ProcedureInfo{
		ID:  "utils:copy",
		New: func() doze.Procedure { return new(Copy) },
	}
}

// Copy must receive the same number of inputs and outputs.
// It will copy each input to its respective output location.
func (Copy) Execute(rule *doze.Rule) error {
	if len(rule.Inputs) != len(rule.Outputs) {
		return fmt.Errorf("received unequal amount of input and output artifacts: %d and %d", len(rule.Inputs), len(rule.Outputs))
	}

	for idx, tag := range rule.Inputs {
		doze.CreateBaseDir(string(rule.Outputs[idx]))
		if err := doze.CopyFile(string(tag), string(rule.Outputs[idx])); err != nil {
			return fmt.Errorf("os.Link failed: %s", err)
		}
	}

	return nil
}
