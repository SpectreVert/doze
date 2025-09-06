package doze

import (
	"fmt"
	"os"
)

func (graph *Graph) Execute(plan []string) {
	for _, ruleHash := range plan {
		rule, ok := graph.rules[ruleHash]
		if !ok {
			panic("rule '" + ruleHash + "' resolved for execution but doesn't exist")
		}

		// TODO: make sure that Resolves checks if primordial Artifacts exist on disk!
		// TODO: only execute a rule if an Input has changed or if an output doesn't exist.

		rule.Execute(ruleHash)
	}
}

// Placeholder function for running a Rule synchronously.
// TODO: return an error rather than calling os.Exit().
func (rule *Rule) Execute(hash string) {
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
