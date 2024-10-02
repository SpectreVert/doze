package doze

import (
	"fmt"
)

// A Dozefile describes how a build should be performed.
type Dozefile struct {
	RuntimeRules     []Rule
	RuntimeArtifacts map[string]Artifact
	NamedRules       map[string][]Rule
}

type Rule struct {
	inputs  []Artifact
	outputs []Artifact
	proc    *Procedure

	isScheduled bool
}

type NamedRule struct {
	name string
	Rule
}

type Artifact struct {
	tag  string
	path string

	isScheduled bool
	isFinal     bool

	creatorRule *Rule
}

// Register a new rule in a Dozefile.
func (df *Dozefile) createRule(inputs []string, outputs []string, procedureTag ProcedureID) error {
	// @fixme maybe we could have some nice error types/enums around here?

	if len(inputs) == 0 || len(outputs) == 0 {
		return fmt.Errorf("rule must contain at least one input and one output")
	}
	if _, err := GetProcedure(string(procedureTag)); err != nil {
		return err
	}

	// TODO logic for updating the linked Artifacts...

	return nil
}

func (df *Dozefile) createNamedRule(inputs []string, outputs []string, procedureTag ProcedureID, name string) error {
	// TODO logic to add a NamedRule (which is like a rule but which needs to be explicitely called)

	return nil
}