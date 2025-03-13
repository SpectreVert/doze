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

func (g *Graph) createRule(inputs []ArtifactTag, outputs []ArtifactTag /* prod */, ruleLocation string) error {
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

// TODO: make a real setExisting func
func (g *Graph) setExisting() {
	for key, artifact := range g.artifacts {
		realTag := artifact.tag.InternalTag()
		_, err := os.Stat(realTag)
		artifact.Exists = (err == nil)
		g.artifacts[key] = artifact
	}
}

func (g *Graph) reset() {
	for key, artifact := range g.artifacts {
		artifact.Touched = false
		artifact.Processed = false
		g.artifacts[key] = artifact
	}
}

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

// returns a list of rule hashes from the rules of the graph which should
// be passed to execute for scheduling.
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
	for _, hash := range plan {
		rule := g.rules[hash]

		// TODO check the cache for outputs linked to this combination of inputs

		// run the rule
		fmt.Println("Running rule", hash)

		// mark the inputs as Processed (?)
		for _, tag := range rule.inputs {
			artifact := g.artifacts[tag.InternalTag()]
			artifact.Processed = true
			g.artifacts[tag.InternalTag()] = artifact
		}

		// mark the outputs as Touched
		for _, tag := range rule.outputs {
			artifact := g.artifacts[tag.InternalTag()]
			artifact.Touched = true
			g.artifacts[tag.InternalTag()] = artifact
		}

		rules := g.resolve()
		if len(rules) > 0 {
			g.execute(rules)
		}
	}
}

func main() {
	location := path.Clean("./samples/sample-dir.out")
	g := Graph{
		rules:     make(map[string]Rule),
		artifacts: make(map[string]Artifact),
	}

	err := g.createRule([]ArtifactTag{ArtifactTag{"parse.y", ""}}, []ArtifactTag{ArtifactTag{"parse.c", ""}, ArtifactTag{"parse.h", ""}}, location)
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

	err = g.createRule([]ArtifactTag{ArtifactTag{"parse.o", ""}, ArtifactTag{"main.o", ""}}, []ArtifactTag{ArtifactTag{"exe", ""}}, location)
	if err != nil {
		fmt.Println("create rule:", err)
		return
	}

	g.setExisting()
	g.touchArtifacts([]ArtifactTag{ArtifactTag{"main.c", location}})

	rules := g.resolve()
	g.execute(rules)
}

// 1. Load the Dozefile
// 2. For all of the referenced artifacts, check which exist or not, or which have been modified.
// 3. For each modified or missing artifacts, schedule the creator-rule
// 4. Run the rules and interupt rules which cannot be run yet.
