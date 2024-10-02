package myproc

import "github.com/spectrevert/doze"

func init() {
	doze.RegisterProcedure(Bigzbi{})
}

type Bigzbi struct {}

func (Bigzbi) DozeProcedure() doze.ProcedureInfo {
	return doze.ProcedureInfo{
		ID: "somenamespace:bigzbi",
		New: func() doze.Procedure { return new(Bigzbi) },
	}
}