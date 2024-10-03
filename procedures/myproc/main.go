package myproc

import "github.com/spectrevert/doze"

func init() {
	doze.RegisterProcedure(Bigzbi{})
	doze.RegisterProcedure(Bigzben{})
}

// Demo procedure module registering two bogus procedures.
// In the future, the procedures will implement traits for hooking
// up to the build system! For more info see procedures.go
// Anyway, this is basically how a procedure is declared in doze.
// Import the packge in ./cmd/main.go to compile it in and that's it.
//
// It's quite cool.
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