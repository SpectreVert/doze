package myproc

import "github.com/spectrevert/doze"

func init() {
	doze.RegisterProcedure(Bigzbi{})
	doze.RegisterProcedure(Bigzben{})
}

type Bigzbi struct {}
type Bigzben struct {}

func (Bigzbi) DozeProcedure() doze.ProcedureInfo {
	return doze.ProcedureInfo{
		ID: "somenamespace:bigzbi",
		New: func() doze.Procedure { return new(Bigzbi) },
	}
}

func (Bigzben) DozeProcedure() doze.ProcedureInfo {
	return doze.ProcedureInfo{
		ID: "somenamespace:bigzben",
		New: func() doze.Procedure { return new(Bigzben) },
	}
}