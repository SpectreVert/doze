// This is heavily inspired from the modules approach of the Caddy webserver.
// We instead used procedures, which are essentially user-written logic for
// processing inputs into their outputs. For instance a procedure could start
// a system thread to compile a C file into an object file.
// A procedure MUST respect the golden rule, that is: no side effects on files
// others than the inputs and the outputs. In functional terms, it should be as
// pure as possible. In some cases that can be hard to prove or downright implement,
// but essentially all files in the build system MUST be accounted for.
// (There is a world in which we bypass this rule and verify ourselves
// that all dependencies have been accounted for, like Tup build system does,
// but this is not the priority here).
// So declare properly all inputs and all outputs while writing Rules.
// The second golden rule is that a sequence `a` of inputs with values [A] must yield
// the same outputs [B] everytime they get processed. That is, the procedure MUST be
// deterministic and return for each input combination a predictable output combination.
// Unless stated otherwise, for example, as a non-deterministic procedure. Again, this
// is doable but not the priority for now.
// The third and last golden rule is that a procedure should be as fast as possible.
// System calls will be possible, provided I make a wrapper around it
// somehow to do some preparing and cleaning up of the task execution,
// but should be avoided for Go funcs, C functions are good as well
// but limit the ability of doze to be compiled easily on all platforms
// (AFAIU, I might be wrong here.)
// (If someone could embed compilation of C++ code in a Go library..
// Would be great for science))
// Anyway, same here no real documentation, code is mostly self-explanatory.
package doze

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// This is the main user-facing interface for doze.
// Procedures are called during rule execution and transform
// inputs into outputs.
// At term we should have interfaces for
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

func GetProcedures(scope string) []ProcedureInfo {
	procsMutex.Lock()
	defer procsMutex.Unlock()

	var procs []ProcedureInfo
	for _, proc := range procedures {
		procs = append(procs, proc)
	}

	// make return value deterministic
	sort.Slice(procs, func(a, b int) bool {
		return procs[a].ID < procs[b].ID
	})

	return procs
}

func GetProcedure(name string) (ProcedureInfo, error) {
	procsMutex.Lock()
	defer procsMutex.Unlock()

	p, ok := procedures[name]
	if !ok {
		return ProcedureInfo{}, fmt.Errorf("procedure not registered: %s", name)
	}
	return p, nil
}

// not sure if this separator will stay, "." is easier on the eyes.
const (
	namespaceSeparator = ":"
)

var (
	procedures = make(map[string]ProcedureInfo)
	procsMutex sync.RWMutex
)
