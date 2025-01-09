package myTestProcedures

import "github.com/spectrevert/doze"

func init() {
	doze.RegisterProcedure(LangCLinker{})
	doze.RegisterProcedure(LangCObjectFile{})
}

type LangCLinker struct{}

func (LangCLinker) DozeProcedure() doze.ProcedureInfo {
	return doze.ProcedureInfo{
		ID:  "lang.c.linker",
		New: func() doze.Procedure { return new(LangCLinker) },
	}
}

type LangCObjectFile struct{}

func (LangCObjectFile) DozeProcedure() doze.ProcedureInfo {
	return doze.ProcedureInfo{
		ID:  "lang.c.objectFile",
		New: func() doze.Procedure { return new(LangCObjectFile) },
	}
}
