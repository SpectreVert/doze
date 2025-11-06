package doze

import (
	"crypto"
	_ "crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// Stores and retrieves Artifacts, used by and created from Rules.
type ArtifactCache interface {
	// Return `true` if the Rule outputs are stored in the Cache, otherwise return `false`.
	HasArtifacts(rule *Rule) bool
	// Store the output Artifacts of a Rule in the Cache.
	// The Rule must have been executed already; both inputs and outputs must exist on disk.
	StoreArtifacts(rule *Rule) error
	// Fetch the output Artifacts of the Rule from the Cache and place them into the working directory.
	// Return an error if an Artifact failed to produce.
	ProduceArtifacts(rule *Rule) error
}

// Keeps a record of primordial Rules executed during the previous run.
type PrimordialCache interface {
	// Check if a primordial rule is in the cache, and if the content of its inputs has changed.
	RuleInLastRun(rule *Rule) bool

	// Queue a record of the rule to the cache.
	RecordRule(rule *Rule)

	// - Delete the old primordial rules (that do not match the new timestamp)
	// - Add the new primordial rules (which needed to be updated, or have been run for the first time)
	FlushRecords() error
}

// Keeps track of which rules should be executed.
type PlanCache interface {
	// Return the plan.
	Plan() []string
	// Reset the internal plan queue.
	ClearPlan()

	// Schedule a Rule in the plan queue.
	ScheduleRule(rule *Rule)
}

type Cache interface {
	ArtifactCache
	PrimordialCache
	PlanCache
}

/* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/

/* Local directory cache */
// The schema of such a cache is as follows:
//
// .doze/						<-	Default name of the cache directory.
// (PRIMORDIAL CACHE)
//		last_run/						<- contains the primordial rules of the last run.
//			aabbccddee_abcdeffedcba		<- hash of the rule + checksum of the inputs of the rule
//			ffee33bb22_74d88f344eff
//
// (ARTIFACT CACHE)
//		df1833d6159.../			<-	Hash of a rule.
//									Any change to rule input/outputs/procedure results in a different Hash.
//									This directory is used to store the hash of inputs to this rule and the
//									different outputs produced by the rule.
//			74d887f344b99.../	<-  Checksum of the contents of the inputs of this rule, at a specific point in time.
//									This folder is used to store the actual output artifacts of the rule with this particular
//									set of inputs.
//
//				main.o			<-  Artifact produced by the rule with inputs with a specific content.
//
//			070789df76f25.../	<-	Another checksum of the contents of the inputs of this rule, at another point in time.
//
//				main.o			<-  Artifact produced by the same rule with inputs of a different content.

type LocalCache struct {
	Dir string

	plan          []string
	recordedRules []string
}

func NewLocalCache( /* TODO add config */ ) *LocalCache {
	cache := &LocalCache{
		Dir: ".doze",
	}
	// Create the directory if it doesn't exist yet.
	if err := os.MkdirAll(cache.Dir, 0775); err != nil { // os.ModeDir ?
		log.Fatalf("Failed to create root cache directory %s: %s", cache.Dir, err)
	}
	return cache
}

// Store a Rule inside the local directory cache.
// The rule must have Inputs and Outputs provided, and the Output Artifacts must exist in the working directory.
// The inputs are checksumed into a directory name, that directory is used to hold the outputs of the rules.
func (cache *LocalCache) StoreArtifacts(rule *Rule) error {
	ruleDir := filepath.Join(cache.Dir, rule.Hash())

	// Create the Rule cache directory if it does not exist yet.
	if _, err := os.Stat(ruleDir); err != nil {
		if err = os.Mkdir(ruleDir, 0775); err != nil {
			return fmt.Errorf("Failed to create rule cache directory %s: %s", ruleDir, err)
		}
	}

	// Compute the checksum of the inputs Artifacts.
	inputsHash := crypto.MD5.New()
	for _, inputTag := range rule.Inputs {
		artifactPath := inputTag.Path()
		content, err := os.ReadFile(artifactPath)
		if err != nil {
			return fmt.Errorf("Failed to read contents from artifact %s: %s", artifactPath, err)
		}
		inputsHash.Write(content)
	}
	checksum := hex.EncodeToString(inputsHash.Sum(nil))

	// Use the checksum of the input Artifacts to create the outputs directory.
	// Check that the Rule's inputs / outputs are not already cached.
	outputsDir := filepath.Join(ruleDir, checksum)
	if _, err := os.Stat(outputsDir); err == nil {
		// FIXME maybe this should return an error instead
		return fmt.Errorf("Attempt to store a rule which already exists in the cache: %s", outputsDir)
	}

	// Create the outputs directory cache, using the checksum of the input Artifacts as the name.
	if err := os.Mkdir(outputsDir, 0775); err != nil {
		return fmt.Errorf("Failed to create outputs cache directory %s: %s", outputsDir, err)
	}

	// Create hardlinks to all of the output files.
	for _, outputTag := range rule.Outputs {
		linkPath := filepath.Join(outputsDir, outputTag.Path())
		if err := os.Link(outputTag.Path(), linkPath); err != nil {
			// FIXME: best effort cleanup: remove all files previously created and the outputsDir
			return fmt.Errorf("Failed to create hardlink from %s to %s: %s", outputTag.Path(), linkPath, err)
		}
	}

	return nil
}

func (cache *LocalCache) HasArtifacts(rule *Rule) bool {
	ruleDir := filepath.Join(cache.Dir, rule.Hash())

	// Compute the checksum of the input Artifacts.
	inputsHash := crypto.MD5.New()
	for _, inputTag := range rule.Inputs {
		artifactPath := inputTag.Path()
		content, err := os.ReadFile(artifactPath)
		if err != nil {
			log.Fatalf("Failed to read contents from artifact %s: %s", artifactPath, err)
		}
		inputsHash.Write(content)
	}
	checksum := hex.EncodeToString(inputsHash.Sum(nil))

	// Check if the outputs directory exists in the cache.
	outputsDir := filepath.Join(ruleDir, checksum)
	if _, err := os.Stat(outputsDir); err != nil {
		return false
	}

	return true
}

func (cache *LocalCache) ProduceArtifacts(rule *Rule) error {
	ruleDir := filepath.Join(cache.Dir, rule.Hash())

	// Compute the checksum of the input Artifacts.
	inputsHash := crypto.MD5.New()
	for _, inputTag := range rule.Inputs {
		artifactPath := inputTag.Path()
		content, err := os.ReadFile(artifactPath)
		if err != nil {
			return fmt.Errorf("Failed to read contents from artifact %s: %s", artifactPath, err)
		}
		inputsHash.Write(content)
	}
	checksum := hex.EncodeToString(inputsHash.Sum(nil))

	// Check if the outputs directory exists in the cache.
	outputsDir := filepath.Join(ruleDir, checksum)
	if _, err := os.Stat(outputsDir); err != nil {
		return err
	}

	// Create hardlinks from each cached output to their respective location.
	for _, outputTag := range rule.Outputs {
		cachedPath := filepath.Join(outputsDir, outputTag.Path())

		// The output might actually already be there.
		// FIXME: can we check this before producing Artifacts?
		if outputInfo, err := os.Stat(outputTag.Path()); err == nil {
			if cachedOutputInfo, err := os.Stat(cachedPath); err == nil {
				if os.SameFile(outputInfo, cachedOutputInfo) {
					fmt.Printf("     [skipped]: %s\n", outputTag.NormalizedTag())
					continue
				}
			}
			fmt.Printf("     [cleaned]: %s\n", outputTag.NormalizedTag())
			if err = os.Remove(outputTag.Path()); err != nil {
				return fmt.Errorf("Failed to remove outdated output %s: %s", outputTag.NormalizedTag(), err)
			}
		}

		// FIXME: again, this assumes that doze is being invoked from the same directory as the Dozefile,
		// and that the Dozefile expresses file relative to it.
		if err := os.Link(cachedPath, outputTag.Path()); err != nil {
			// FIXME: use a temporary directory instead, so that we don't clog the working directory with stray files.
			return fmt.Errorf("Failed to create hardlink from %s to %s: %s", cachedPath, outputTag.Path(), err)
		}
	}

	return nil
}

func (cache *LocalCache) RuleInLastRun(rule *Rule) bool {
	lastRunDir := filepath.Join(cache.Dir, "last_run")
	if _, err := os.Stat(lastRunDir); err != nil {
		return false
	}

	// Compute the checksum of the inputs Artifacts.
	inputsHash := crypto.MD5.New()
	for _, inputTag := range rule.Inputs {
		artifactPath := inputTag.Path()
		content, err := os.ReadFile(artifactPath)
		if err != nil {
			log.Fatalf("Failed to read contents from artifact %s: %s", artifactPath, err)
		}
		inputsHash.Write(content)
	}
	checksum := hex.EncodeToString(inputsHash.Sum(nil))
	ruleID := rule.Hash() + "_" + checksum

	ruleDir := filepath.Join(lastRunDir, ruleID)
	if _, err := os.Stat(ruleDir); err == nil {
		return true
	}

	fmt.Printf("Rule not in last_run: %s\n", ruleID)

	return false
}

func (cache *LocalCache) RecordRule(rule *Rule) {
	// Compute the checksum of the inputs Artifacts.
	inputsHash := crypto.MD5.New()
	for _, inputTag := range rule.Inputs {
		artifactPath := inputTag.Path()
		content, err := os.ReadFile(artifactPath)
		if err != nil {
			log.Fatalf("Failed to read contents from artifact %s: %s", artifactPath, err)
		}
		inputsHash.Write(content)
	}
	checksum := hex.EncodeToString(inputsHash.Sum(nil))
	ruleID := rule.Hash() + "_" + checksum

	cache.recordedRules = append(cache.recordedRules, ruleID)
}

func (cache *LocalCache) FlushRecords() error {
	lastRunDir := filepath.Join(cache.Dir, "last_run")
	if _, err := os.Stat(lastRunDir); err != nil {
		if err = os.Mkdir(lastRunDir, 0775); err != nil {
			return fmt.Errorf("Failed to create last run cache directory %s: %s", lastRunDir, err)
		}
	}

	// Delete all the previously recorded rules.
	lastRunDirEntries, err := os.ReadDir(lastRunDir)
	if err != nil {
		return fmt.Errorf("Failed to read last run cache directory %s :%s", lastRunDir, err)
	}
	for _, entry := range lastRunDirEntries {
		path := filepath.Join(lastRunDir, entry.Name())
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("Failed to delete last run cache entry %s: %s", path, err)
		}
	}

	// Write the new recorded rules.
	for _, ruleID := range cache.recordedRules {
		ruleDir := filepath.Join(lastRunDir, ruleID)
		if err := os.Mkdir(ruleDir, 0775); err != nil {
			return fmt.Errorf("Failed to create last run cache entry %s: %s", ruleDir, err)
		}
	}

	return nil
}

func (cache *LocalCache) Plan() []string {
	return cache.plan
}

func (cache *LocalCache) ClearPlan() {
	cache.plan = cache.plan[:0]
}

func (cache *LocalCache) ScheduleRule(rule *Rule) {
	cache.plan = append(cache.plan, rule.Hash())
}
