package doze

import (
	"fmt"
	"os"
)

// Reset all Artifacts Modified and Exists flags.
// For now, we assume that primordial Artifacts are never Modified...
func (graph *Graph) SetArtifactStates() {
	for normalizedTag, artifact := range graph.artifacts {
		_, err := os.Stat(normalizedTag)
		artifact.Exists = (err == nil)
		// We don't know yet if the Artifact was modified.
		artifact.Modified = false
	}
}

// Execute to completion a topologically sorted list of Rules (passed by their Hash), with the goal of bringing the Graph up-to-date.
// Executes synchronously and single-threadedly.
func (graph *Graph) Execute(plan []string) {
	// Check if Artifacts exist and if they have been Modified.
	graph.SetArtifactStates()

	var executedRules = 0
	for _, ruleHash := range plan {
		rule, ok := graph.rules[ruleHash]
		if !ok {
			panic("rule [" + ruleHash + "] resolved for execution but doesn't exist")
		}

		var ruleIsOutdated = false
		for _, inputTag := range rule.Inputs {
			if graph.artifacts[inputTag.NormalizedTag()].Modified {
				ruleIsOutdated = true
				goto OutdatedRule
			}
		}
		for _, outputTag := range rule.Outputs {
			if !graph.artifacts[outputTag.NormalizedTag()].Exists {
				ruleIsOutdated = true
				goto OutdatedRule
			}
		}

	OutdatedRule:
		if ruleIsOutdated {
			rule.Execute()
			executedRules += 1

			// Mark all output Artifacts as modified.
			for _, outputTag := range rule.Outputs {
				graph.artifacts[outputTag.NormalizedTag()].Modified = true
			}
		}
	}

	if executedRules == 0 {
		fmt.Println("doze: Nothing to do.")
	}
}

// Placeholder function for running a Rule synchronously.
// TODO: return an error rather than calling os.Exit().
func (rule *Rule) Execute() {
	procInfo, err := GetProcedure(rule.procID)
	if err != nil {
		fmt.Println("rule.Execute:", err)
		os.Exit(2)
	}
	proc := procInfo.New()
	if err = proc.Execute(rule); err != nil {
		fmt.Println("rule.Execute:", err)
		os.Exit(2)
	}
}
