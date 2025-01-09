package doze

import "fmt"

// Things to remember
// 1. Terminal inputs being the same as a cached value (hash) means that its outputs must be the same
//    until they are created by a dynamic rule.
// 2. The above rule is only useful if we already know the list of Artifacts which changed since the last
//    run.
// 3. Otherwise, we have to scan the Graph entirely to first determine which Artifacts have changed.

// 5. The Rebuilder takes in a list of Artifacts to bring up-to-date and creates a list of Artifacts to
// 6. The Scheduler takes in a list of Rules to execute and runs them to completion in the correct order.

// 7. The Logger stores/prints a trace of all the work done by all of the other elements.

type Graph struct {
	rules     []Rule
	artifacts map[ArtifactTag]Artifact
}

type Rule struct {
	inputs   []ArtifactTag
	outputs  []ArtifactTag
	procInfo ProcedureInfo

	scheduled bool
}

type Artifact struct {
	tag         ArtifactTag
	creatorRule *Rule
}

// An ArtifactTag is the combined name + path of an Artifact.
type ArtifactTag string

type Rebuilder interface {
	rebuild([]ArtifactTag, *Graph) ([]Rule, error)
}

func NewGraph() *Graph {
	return &Graph{
		artifacts: make(map[ArtifactTag]Artifact),
	}
}

// Register a new rule in the Graph
func (g *Graph) createRule(inputs []ArtifactTag, outputs []ArtifactTag, procID ProcedureID) error {
	var r Rule
	var err error

	// @fixme nicer error types
	if len(inputs) == 0 || len(outputs) == 0 {
		return fmt.Errorf("rule must contain at least one input and one output")
	}
	if r.procInfo, err = GetProcedure(procID); err != nil {
		return err
	}
	for _, outputTag := range outputs {
		// @nocheckin for duplicate outputs
		// also need to check for outputs in inputs

		outputArtifact, ok := g.artifacts[outputTag]
		if !ok {
			// no reference to outputTag, create and store it
			g.artifacts[outputTag] = Artifact{
				tag: outputTag,
			}
			outputArtifact = g.artifacts[outputTag]
		} else if outputArtifact.creatorRule != nil {
			return fmt.Errorf("artifact %s cannot be output by two different rules", outputTag)
		}
		r.outputs = append(r.outputs, outputTag)
		outputArtifact.creatorRule = &r
		g.artifacts[outputTag] = outputArtifact
	}
	for _, inputTag := range inputs {
		_, ok := g.artifacts[inputTag]
		if !ok {
			g.artifacts[inputTag] = Artifact{
				tag: inputTag,
			}
		}
		r.inputs = append(r.inputs, inputTag)
	}
	g.rules = append(g.rules, r)
	return nil
}

// A Graph contains all the loaded rules and artifacts.
/* type Graph struct {
	rules     []Rule
	artifacts map[ArtifactTag]Artifact
}

// A Rule represents the action of consuming input artifacts
// to create output artifacts by executing a Procedure.
type Rule struct {
	inputs   []ArtifactTag
	outputs  []ArtifactTag
	procInfo ProcedureInfo

	scheduled bool
}

// An Artifact represents a single file on disk.
type Artifact struct {
	tag         ArtifactTag
	creatorRule *Rule

	scheduled bool
}

// An ArtifactTag is the combined name + path of an Artifact.
type ArtifactTag string

// A Rebuilder takes in a list of Artifacts to bring up-to-date (and a Graph) and creates a list of rules to execute.
type Rebuilder interface {
	// Returns a list of rules to execute, from last to first.
	rebuild([]ArtifactTag, *Graph) ([]Rule, error)
	// Cleans up the Graph after the list of rules to execute has been executed by a Scheduler.
	cleanup(*Graph)
}

// A Scheduler takes in a list of Rules to execute (and a Graph) and runs them to completion in the correct order.
// The `execute` function returns when the build is completed or in case of error.
type Scheduler interface {
	execute([]Rule, *Graph) error
}

// -- Actual code under this -- //

func NewGraph() *Graph {
	return &Graph{
		artifacts: make(map[ArtifactTag]Artifact),
	}
}

// Register a new rule in the Graph
func (g *Graph) createRule(inputs []ArtifactTag, outputs []ArtifactTag, procID ProcedureID) error {
	var r Rule
	var err error

	// @fixme nicer error types
	if len(inputs) == 0 || len(outputs) == 0 {
		return fmt.Errorf("rule must contain at least one input and one output")
	}
	if r.procInfo, err = GetProcedure(procID); err != nil {
		return err
	}
	for _, outputTag := range outputs {
		// @nocheckin for duplicate outputs
		// also need to check for outputs in inputs

		outputArtifact, ok := g.artifacts[outputTag]
		if !ok {
			// no reference to outputTag, create and store it
			g.artifacts[outputTag] = Artifact{
				tag: outputTag,
			}
			outputArtifact = g.artifacts[outputTag]
		} else if outputArtifact.creatorRule != nil {
			return fmt.Errorf("artifact %s cannot be output by two different rules", outputTag)
		}
		r.outputs = append(r.outputs, outputTag)
		outputArtifact.creatorRule = &r
		g.artifacts[outputTag] = outputArtifact
	}
	for _, inputTag := range inputs {
		_, ok := g.artifacts[inputTag]
		if !ok {
			g.artifacts[inputTag] = Artifact{
				tag: inputTag,
			}
		}
		r.inputs = append(r.inputs, inputTag)
	}
	g.rules = append(g.rules, r)
	return nil
} */
