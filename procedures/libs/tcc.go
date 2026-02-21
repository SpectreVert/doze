package tcc

import (
	"fmt"
	"path/filepath"

	"github.com/spectrevert/doze"
)

/*
	#cgo LDFLAGS: -ltcc
	#include <libtcc.h>
*/
import "C"

func init() {
	doze.RegisterProcedure(Obj{})
	doze.RegisterProcedure(Exe{})
}

/*
	TODO

	- add checks to see if the inputs and outputs have the correct extensions.
	- add checks to see if the number of inputs / outputs is correct.
*/

// Compile an object file from a .c source file.
// .h header files can also be passed as inputs so that changing them triggers a recompilation.
// Takes at least one input (including exactly one .c source file) and exactly one .o output file.
type Obj struct {
}

func (Obj) GetProcedureInfo() doze.ProcedureInfo {
	return doze.ProcedureInfo{
		ID:  "libs:tcc:obj",
		New: func() doze.Procedure { return new(Obj) },
	}
}

func (Obj) Execute(rule *doze.Rule) error {
	if len(rule.Outputs) != 1 {
		return fmt.Errorf("must receive exactly one output, received %d", len(rule.Outputs))
	}
	obj := rule.Outputs[0]

	tccState := C.tcc_new()
	defer C.tcc_delete(tccState)

	C.tcc_set_output_type(tccState, C.TCC_OUTPUT_OBJ)

	gotSourceFile := false
	for _, tag := range rule.Inputs {
		switch filepath.Ext(string(tag)) {
		case ".h":
		case ".c":
			if C.tcc_add_file(tccState, C.CString(string(tag))) == -1 {
				return fmt.Errorf("failed to compile %s", string(tag))
			}
			gotSourceFile = true
		default:
			return fmt.Errorf("input can only be C source or header file (.c or .h), got %s", string(tag))
		}
	}
	if !gotSourceFile {
		return fmt.Errorf("expected at least one C source file (.c), got none")
	}

	if C.tcc_output_file(tccState, C.CString(string(obj))) == -1 {
		return fmt.Errorf("failed to produce object file %s", string(obj))
	}

	return nil
}

// Compile an executable from .c source files, or link it from .o files. C source files and object files can be mixed.
// .h header files can also be passed as inputs so that changing them triggers a recompilation.
// Takes at least one input and exactly one output (the executable name).
type Exe struct {
}

func (Exe) GetProcedureInfo() doze.ProcedureInfo {
	return doze.ProcedureInfo{
		ID:  "libs:tcc:exe",
		New: func() doze.Procedure { return new(Exe) },
	}
}

func (Exe) Execute(rule *doze.Rule) error {
	if len(rule.Outputs) != 1 {
		return fmt.Errorf("must receive exactly one output, received %d", len(rule.Outputs))
	}
	exe := rule.Outputs[0]

	tccState := C.tcc_new()
	defer C.tcc_delete(tccState)

	C.tcc_set_output_type(tccState, C.TCC_OUTPUT_EXE)

	gotSourceOrObjectFile := false
	for _, tag := range rule.Inputs {
		switch filepath.Ext(string(tag)) {
		case ".h":
		case ".c":
			if C.tcc_add_file(tccState, C.CString(string(tag))) == -1 {
				return fmt.Errorf("failed to compile C source file %s", string(tag))
			}
			gotSourceOrObjectFile = true
		case ".o":
			if C.tcc_add_file(tccState, C.CString(string(tag))) == -1 {
				return fmt.Errorf("failed to link object file %s", string(tag))
			}
			gotSourceOrObjectFile = true
		default:
			return fmt.Errorf("input can only be C source, header or object file (.c .h or .o), got %s", string(tag))
		}
	}
	if !gotSourceOrObjectFile {
		return fmt.Errorf("expected at least one C source file or object file (.c or .o), got none")
	}

	if C.tcc_output_file(tccState, C.CString(string(exe))) == -1 {
		return fmt.Errorf("failed to produce executable %s", string(exe))
	}

	return nil
}

// TODO: add interface guards here.
