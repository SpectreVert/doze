package doze

import (
	"fmt"
)

// A Rebuilder is given a list of Artifacts to bring up to date and computes a list of Rules to execute
// or inspect to achieve the wanted goal.
type Rebuilder interface {
	cleanup(*Graph) // should be called before each rebuild
	rebuild([]ArtifactTag, *Graph) ([]Rule, error)
}

type DeepRebuilder struct {
	nothingNewFound bool
}

func (dr *DeepRebuilder) rebuild(targets []ArtifactTag, level int, g *Graph) ([]*Rule, error) {
	if g == nil {
		return nil, fmt.Errorf("graph is nil")
	}
	if targets == nil {
		return nil, nil
	}
	dr.nothingNewFound = true
	var plan []*Rule
	for _, target := range targets {
		a, ok := g.artifacts[target]
		if !ok {
			return nil, fmt.Errorf("target artifact %s does not exist", target)
		}
		if a.creatorRule == nil {
			// `a` is a primordial input
			if level == 0 {
				// if level == 0, `a` *is* an original target also, which is an error
				return nil, fmt.Errorf("target artifact %s is a primordial input", target)
			}
			// if all of the other artifacts are primordial or scheduled, the plan will be empty.
			// this lets the caller know to schedule the originating creatorRule. If the plan
			// ends up not empty, then there is a higher creatorRule and we must continue.
			dr.nothingNewFound = false
			continue
		}

		// the creatorRule exists
		if a.creatorRule.scheduled {
			// if it has already been scheduled, there is nothing to do.
			continue
		}
		// we have to check all of the creatorRule's inputs, meaning we are going up one level
		tmpPlan, err := dr.rebuild(a.creatorRule.inputs, level+1, g)
		if err != nil {
			return nil, err
		}
		if tmpPlan == nil {
			// we couldn't get higher; the creatorRule should be scheduled
			if dr.nothingNewFound {
				// except if nothing new was found in the previous run, which means we reached only rules that were already scheduled
				continue
			}
			a.creatorRule.scheduled = true
			plan = append(plan, a.creatorRule)
		} else {
			// we went higher in the dependency graph, schedule what we got back
			plan = append(plan, tmpPlan...)
		}
	}
	return plan, nil
}
