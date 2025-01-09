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

// A ProcedureID is the combined namespace + name of a Procedure.
// It has to be unique.
type ProcedureID string

// A ProcedureError is raised if a Procedure is registered twice.
type ProcedureError struct {
	name string
}

func (err ProcedureError) Error() string {
	return fmt.Sprintf("procedure not registered: %s", err.name)
}

// Returns the namespace portion of a ProcedureID
func (id ProcedureID) Namespace() string {
	lastSeparator := strings.LastIndex(string(id), namespaceSeparator)
	if lastSeparator < 0 {
		return ""
	}
	return string(id)[:lastSeparator]
}

// Returns the name portion of a ProcedureID.
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

	if _, ok := procedures[proc.ID]; ok {
		panic(fmt.Sprintf("procedure already registered: %s", proc.ID))
	}
	procedures[proc.ID] = proc
}

func GetProcedure(procID ProcedureID) (ProcedureInfo, error) {
	procsMutex.Lock()
	defer procsMutex.Unlock()

	p, ok := procedures[procID]
	if !ok {
		return ProcedureInfo{}, ProcedureError{}
	}
	return p, nil
}

const namespaceSeparator = "."

var (
	procedures = make(map[ProcedureID]ProcedureInfo)
	procsMutex sync.RWMutex
)
