package doze

import (
	"fmt"
)

// A Dozefile describes how a build should be performed.
type Dozefile struct {
	runtimeRules     []Rule
	runtimeArtifacts map[string]Artifact
	namedRules       map[string][]Rule
}

type Rule struct {
	inputs   []*Artifact
	outputs  []*Artifact
	procInfo ProcedureInfo

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

func NewDozefile() Dozefile {
	return Dozefile{
		runtimeArtifacts: make(map[string]Artifact),
		namedRules: make(map[string][]Rule),
	}
}

// Register a new rule in a Dozefile.
func (df *Dozefile) createRule(inputTags []string, outputTags []string, procedureTag ProcedureID) error {
	var newRule Rule
	var err error
	// @fixme we could use some nicer error types around here
	if len(inputTags) == 0 || len(outputTags) == 0 {
		return fmt.Errorf("rule must contain at least one input and one output")
	}
	if newRule.procInfo, err = GetProcedure(string(procedureTag)); err != nil {
		return err
	}

	for _, outputTag := range outputTags {
		// @nocheckin for duplicate outputs
		// also need to check for outputs in *inputs*...

		outputArtifact, ok := df.runtimeArtifacts[outputTag]
		if !ok {
			// no reference for outputTag, create and store it
			df.runtimeArtifacts[outputTag] = Artifact{
				tag: outputTag,
				isFinal: true,
			}
		} else if outputArtifact.creatorRule != nil {
			return fmt.Errorf("artifact %s cannot be output more than once", outputTag)
		}
		newRule.outputs = append(newRule.outputs, &outputArtifact)
		outputArtifact.creatorRule = &newRule
	}

	for _, inputTag := range inputTags {
		inputArtifact, ok := df.runtimeArtifacts[inputTag]
		if !ok {
			df.runtimeArtifacts[inputTag] = Artifact{
				tag: inputTag,
			}
		} else if (inputArtifact.isFinal) {
			inputArtifact.isFinal = false
		}
		newRule.inputs = append(newRule.inputs, &inputArtifact)
	}

	// @nocheckin duplicates
	df.runtimeRules = append(df.runtimeRules, newRule)

	return nil
}

func (df *Dozefile) createNamedRule(inputs []string, outputs []string, procedureTag ProcedureID, name string) error {
	// TODO logic to add a NamedRule (which is like a rule but which needs to be explicitely called)

	return nil
}