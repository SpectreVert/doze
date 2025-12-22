package doze

import (
	"fmt"
	"strings"
)

// A Procedure is a type used as a Doze processing function.
type Procedure interface {
	GetProcedureInfo() ProcedureInfo
	Execute(*Rule) error
}

// A ProcedureInfo represents a Procedure registered with Doze.
type ProcedureInfo struct {
	ID  ProcedureID
	New func() Procedure
}

type ProcedureID string

func (id ProcedureID) Namespace() string {
	lastSeparator := strings.LastIndex(string(id), procedureIDSeparator)
	if lastSeparator < 0 {
		return ""
	}
	return string(id)[:lastSeparator]
}

func (id ProcedureID) Name() string {
	if id == "" {
		return ""
	}
	portions := strings.Split(string(id), procedureIDSeparator)
	return portions[len(portions)-1]
}

func RegisterProcedure(instance Procedure) {
	procInfo := instance.GetProcedureInfo()

	if procInfo.ID == "" {
		panic("procedure ID is missing")
	}
	if procInfo.New == nil {
		panic("missing ProcedureInfo.New")
	}
	if p := procInfo.New(); p == nil {
		panic("ProcedureInfo.New must return a non-nil procedure instance")
	}
	procedures[procInfo.ID] = procInfo
}

func GetProcedure(id ProcedureID) (ProcedureInfo, error) {
	p, ok := procedures[id]
	if !ok {
		return ProcedureInfo{}, fmt.Errorf("procedure not registered: %v", id)
	}
	return p, nil
}

var procedures = make(map[ProcedureID]ProcedureInfo)

const procedureIDSeparator = ":"
