package doze

import (
	"container/list"
	"fmt"
	"slices"
)

type Graph struct {
	rules     map[string]*Rule
	artifacts map[string]*Artifact

	checksum string
}

type ResolveMode int

func NewGraph() *Graph {
	return &Graph{
		rules:     make(map[string]*Rule),
		artifacts: make(map[string]*Artifact),
	}
}

// Returns the number of scheduled rules, or an error if something bad happened.
func Resolve(graph *Graph, cache Cache, mode ResolveMode) (int, error) {
	cache.ClearPlan()

	// Computes the topological order of the Graph. (Kahn's Algorithm)

	var rulesToInspect = list.New() // This list gets modified in-place to iterate over the Rules to inspect.
	var scheduledRules []string     // Because the above list gets modified, we need to keep track of which Rules were inspected or scheduled for inspection already.
RulesToInspectLoop:
	// Make an initial list of Rules whose input Artifacts do not have a creator Rule (primordial Artifacts).
	// While making the list, check if these rules (primordial Rules) were run last time or if they need to be scheduled.
	for checksum, rule := range graph.rules {
		for _, tag := range rule.Inputs {
			if graph.artifacts[string(tag)].creator != nil {
				// We can't schedule that Rule because it has a creator rule. We will eventually reach it and schedule it normally (when it is ready).
				continue RulesToInspectLoop
			}
			// NOTE: here we could still check if the input Artifact exists on disk.
		}

		// If the rule is missing from the `last_run` cache or if we are in full mode, schedule the rule.
		if !cache.RuleInLastRun(rule) || mode == FullMode {
			rulesToInspect.PushBack(checksum)
			scheduledRules = append(scheduledRules, checksum)
		}
		cache.RecordRule(rule)
	}

	// Run this loop until rulesToInspect is empty.
	for e := rulesToInspect.Front(); e != nil; e = rulesToInspect.Front() {
		checksum := rulesToInspect.Remove(e).(string)
		cache.ScheduleRule(graph.rules[checksum])

		// Iterate over each output Artifact of the Rule to check their consumer Rules.
		for _, tag := range graph.rules[checksum].Outputs {
		CheckConsumerRules:
			// Iterate over each consumer Rule of the input Artifact being inspected.
			for _, consumerRule := range graph.artifacts[string(tag)].consumers {
				// The consumer rule is scheduled if:
				// - (At least) one input of the consumerRule is missing.
				// - All inputs of the consumerRule have either no creator Rule or its creator Rule is already scheduled.

				// Iterate over each input of the consumer Rule.
				for _, consumerTag := range consumerRule.Inputs {
					if consumerTag.Exists() || consumerTag == tag {
						continue
					} else if graph.artifacts[string(consumerTag)].creator != nil &&
						!slices.Contains(scheduledRules, graph.artifacts[string(consumerTag)].creator.Checksum()) {
						continue CheckConsumerRules
					}
				}
				// Schedule the consumerRule if it was not already scheduled.
				if !slices.Contains(scheduledRules, consumerRule.Checksum()) {
					rulesToInspect.PushBack(consumerRule.Checksum())
					scheduledRules = append(scheduledRules, consumerRule.Checksum())
				}
			}
		}
	}

	return len(scheduledRules), nil
}

func Execute(graph *Graph, cache Cache) (int, int, error) {
	var executedRules = 0
	var fetchedRules = 0

	for _, checksum := range cache.Plan() {
		rule, ok := graph.rules[checksum]
		if !ok {
			return 0, 0, fmt.Errorf("rule [%s] resolved for execution but doesn't exist", checksum)
		}

		if cache.HasArtifacts(rule) {
			fmt.Printf("cached: [%s]\n", checksum)
			if err := cache.ProduceArtifacts(rule); err != nil {
				return 0, 0, err
			}
			fetchedRules += 1
		} else {
			fmt.Printf("run: [%s]\n", checksum)
			rule.Execute()
			executedRules += 1

			if err := cache.StoreArtifacts(rule); err != nil {
				return 0, 0, err
			}
		}
	}

	if err := cache.FlushRecords(); err != nil {
		return 0, 0, err
	}

	return executedRules, fetchedRules, nil
}

const (
	// TerseMode checks if any of the primoardial inputs of the Graph have changed since the last run. If the cache is empty, then all rules descending from the primordial inputs
	// are executed. If the cache already has traces of a previous run, only primordial inputs which have changed since this last run trigger the execution of their Rules.
	// If an output has been deleted externally (not during a Doze run) and the inputs of the Rule have not changed, Doze will not select that Rule for execution.
	TerseMode ResolveMode = iota

	// FullMode schedules all the Rules scheduled by TerseMode, with the addition of Rules that are missing at least an output in the build directory.
	// Useful when some files have been deleted manually outside of Doze execution, or if you changed Dozefile.yml since the last run.
	FullMode
)
