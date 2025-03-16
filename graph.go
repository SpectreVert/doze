package main

import (
	"crypto"
	_ "crypto/md5"
	"fmt"
	"os"
	"path"
	"slices"
	"strings"
)

type ArtifactTag struct {
	ShortTag string
	Location string
}

// this is the real tag that points to the artifact in the graph internal map
func (a *ArtifactTag) InternalTag() string {
	return strings.ToLower(a.RealLocation())
}

// this is the real location of the artifact on disk
func (a *ArtifactTag) RealLocation() string {
	return path.Join(a.Location, a.ShortTag)
}

func CompArtifacts(first ArtifactTag, second ArtifactTag) int {
	return strings.Compare(first.InternalTag(), second.InternalTag())
}

type Artifact struct {
	tag     ArtifactTag
	creator *Rule

	Exists    bool
	Touched   bool
	Processed bool
}

type Rule struct {
	inputs  []ArtifactTag
	outputs []ArtifactTag
	// procedure
}

func (r *Rule) Hash() string {
	hash := crypto.MD5.New()

	slices.SortFunc(r.inputs, CompArtifacts)
	for _, inputTag := range r.inputs {
		hash.Write([]byte(inputTag.InternalTag()))
	}

	slices.SortFunc(r.outputs, CompArtifacts)
	for _, outputTag := range r.outputs {
		hash.Write([]byte(outputTag.InternalTag()))
	}

	return fmt.Sprintf("%x", hash.Sum(nil))
}

type Graph struct {
	rules     map[string]Rule
	artifacts map[string]Artifact
}

// Store a new Rule in the Graph.
// It creates or updates Artifacts based on the input and output ArtifactTags.
// The rule is stored in a map in which the key is a hash of all its inputs/outputs Tags.
func (g *Graph) createRule(inputs []ArtifactTag, outputs []ArtifactTag /* proc */, ruleLocation string) error {
	if len(inputs) == 0 {
		return fmt.Errorf("no inputs were provided?!")
	}
	if len(outputs) == 0 {
		return fmt.Errorf("no outputs were provided?!")
	}

	r := Rule{
		inputs:  inputs,
		outputs: outputs,
	}
	ruleHash := r.Hash()

	// duplicate check
	_, exists := g.rules[ruleHash]
	if exists {
		return fmt.Errorf("rule already declared")
	}

	// same but for the artifacts
	for idx, tag := range r.inputs {
		tag.Location = path.Join(ruleLocation, tag.Location)
		_, exists := g.artifacts[tag.InternalTag()]
		if !exists {
			g.artifacts[tag.InternalTag()] = Artifact{
				tag: tag,
			}
		}
		r.inputs[idx] = tag // since we modified tag.Location we must re-set it
	}
	for idx, tag := range r.outputs {
		tag.Location = path.Join(ruleLocation, tag.Location)
		artifact, exists := g.artifacts[tag.InternalTag()]
		if !exists {
			g.artifacts[tag.InternalTag()] = Artifact{
				tag:     tag,
				creator: &r,
			}
		} else if artifact.creator != nil {
			// Artifact was already used as output previously.
			return fmt.Errorf("artifact %s cannot be output two different times", tag.InternalTag())
		} else {
			// Artifact already exists but must be updated.
			artifact.creator = &r
			g.artifacts[tag.InternalTag()] = artifact
		}
		r.outputs[idx] = tag // since we modified tag.Location we must re-set it
	}

	g.rules[ruleHash] = r

	return nil
}

// Set the Existing marker on all Artifacts depending if they are found on disk or not.
func (g *Graph) setExisting() {
	for internalTag, artifact := range g.artifacts {
		_, err := os.Stat(internalTag)
		artifact.Exists = (err == nil)
		g.artifacts[internalTag] = artifact
	}
}

// Reset Touched and Processed markers from the Graph.
func (g *Graph) reset() {
	for key, artifact := range g.artifacts {
		artifact.Touched = false
		artifact.Processed = false
		g.artifacts[key] = artifact
	}
}

// Mark touchedArtifacts as having been modified outside of the build system
func (g *Graph) touchArtifacts(touchedArtifacts []ArtifactTag) {
	for _, tag := range touchedArtifacts {
		artifact, ok := g.artifacts[tag.InternalTag()]
		if !ok {
			panic("artifact " + tag.RealLocation() + " does not exist")
		}
		fmt.Println("found a modified file:", tag.InternalTag())
		artifact.Touched = true
		g.artifacts[tag.InternalTag()] = artifact
	}
}

// Iterates on the bucket of Rules stored in the Graph.
// For each Rule, we inspect its input(s) then its output(s). If any input is found to have been
// modified since the last doze run the Rule is added to a list and the next Rule is immediately
// inspected.
// Otherwise its outputs are also checked and if any are missing on disk the Rule is added to the list
// and the next Rule is immediately inspected.
// The final list of Rules only contains the hashes from the Rules, which serve as keys in the
// internal Graph map for the Rules.
// The order of this list of hashes is deemed non-deterministic.
func (g *Graph) resolve() []string {
	var ruleHashes []string

RuleLoop:
	for key, rule := range g.rules {
		for _, tag := range rule.inputs {
			if g.artifacts[tag.InternalTag()].Touched && !g.artifacts[tag.InternalTag()].Processed {
				// Rule has a modified input, which has not been processed yet. Schedule the rule.
				if !slices.Contains(ruleHashes, key) {
					ruleHashes = append(ruleHashes, key)
				}
				continue RuleLoop
			}
		}
		for _, tag := range rule.outputs {
			if !g.artifacts[tag.InternalTag()].Exists {
				// Rule is missing at least an output. Schedule the rule.
				if !slices.Contains(ruleHashes, key) {
					ruleHashes = append(ruleHashes, key)
				}
				continue RuleLoop
			}
		}
	}

	return ruleHashes
}

func (g *Graph) execute(plan []string) {
RuleLoop:
	for _, hash := range plan {
		rule := g.rules[hash]

		// check if we can actually run the rule yet
		// that is, check if all the inputs exist
		for _, tag := range rule.inputs {
			artifact := g.artifacts[tag.InternalTag()]
			if !artifact.Exists {
				fmt.Println("input doesnt exist yet in", hash)
				continue RuleLoop
			}
		}

		// run the rule
		fmt.Println("Running rule", hash)

		// mark the inputs as Processed
		for _, tag := range rule.inputs {
			artifact := g.artifacts[tag.InternalTag()]
			artifact.Processed = true
			g.artifacts[tag.InternalTag()] = artifact
		}

		// mark the outputs as Touched and Existing
		for _, tag := range rule.outputs {
			artifact := g.artifacts[tag.InternalTag()]
			artifact.Touched = true
			artifact.Exists = true
			g.artifacts[tag.InternalTag()] = artifact
		}

	}
	rules := g.resolve()
	if len(rules) > 0 {
		g.execute(rules)
	}
}

func main() {
	location := path.Clean("./samples/sample-dir.in")
	g := Graph{
		rules:     make(map[string]Rule),
		artifacts: make(map[string]Artifact),
	}

	err := g.createRule([]ArtifactTag{ArtifactTag{"parse.o", ""}, ArtifactTag{"main.o", ""}}, []ArtifactTag{ArtifactTag{"exe", ""}}, location)
	if err != nil {
		fmt.Println("create rule:", err)
		return
	}

	err = g.createRule([]ArtifactTag{ArtifactTag{"parse.h", ""}, ArtifactTag{"parse.c", ""}}, []ArtifactTag{ArtifactTag{"parse.o", ""}}, location)
	if err != nil {
		fmt.Println("create rule:", err)
		return
	}

	err = g.createRule([]ArtifactTag{ArtifactTag{"parse.h", ""}, ArtifactTag{"main.c", ""}}, []ArtifactTag{ArtifactTag{"main.o", ""}}, location)
	if err != nil {
		fmt.Println("create rule:", err)
		return
	}

	err = g.createRule([]ArtifactTag{ArtifactTag{"parse.y", ""}}, []ArtifactTag{ArtifactTag{"parse.c", ""}, ArtifactTag{"parse.h", ""}}, location)
	if err != nil {
		fmt.Println("create rule:", err)
		return
	}

	g.setExisting()
	g.touchArtifacts([]ArtifactTag{ArtifactTag{"main.c", location}})

	fmt.Println(g.rules)

	rules := g.resolve()
	g.execute(rules)
}

// 1. Load the Dozefile
// 2. For all of the referenced artifacts, check which exist or not, or which have been modified.
// 3. For each modified or missing artifacts, schedule the creator-rule
// 4. Run the rules and interupt rules which cannot be run yet.
