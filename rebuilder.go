package doze

import (
	"fmt"
)

type DeepRebuilder struct{}

func (dr *DeepRebuilder) rebuild(targets []ArtifactTag, level int, graph *Graph) ([]*Rule, error) {
	if graph == nil {
		return nil, fmt.Errorf("graph is nil")
	}
	if targets == nil {
		return nil, nil
	}

	var plan []*Rule

	for _, target := range targets {
		a, ok := graph.artifacts[target]
		if !ok {
			return nil, fmt.Errorf("target artifact %s does not exist", target)
		}
		if a.creatorRule == nil {
			// a is a primordial input
			if level == 0 {
				// if level == 0, `a` *is* an original target also, which is an error
				return nil, fmt.Errorf("target artifact %s is a primordial input", target)
			}
			// if all of the other artifacts are primordial or scheduled, the plan will be empty.
			// this lets the caller know to schedule the originating creatorRule. If the plan
			// ends up not empty, then there is a higher creatorRule and we must continue.
			continue
		}

		// the creatorRule exists
		if a.creatorRule.scheduled {
			// if it has already been scheduled, there is nothing to do.
			continue
		}
		// we have to check all of the creatorRule's inputs, meaning we are going up one level
		tmpPlan, err := dr.rebuild(a.creatorRule.inputs, level+1, graph)
		if err != nil {
			return nil, err
		} else if tmpPlan == nil {
			// we couldn't get higher; the creatorRule should be scheduled (except if it's because everything was already scheduled..)
			a.creatorRule.scheduled = true
			plan = append(plan, a.creatorRule)
		} else {
			// we went higher in the dependency graph, schedule the results
			plan = append(plan, tmpPlan...)
		}
	}
	return plan, nil
}

/* // The SimpleRebuilder builds a list of Rules required to bring up-to-date the targets without taking
// any previous runs into account. The resulting list of Rules makes up the partial diagram(s) linking
// the target(s) their primordial input(s).
type SimpleRebuilder struct {
}

func (sr *SimpleRebuilder) rebuild(targets []ArtifactTag, graph *Graph) ([]Rule, error) {
	if graph == nil {
		return nil, fmt.Errorf("graph is nil")
	}
	if targets == nil {
		return nil, nil
	}

	var p []Rule
	var todoList []ArtifactTag
	for _, target := range targets {
		a, ok := graph.artifacts[target]
		if !ok {
			return nil, fmt.Errorf("target artifact %s does not exist", target)
		}

		if a.creatorRule == nil {
			// Artifact is a primordial input
			continue
		}
		if a.creatorRule.scheduled {
			continue
		}
		if a.scheduled {
			continue
		}

		// schedule the Artifact and its creatorRule
		a.creatorRule.scheduled = true
		a.scheduled = true
		p = append(p, *a.creatorRule)

		// add inputs of the creatorRule to the todoList as we prepare to go up the dependency tree
		// check for duplicates
		for _, input := range a.creatorRule.inputs {
			if !slices.Contains(todoList, input) {
				todoList = append(todoList, input)
			}
		}
	}
	parent, err := sr.rebuild(todoList, graph)
	return append(p, parent...), err
}

func (*SimpleRebuilder) cleanup(graph *Graph) {
	for tag, a := range graph.artifacts {
		a.scheduled = false
		graph.artifacts[tag] = a
	}
	for idx, rule := range graph.rules {
		rule.scheduled = false
		graph.rules[idx] = rule
	}
}
*/
