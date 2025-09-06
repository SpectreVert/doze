package doze

import (
	"container/list"
	"crypto"
	_ "crypto/md5"
	"fmt"
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
// Inputs and Outputs contain only the Tags of the Artifacts.
// Its Hash is computed by digesting the input and output Tags, as well as the Procedure ID.
type Rule struct {
	Inputs, Outputs []ArtifactTag
	procID          ProcedureID

	Scheduled bool
}

// An Artifact represents a file which is processed, or created by Doze in the context of a build.
// Internally and to the operator, Artifacts are referred to by their ArtifactTag.
type Artifact struct {
	tag       ArtifactTag
	creator   *Rule   // The Rule which creates this Artifact. If nil, the Artifact is said to be 'primordial'.
	consumers []*Rule // The Rule(s) which depend on this Artifact to be run.
}

// An ArtifactTag represents the path on disk to an Artifact.
// Internally, its value is split between its file name, and its location.
// Its Hash is computed by unifying these two values as a NormalizedTag, which points to the real location of the Artifact on disk.
// Artifacts with the same NormalizedTag cannot be distinct.
type ArtifactTag struct {
	name, location string
}

/* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/

/* Graph */

func NewGraph() *Graph {
	return &Graph{
		rules:     make(map[string]*Rule),
		artifacts: make(map[string]*Artifact),
	}
}

// Resolve computes a list of Rules for Graph, ordered topologically based on their dependencies.
// This list is called the execution plan.
func (graph *Graph) Resolve() []string {
	// Algorith to compute the topological order of the Graph. (Kahn's Algorithm)
	// The hard thing to grasp is that a Rule makes up both nodes and edges.
	// Essentially, a node is a group of input or output Artifacts. An edge is the Rule that transforms them.

	// Make a list of all Rules whose input Artifacts do not have a creator Rule. These are called primordial Rules.
	var primordialRules = list.New()
PrimordialRulesLoop:
	for hash, rule := range graph.rules {
		for _, tag := range rule.Inputs {
			if graph.artifacts[tag.NormalizedTag()].creator != nil {
				continue PrimordialRulesLoop
			}
		}
		primordialRules.PushBack(hash)
	}

	var plan []string
	// While primordialRules is not empty
	for e := primordialRules.Front(); e != nil; e = primordialRules.Front() {
		ruleHash := primordialRules.Remove(e).(string)
		graph.rules[ruleHash].Scheduled = true
		plan = append(plan, ruleHash)

		// Iterate over each output Artifact of Rule `ruleHash` to inspect Rules which depend on them.
		for _, outputTag := range graph.rules[ruleHash].Outputs {
		CheckConsumerRules:
			// `consumerRule` is a Rule that depends on `outputTag` (representing an output Artifact of `ruleHash`).
			// `consumerRule` might have all its input Artifacts ready for consumption, meaning that these Artifacts either have no creator Rule, or that their
			// creator Rule is already scheduled (and in the plan). In that case, the `consumerRule` can be scheduled and added to the plan.
			for _, consumerRule := range graph.artifacts[outputTag.NormalizedTag()].consumers {
				for _, consumerTag := range consumerRule.Inputs {
					if graph.artifacts[consumerTag.NormalizedTag()].creator != nil && !graph.artifacts[consumerTag.NormalizedTag()].creator.Scheduled {
						continue CheckConsumerRules
					}
				}
				// It's also possible that this Rule was already scheduled during the same iterator over `consumers` Rule of `outputTag`.
				if !graph.rules[consumerRule.Hash()].Scheduled {
					graph.rules[consumerRule.Hash()].Scheduled = true
					primordialRules.PushBack(consumerRule.Hash())
				}
			}
		}
	}

	return plan
}

// AddRule registers a new Rule with the Graph.
//
// `inputs` and `outputs` are lists of file names (not Normalized Tags!)
//
// `inputLocation` and `outputLocation` point to a directory containing, respectively, the input files and the output files, relative to the Dozefile.
// `inputLocation` and `outputLocation` may be empty, if the files are in the same directory as the Dozefile.
//
// `procID` is the Procedure ID that will be used in the Rule.
//
// Returns an error if something went wrong, and the Rule could not be registered.
//
// ---------------------------------------------------------
// TODO: Check that an input is also not provided as output.
func (graph *Graph) AddRule(
	inputs, outputs []string,
	procID ProcedureID,
	inputLocation, outputLocation string,
) error {
	if len(inputs) == 0 {
		return fmt.Errorf("no inputs provided")
	}
	if len(outputs) == 0 {
		return fmt.Errorf("no outputs provided")
	}

	rule := &Rule{
		procID: procID,
	}

	// For each input, create the Artifact, if it doesn't exist yet, otherwise fetch it.
	// Add `rule` as a consumer Rule of the Artifact.
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
		rule.Inputs = append(rule.Inputs, newTag)
		graph.artifacts[newTag.NormalizedTag()].consumers = append(graph.artifacts[newTag.NormalizedTag()].consumers, rule)
	}

	// For each output, create the Artifact, if it doesn't exist yet, otherwise fetch it.
	// Check that the Artifact does not already have a creator Rule.
	// Add `rule` as the creator Rule of the Artifact.
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
		rule.Outputs = append(rule.Outputs, newTag)
	}

	// The Hash will be used to check for duplicates, and as a key to the Rule in the Graph.
	ruleHash := rule.Hash()

	_, ok := graph.rules[ruleHash]
	if ok {
		return fmt.Errorf("rule already exists with the same data")
	}

	graph.rules[ruleHash] = rule
	return nil
}

/* Rule */

// The Hash function of a Rule. Obviously, must be deterministic.
// Takes into account the input and output ArtifactTags, and the ProcedureID.
func (rule *Rule) Hash() string {
	hash := crypto.MD5.New()

	slices.SortFunc(rule.Inputs, CompareArtifactTags)
	for _, inputTag := range rule.Inputs {
		hash.Write([]byte(inputTag.NormalizedTag()))
	}
	slices.SortFunc(rule.Outputs, CompareArtifactTags)
	for _, outputTag := range rule.Outputs {
		hash.Write([]byte(outputTag.NormalizedTag()))
	}
	hash.Write([]byte(rule.procID))

	return fmt.Sprintf("%x", hash.Sum(nil))
}

/* ArtifactTag */

// A NormalizedTag is the actual path of an Artifact on disk.
// It's used as the key to an Artifact object in the Graph.
func (tag *ArtifactTag) NormalizedTag() string {
	return path.Join(tag.location, tag.name)
}

// Returns a deterministic ordering for ArtifactTags, used when hashing a Rule.
func CompareArtifactTags(first, second ArtifactTag) int {
	return strings.Compare(first.NormalizedTag(), second.NormalizedTag())
}
