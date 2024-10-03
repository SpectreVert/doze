package doze

import (
	"fmt"
	"slices"
)

// Describes how a build should be performed.
//
// This is basically a DAG of dependencies with
// artifacts  and rules as nodes and edges, respectively.
//
// @todo implement namedRules, which are user-triggerable rules part of the global DAG.
//
type Dozefile struct {
	runtimeRules     []Rule
	runtimeArtifacts map[string]Artifact
	namedRules       map[string][]Rule
}

type Rule struct {
	inputs   []string
	outputs  []string
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
	isTerminal  bool

	creatorRule *Rule
}

func NewDozefile() Dozefile {
	return Dozefile{
		runtimeArtifacts: make(map[string]Artifact),
		namedRules:       make(map[string][]Rule),
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
				tag:         outputTag,
				isTerminal:  true,
				creatorRule: &newRule,
			}
			outputArtifact, _ = df.runtimeArtifacts[outputTag]
		} else if outputArtifact.creatorRule != nil {
			return fmt.Errorf("artifact %s cannot be output more than once", outputTag)
		}
		outputArtifact.creatorRule = &newRule
		df.runtimeArtifacts[outputTag] = outputArtifact
		newRule.outputs = append(newRule.outputs, outputTag)
	}

	for _, inputTag := range inputTags {
		inputArtifact, ok := df.runtimeArtifacts[inputTag]
		if !ok {
			// no reference to inputTag, create and store it
			df.runtimeArtifacts[inputTag] = Artifact{
				tag:        inputTag,
				isTerminal: false,
			}
		} else if inputArtifact.isTerminal {
			// if the artifact existed before then it's not a terminal output anymore
			inputArtifact.isTerminal = false
			df.runtimeArtifacts[inputTag] = inputArtifact
		}
		newRule.inputs = append(newRule.inputs, inputTag)
	}

	// @nocheckin duplicates
	df.runtimeRules = append(df.runtimeRules, newRule)

	return nil
}

func (df *Dozefile) createNamedRule(inputs []string, outputs []string, procedureTag ProcedureID, name string) error {
	// TODO logic to add a NamedRule (which is like a rule but which needs to be explicitely called by the user)

	return nil
}

// Return a linear plan for a build generated topologically.
func (df *Dozefile) MakePlan(targetTags []string) ([]Rule, error) {
	if targetTags == nil {
		// by default, schedule all terminal outputs (maybe slow?)
		for _, artifact := range df.runtimeArtifacts {
			if artifact.isTerminal == true {
				targetTags = append(targetTags, artifact.tag)
			}
		}
	}

	// Debug
	for _, target := range targetTags {
		fmt.Println("target artifact:", target)
	}

	// reset status
	df.cleanup()

	plan, err := df.topoSchedule(targetTags)
	slices.Reverse(plan)
	return plan, err
}

// Create a topological plan of execution for the rules in runtimeRules,
// based on the dependencies between the artifacts in runtimeArtifacts.
//
// At term this should make use of `constructive traces`,
// that  is local and remote cache functionalities to speed up
// compilation times and enable cloud build type setups.
func (df *Dozefile) topoSchedule(targetTags []string) ([]Rule, error) {
	if targetTags == nil {
		return nil, nil
	}

	var plan []Rule
	var todoList []string
	for _, targetTag := range targetTags {
		artifact, ok := df.runtimeArtifacts[targetTag]
		if !ok {
			return nil, fmt.Errorf("target artifact %s does not exist", targetTag)
		}

		// TODO implement hashing of the artifact / rules / whatnot

		// TODO check if artifact is in cloud

		// TODO check if artifact is in local cache

		// TODO record the new value as a constructive trace

		// default path: schedule the artifact's children (outputs) and creatorRule
		if artifact.creatorRule == nil { /* primordial input */
			continue
		}
		if artifact.creatorRule.isScheduled { /* rule already scheduled */
			continue
		}
		if artifact.isScheduled { /* artifact already scheduled */
			continue
		}

		// add the creatorRule to the plan
		artifact.creatorRule.isScheduled = true
		artifact.isScheduled = true
		plan = append(plan, *artifact.creatorRule)

		// add inputs of the creatorRule to the todoList, as we prepare to go up the dependency tree
		// (check for duplicates)
		for _, inputTag := range artifact.creatorRule.inputs {
			if !slices.Contains(todoList, inputTag) {
				todoList = append(todoList, inputTag)
			}
		}
	}
	morePlan, err := df.topoSchedule(todoList)
	return append(plan, morePlan...), err
}

// Clears the scheduled status, for now only used by topoSchedule.
func (df *Dozefile) cleanup() {
	for _, a := range df.runtimeArtifacts {
		a.isScheduled = false
	}
	for _, r := range df.runtimeRules {
		r.isScheduled = false
	}
}