package doze

import "fmt"

/* Things to remember - Minimalist architecture plan
 *
 * - Primordial inputs are Artifacts which are not created by any Rule.
 * - Final outputs are Artifacts which do not take part in any Rule to create other Artifacts.
 *
 * - Primordial inputs not having changed since last run means their final outputs are unchanged as well.
 * - The above is only true if all tasks are deterministic. It's only useful if we already have the list of Artifacts that changed since last run.
 * - Otherwise, we still have to scan the Graph entirely to first determine which Artifacts have changed.
 *
 * - The Rebuilder takes in a list of target Artifacts to bring up-to-date a creates a list of Rules to execute.
 * - The Scheduler takes in a list of Rules to execute and runs them to completion until target Artifacts are updated.
 *
 * - The Logger stores/prints a trace of all work done by other elements.
 *
 */

// A Graph contains the loaded Rules and Artifacts for a build.
type Graph struct {
	rules     []Rule
	artifacts map[ArtifactTag]Artifact
}

// A Rule represents the action of processing input Artifacts to create
// output Artifacts by executing a Procedure.
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
}

// An ArtifactTag is the combined path+name of an Artifact.
type ArtifactTag string

func NewGraph() *Graph {
	return &Graph{
		artifacts: make(map[ArtifactTag]Artifact),
	}
}

// Register a new Rule in the Graph. Artifacts are created from their tags if they don't yet exist.
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
		// @nocheckin for outputs in inputs

		outputArtifact, ok := g.artifacts[outputTag]
		if !ok {
			// no reference to outputTag, create and store it
			g.artifacts[outputTag] = Artifact{
				tag: outputTag,
			}
			outputArtifact = g.artifacts[outputTag]
		} else if outputArtifact.creatorRule != nil {
			return fmt.Errorf("artifact %s cannot be output two different times", outputTag)
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
