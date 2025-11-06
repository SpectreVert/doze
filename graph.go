package doze

import (
	"container/list"
	_ "crypto/md5"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// A Graph contains Rules and Artifacts of the current build.
// The Rules and Artifacts are mapped using a Hash computed from their contents.
type Graph struct {
	rules      map[string]*Rule
	artifacts  map[string]*Artifact
	cachedHash string
}

// A Rule represents an action that processes input Artifacts into output Artifacts by executing a Procedure.
// Inputs and Outputs contain only the Tags of the Artifacts.
// Its Hash is computed by digesting the names of the input and output Artifacts (their ArtifactTag), as well as the ProcedureID.
// Its Checksum is computed by digesting the contents of the input Artifacts.
type Rule struct {
	Inputs, Outputs []ArtifactTag
	procID          ProcedureID
	cachedHash      string

	// TODO: REMOVE and replace with a local map of bools in Resolve
	Scheduled bool
}

// An Artifact represents a file which is processed, or created by Doze in the context of a build.
// Internally and to the operator, Artifacts are referred to by their ArtifactTag.
type Artifact struct {
	tag       ArtifactTag
	creator   *Rule   // The Rule which creates this Artifact. If nil, the Artifact is said to be 'primordial'.
	consumers []*Rule // The Rule(s) which depend on this Artifact to be run.

	// Exists: the artifact exists on disk.
	// Built: the artifact was built by a Rule execution.
	// Fetched: the artifact was fetched from the cache, not built by a Rule.
	// TODO: REMOVE
	Exists, Built, Fetched bool
}

// An ArtifactTag represents the path on disk to an Artifact.
// Internally, its value is split between its file name, and its location.
// Its Hash is computed by unifying these two values as a NormalizedTag, which points to the real location of the Artifact on disk.
// Artifacts with the same NormalizedTag cannot be distinct.
type ArtifactTag struct {
	name, location string
}

// The resolve mode is passed to Graph.Resolve to indicate which policy to apply when creating the plan.
type ResolveMode int

const (
	// Terse: schedule nothing if all primordial rules are up-to-date.
	TerseMode ResolveMode = iota
	// Full: schedule all rules from the graph.
	FullMode
	// Target: schedule only rules that bring a specific artifact up-to-date. NOT IMPLEMENTED YET.
	TargetMode
)

/* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/

/* Graph */

func NewGraph() *Graph {
	return &Graph{
		rules:     make(map[string]*Rule),
		artifacts: make(map[string]*Artifact),
	}
}

// Execute to completion a topologically sorted list of Rules (identified by their Hash),
// with the goal of bringing the Graph up-to-date.
// Executes synchronously and single-threadedly.
// Returns the number of executed rules, fetched rules, and an error if one happened.
func (graph *Graph) Execute(cache Cache) (int, int, error) {
	var executedRules = 0
	var fetchedRules = 0

	for _, ruleHash := range cache.Plan() {
		rule, ok := graph.rules[ruleHash]
		if !ok {
			return 0, 0, fmt.Errorf("rule [" + ruleHash + "] resolved for execution but doesn't exist")
		}

		if cache.HasArtifacts(graph.rules[ruleHash]) {
			fmt.Println("rule [cached]:", ruleHash)
			if err := cache.ProduceArtifacts(graph.rules[ruleHash]); err != nil {
				// FIXME: add cleanup
				return 0, 0, err
			}
			fetchedRules += 1
		} else {
			fmt.Println("rule:", ruleHash)
			rule.Execute()

			executedRules += 1

			if err := cache.StoreArtifacts(graph.rules[ruleHash]); err != nil {
				// FIXME: add cleanup
				return 0, 0, err
			}
		}
	}

	if err := cache.FlushRecords(); err != nil {
		return 0, 0, err
	}

	return executedRules, fetchedRules, nil
}

// Reset all Artifacts flags.
func (graph *Graph) SetArtifactStates() {
	for normalizedTag, artifact := range graph.artifacts {
		_, err := os.Stat(normalizedTag)
		artifact.Exists = (err == nil)
	}
}

// Resolve computes a list of Rules for Graph, ordered topologically based on their dependencies and store it as the Graph's plan.
//
// NOTE: This always resolves fully. That is, whatever the state of each Rule's outputs. we will always return the entire graph
// ordered topologically, and then have Resolve decide dynamically what to run... This is probably not what we want.
//
// What we could do instead, is:
// -terse (default): only schedule the descendants of primordial rules which are not up-to-date.
// -full: resolve all rules from the graph. Either run them to completion or use the cache.
// -target: only schedule rules that contribute to building a specific artifact.

// TODO: Use the cache. resolve should read the cache and execute should fetch/write to it.
func (graph *Graph) Resolve(cache Cache) {
	graph.SetArtifactStates()

	cache.ClearPlan()

	// Computes the topological order of the Graph. (Kahn's Algorithm)
	// The peculiar thing is that a Rule makes up both nodes and edges.
	// Essentially, a node is a group of input or output Artifacts. An edge is the Rule that transforms them.

	// Make an initial list of Rules whose input Artifacts do not have a creator Rule.
	// Also check if the rules were executed in the last run.
	// These are called primordial Rules.
	var rulesToInspect = list.New()
RulesToInspectLoop:
	for hash, rule := range graph.rules {
		for _, tag := range rule.Inputs {
			if graph.artifacts[tag.NormalizedTag()].creator != nil {
				continue RulesToInspectLoop
			}
			// TODO: check that this exists.
			// The function should maybe return errors.
		}

		if !cache.RuleInLastRun(rule) /* OR IF FULL MODE */ {
			graph.rules[hash].Scheduled = true
			rulesToInspect.PushBack(hash)
		}
		cache.RecordRule(rule)

		// GET TIMESTAMP
		// FOREACH RULE:
		//	IF RULE IS IN CACHE:
		//		DO NOT SCHEDULE
		//	ELSE RULE NOT IN CACHE:
		//		# the rule is new
		//		QUEUE ADDING THE RULE IN CACHE, WITH HASH, CHECKSUM
		//		SCHEDULE RULE
		//	QUEUE RECORD RULE IN CACHE
		//
		//  UPDATE ALL QUEUED RULES IN CACHE
		//
		// FOREACH RULE IN CACHE:
		//	IF TIMESTAMP NOT UPTODATE:
		//		DELETE RULE

	}

	// IF TERSE MODE:
	//		ONLY UPDATE RULES THAT CHANGED INPUTS
	//		OR IF ALL RULES CHANGED GOTO FULL MODE
	// IF FULL MODE:
	//		CLEAR THE PIMORDIAL CACHE

	// While rulesToInspect is not empty
	for e := rulesToInspect.Front(); e != nil; e = rulesToInspect.Front() {
		ruleHash := rulesToInspect.Remove(e).(string)
		cache.ScheduleRule(graph.rules[ruleHash])

		// Iterate over each output Artifact of Rule `ruleHash` to inspect Rules which depend on them.
		for _, outputTag := range graph.rules[ruleHash].Outputs {
		CheckConsumerRules:
			// `consumerRule` is a Rule that depends on `outputTag` (representing an output Artifact of `ruleHash`).
			// `consumerRule` might have all its input Artifacts ready for consumption, meaning that these Artifacts either have no creator Rule, or that their
			// creator Rule is already scheduled (and in the plan). In that case, the `consumerRule` can be scheduled and added to the plan.
			for _, consumerRule := range graph.artifacts[outputTag.NormalizedTag()].consumers {
				for _, consumerTag := range consumerRule.Inputs {
					if graph.artifacts[consumerTag.NormalizedTag()].Exists || consumerTag == outputTag {
						continue
					} else if graph.artifacts[consumerTag.NormalizedTag()].creator != nil && !graph.artifacts[consumerTag.NormalizedTag()].creator.Scheduled {
						continue CheckConsumerRules
					}
				}
				// It's also possible that this Rule was already scheduled as a creator rule, or during a previous iteration over consumer rules.
				if !graph.rules[consumerRule.Hash()].Scheduled {
					graph.rules[consumerRule.Hash()].Scheduled = true
					rulesToInspect.PushBack(consumerRule.Hash())
				}
			}
		}
	}
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

/* ArtifactTag */

// Returns the actual path of the Artifact on disk.
func (tag *ArtifactTag) Path() string {
	return filepath.Join(tag.location, tag.name)
}

// A NormalizedTag is the string identifier used by doze internally as the key
// to the Artifact object in the Graph.
func (tag *ArtifactTag) NormalizedTag() string {
	return path.Join(tag.location, tag.name)
}

// Returns a deterministic ordering for ArtifactTags, used when hashing a Rule.
func CompareArtifactTags(first, second ArtifactTag) int {
	return strings.Compare(first.NormalizedTag(), second.NormalizedTag())
}
