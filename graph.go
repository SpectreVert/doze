package doze

import (
	"crypto"
	_ "crypto/md5"
	"fmt"
	"os"
	"path"
	"slices"
	"strings"
)

// A Graph contains Rules and Artifacts of the current build.
// The Rules and Artifacts are mapped using a Hash computed from their contents.
type Graph struct {
	rules     map[string]*Rule
	artifacts map[string]*Artifact
}

// A Rule represents an action that processes input Artifacts into output Artifacts by executing a Procedure.
// Its Hash is computer by digesting
type Rule struct {
	inputs, outputs []ArtifactTag
	// procedure Procedure
	Processed bool
}

// An Artifact represents a file which is processed or created by Doze in the context of a build.
// Internally and to the operator, Artifacts are represented by their ArtifactTag.
// The Artifact keeps track of its own status:
//   - Exists is true if the Artifact is found to exist on the disk.
//   - Modified is true if the Artifact was touched by the operator since the last build or
//     if the Artifact was created by a Doze rule in the current build.
type Artifact struct {
	tag     ArtifactTag
	creator *Rule

	Exists, Modified bool
}

// An ArtifactTag represents the path on disk to an Artifact.
// Internally its value is split into its name and its location.
// Its Hash is computer by digesting the normalized path made of the location and the name,
// thus reconstructing its real path on disk. Artifacts therefore cannot have duplicates.
type ArtifactTag struct {
	name, location string
}

/* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/

/* Graph */

// Execute to completion a non-sorted list of Rules, passed by their Hashes, with the goal
// of bringing the Graph up-to-date. Rules which are not ready for execution yet are postponed.
// Before returning the function resolves the Graph again, taking into account the updated status
// of the Rules that were exected, and calls itself recursively with the new list of resolved Rules,
// if it's not empty.
func (graph *Graph) Execute(ruleHashes []string) {
RuleLoop:
	for _, hash := range ruleHashes {
		rule, ok := graph.rules[hash]
		if !ok {
			panic("rule '" + hash + "' resolved for execution but doesn't exist")
		}

		// if any input is missing, the Rule has to be postponed
		for _, tag := range rule.inputs {
			artifact := graph.artifacts[tag.NormalizedTag()]
			if !artifact.Exists {
				// if the artifact doesn't have any creatorRule and it doesn't exist that's not good...
				if artifact.creator == nil {
					panic("artifact '" + tag.NormalizedTag() + "' doesn't exist and doesn't have a creator Rule")
				}
				continue RuleLoop
			}
		}

		fmt.Println("Running rule", hash)

		// mark the Rule as Processed
		rule.Processed = true

		// mark its outputs as Existing and Modified
		for _, tag := range rule.outputs {
			artifact := graph.artifacts[tag.NormalizedTag()]
			artifact.Exists = true
			artifact.Modified = true
		}
	}
	newRuleHashes := graph.Resolve()
	if len(newRuleHashes) > 0 {
		graph.Execute(newRuleHashes)
	}
}

// Resolve and return a non-sorted list of Hashes of Rules. These Rules should be executed to bring
// the Graph up-to-date. Of course, if the returned list is nil, then there is nothing to do.
func (graph *Graph) Resolve() []string {
	var ruleHashes []string

RuleLoop:
	for hash, rule := range graph.rules {
		if rule.Processed {
			// Rule's been processed already. Go next.
			continue
		}
		for _, tag := range rule.inputs {
			if graph.artifacts[tag.NormalizedTag()].Modified {
				// Rule has a modified input. Schedule the Rule.
				if !slices.Contains(ruleHashes, hash) {
					ruleHashes = append(ruleHashes, hash)
				}
				continue RuleLoop
			}
		}
		for _, tag := range rule.outputs {
			if !graph.artifacts[tag.NormalizedTag()].Exists {
				// Rule has a non-existing output. Schedule the Rule.
				if !slices.Contains(ruleHashes, hash) {
					ruleHashes = append(ruleHashes, hash)
				}
				continue RuleLoop
			}
		}
	}
	return ruleHashes
}

// Registers a new Rule with the Graph.
// input and outputs are Artifact names.
// inputLocation and outputLocation are Artifact locations.
// Returns an error if something went wrong and the Rule could not be registered.
// @todo add support for Procedure.
func (graph *Graph) AddRule(
	inputs, outputs []string,
	/* procedure, */
	inputLocation, outputLocation string,
) error {
	if len(inputs) == 0 {
		return fmt.Errorf("no inputs provided")
	}
	if len(outputs) == 0 {
		return fmt.Errorf("no outputs provided")
	}

	// @todo set Procedure here.
	rule := &Rule{}

	// create, then register input Artifacts, if they don't already exist.
	for _, name := range inputs {
		newTag := ArtifactTag{
			name,
			inputLocation,
		}
		_, ok := graph.artifacts[newTag.NormalizedTag()]
		if !ok {
			graph.artifacts[newTag.NormalizedTag()] = &Artifact{
				tag: newTag,
			}
		}
		rule.inputs = append(rule.inputs, newTag)
	}

	// ditto for output Artifacts, but also handle its creator Rule,
	// and check that the Artifact is not already an output of another Rule.
	for _, name := range outputs {
		newTag := ArtifactTag{
			name,
			outputLocation,
		}
		artifact, ok := graph.artifacts[newTag.NormalizedTag()]
		if !ok {
			graph.artifacts[newTag.NormalizedTag()] = &Artifact{
				tag:     newTag,
				creator: rule,
			}
		} else if artifact.creator != nil {
			return fmt.Errorf("artifact '%v' cannot be output by more than one rule", newTag.NormalizedTag())
		} else {
			artifact.creator = rule
		}
		rule.outputs = append(rule.outputs, newTag)
	}

	// the Hash will be used to check for duplicates and to store the Rule in the Graph.
	ruleHash := rule.Hash()

	// duplicate check of the Rule.
	_, ok := graph.rules[ruleHash]
	if ok {
		return fmt.Errorf("rule already exists with the same data")
	}

	graph.rules[ruleHash] = rule
	return nil
}

// Set the Exists flag to true for Artifacts which exist on disk.
func (graph *Graph) MarkArtifactsAsExisting() {
	for normalizedTag, artifact := range graph.artifacts {
		_, err := os.Stat(normalizedTag)
		artifact.Exists = (err == nil) // @robustness
	}
}

// Set the Modified flag to true for Artifacts whose tags are passed.
func (graph *Graph) markArtifactsAsModified(tags []ArtifactTag) {
	for _, tag := range tags {
		artifact, ok := graph.artifacts[tag.NormalizedTag()]
		if !ok {
			panic("artifact '" + tag.NormalizedTag() + "'' does not exist")
		}
		artifact.Modified = true
	}
}

// Reset to false the Processed status of all Rules. This is run after every build.
func (graph *Graph) resetProcessedRules() {
	for _, rule := range graph.rules {
		rule.Processed = false
	}
}

// Reset to false the Modified status of all Artifacts. This is run after every build.
func (graph *Graph) resetModifiedArtifacts() {
	for _, artifact := range graph.artifacts {
		artifact.Modified = false
	}
}

func NewGraph() *Graph {
	return &Graph{
		rules:     make(map[string]*Rule),
		artifacts: make(map[string]*Artifact),
	}
}

/* Rule */

// The Hash function of a Rule. Obviously, must be deterministic.
// Takes into account the ArtifactTags.
// @todo should also take into account the Procedure. Need to think about how.
func (rule *Rule) Hash() string {
	hash := crypto.MD5.New()

	slices.SortFunc(rule.inputs, CompareArtifactTags)
	for _, inputTag := range rule.inputs {
		hash.Write([]byte(inputTag.NormalizedTag()))
	}

	slices.SortFunc(rule.outputs, CompareArtifactTags)
	for _, outputTag := range rule.outputs {
		hash.Write([]byte(outputTag.NormalizedTag()))
	}

	return fmt.Sprintf("%x", hash.Sum(nil))
}

/* ArtifactTag */

// This is what the Graph uses as a key to an Artifact object.
// It's also the actual path of the Artifact on disk.
func (tag *ArtifactTag) NormalizedTag() string {
	return path.Join(tag.location, tag.name)
}

// Required for respecting a deterministic order when hashing a Rule.
func CompareArtifactTags(first, second ArtifactTag) int {
	return strings.Compare(first.NormalizedTag(), second.NormalizedTag())
}
