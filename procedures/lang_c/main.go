package lang_c

import (
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/spectrevert/doze"
)

func init() {
	doze.RegisterProcedure(ObjectFile{})
	doze.RegisterProcedure(Executable{})
}

type ObjectFile struct {
}

func (ObjectFile) GetProcedureInfo() doze.ProcedureInfo {
	return doze.ProcedureInfo{
		ID:  "lang:c:object_file",
		New: func() doze.Procedure { return new(ObjectFile) },
	}
}

// ObjectFile.Execute creates an object file from the C source code file passed to it.
// Inputs must contain exactly one '.c' file and outputs must contain exactly one '.o' file.
// Any amount of headers are accepted in the inputs (and ignored).
func (ObjectFile) Execute(rule *doze.Rule) error {
	var sourceFile, objectFile string
	for _, input := range rule.Inputs {
		if strings.HasSuffix(input.NormalizedTag(), ".c") {
			if sourceFile != "" {
				return fmt.Errorf("found more than one '.c' file in rule %v", rule)
			}
			sourceFile = input.NormalizedTag()
		} else if !strings.HasSuffix(input.NormalizedTag(), ".h") {
			return fmt.Errorf("found a non-c source code file declared in rule %v", rule)
		}
	}
	for _, output := range rule.Outputs {
		if strings.HasSuffix(output.NormalizedTag(), ".o") {
			if objectFile != "" {
				return fmt.Errorf("found more than one '.o' file in rule %v", rule)
			}
			objectFile = output.NormalizedTag()
		} else {
			return fmt.Errorf("found a non-c object file declared in rule %v", rule)
		}
	}

	cmd := exec.Command("gcc", "-c", sourceFile, "-o", objectFile)
	fmt.Println("doze: gcc -c", sourceFile, "-o", objectFile)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("error creating stderr pipeline: %v", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error creating stdout pipeline: %v", err)
	}

	if err = cmd.Start(); err != nil {
		return fmt.Errorf("error starting 'gcc': %v", err)
	}

	slurp, _ := io.ReadAll(stdout)
	if len(slurp) > 0 {
		fmt.Printf("%s\n", slurp)
	}
	slurp, _ = io.ReadAll(stderr)
	if len(slurp) > 0 {
		fmt.Printf("%s\n", slurp)
	}

	if err = cmd.Wait(); err != nil {
		return fmt.Errorf("error waiting for 'gcc': %v", err)
	}

	return nil
}

type Executable struct {
}

func (Executable) GetProcedureInfo() doze.ProcedureInfo {
	return doze.ProcedureInfo{
		ID:  "lang:c:executable",
		New: func() doze.Procedure { return new(Executable) },
	}
}

func (Executable) Execute(rule *doze.Rule) error {
	var inputs []string

	for _, input := range rule.Inputs {
		if !strings.HasSuffix(input.NormalizedTag(), ".o") {
			return fmt.Errorf("found a non-c object fiel declared in rule %v", rule)
		}
		inputs = append(inputs, input.NormalizedTag())
	}
	if len(rule.Outputs) > 1 {
		return fmt.Errorf("found more than one output in rule %v", rule)
	}
	var executable = rule.Outputs[0].NormalizedTag()
	args := []string{"-o", executable}
	args = append(args, inputs...)

	cmd := exec.Command("gcc", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error executing 'gcc': %v", err)
	}

	fmt.Println("doze: gcc", "-o", executable, inputs)
	return nil
}

// interface guard
var _ doze.Procedure = (*ObjectFile)(nil)
var _ doze.Procedure = (*Executable)(nil)
