package doze

import (
	"fmt"
	"strings"
	"sync"
)

type Procedure interface {
	DozeProcedure() ProcedureInfo
}

type ProcedureInfo struct {
	ID  ProcedureID
	New func() Procedure
}

// String that uniquely indentifies a Doze procedure in the form:
// <namespace>:<name>
type ProcedureID string

// Return the namespace portion of a procedure id
func (id ProcedureID) Namespace() string {
	lastSeparator := strings.LastIndex(string(id), namespaceSeparator)
	if lastSeparator < 0 {
		return ""
	}
	return string(id)[:lastSeparator]
}

// Return the name portion of a procedure id
func (id ProcedureID) Name() string {
	if id == "" {
		return ""
	}
	portions := strings.Split(string(id), namespaceSeparator)
	return portions[len(portions)-1]
}

func RegisterProcedure(instance Procedure) {
	proc := instance.DozeProcedure()

	if proc.ID == "" {
		panic("procedure ID missing")
	}
	if proc.New == nil {
		panic("missing ProcedureInfo.New")
	}
	if p := proc.New(); p == nil {
		panic("ProcedureInfo.New must return a non-nil procedure instance")
	}

	procsMutex.Lock()
	defer procsMutex.Unlock()

	if _, ok := procedures[string(proc.ID)]; ok {
		panic(fmt.Sprintf("procedure already registered: %s", proc.ID))
	}
	procedures[string(proc.ID)] = proc
}

const (
	namespaceSeparator = ':'
)

var (
	procedures = make(map[string]ProcedureInfo)
	procsMutex = sync.RWMutex
)
