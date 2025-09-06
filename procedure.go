package doze

import (
	"fmt"
	"strings"
)

// Procedure is an interface that is implemented by the user to customize the type of work that Doze can do.
// A Procedure can only be run by Doze after being registered.
type Procedure interface {
	// Returns the Procedure's ProcedureInfo.
	GetProcedureInfo() ProcedureInfo
	// Execute the Rule with the Procedure.
	Execute(*Rule) error
}

// A ProcedureInfo represents a Procedure registered with the Doze executable.
// - ID is the full name of the Procedure. It is unique and namespaced.
// - New() returns a pointer to a new instance of the Procedure.
type ProcedureInfo struct {
	ID  ProcedureID
	New func() Procedure
}

// ProcedureID is a string that uniquely identifies a Doze Procedure.
// It is made up of colon separated labels which form a hierarchy from left to right,
// with the last label making up the Procedure name and the labels before that making its namespace.
type ProcedureID string

// Namespace returns the namespace portion of a ProcedureID, if it exists, or an empty string.
func (id ProcedureID) Namespace() string {
	lastSeparator := strings.LastIndex(string(id), procedureIDSeparator)
	if lastSeparator < 0 {
		return ""
	}
	return string(id)[:lastSeparator]
}

// Name returns the name portion of a ProcedureID.
func (id ProcedureID) Name() string {
	if id == "" {
		return ""
	}
	portions := strings.Split(string(id), procedureIDSeparator)
	return portions[len(portions)-1]
}

// RegisterProcedure registers a new Procedure in the Doze global Procedure map.
// Any amount of Procedure instances may be created from it.
// Duplicate Procedures are not accepted.
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

	if _, ok := procedures[procInfo.ID]; ok {
		panic(fmt.Sprintf("procedure already registered: '%v'", procInfo.ID))
	}
	procedures[procInfo.ID] = procInfo
}

// GetProcedure returns a Procedure information from its ID.
func GetProcedure(id ProcedureID) (ProcedureInfo, error) {
	p, ok := procedures[id]
	if !ok {
		return ProcedureInfo{}, fmt.Errorf("procedure not registered: '%v'", id)
	}
	return p, nil
}

const procedureIDSeparator = ":"

// The Procedures registered with the Doze executable.
var procedures = make(map[ProcedureID]ProcedureInfo)
