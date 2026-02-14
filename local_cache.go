package doze

import (
	"fmt"
	"os"
	"path/filepath"
)

type LocalCache struct {
	CacheOptions

	recordedRules []string
	plan          []string
}

// The LocalCache uses a directory structure to store output artifacts.
//
// .doze/						<-	Default name of the cache directory.
// (PRIMORDIAL CACHE)
//		last_run/						<- Contains the primordial rules of the last run.
//			aabbccddee_abcdeffedcba		<- Rule checksum + _ + checksum of its inputs.
//			ffee33bb22_74d88f344eff
//
// (ARTIFACT CACHE)						<- Contains the artifacts of the previous runs.
//		df1833d6159.../					<- Rule checksum.
//			74d887f344b99.../			<- Checksum of its inputs.
//				main.o					<- All outputs created by the Rule, with these exact inputs.
//			070789df76f25.../			<- Checksum of different inputs.
//				main.o					<- All outputs created by the Rule, with these exact inputs.

func NewLocalCache(options CacheOptions) (*LocalCache, error) {
	if options.Path == "" {
		return nil, fmt.Errorf("CacheOptions must have Path set to an accessible location")
	}
	if err := os.MkdirAll(options.Path, 0o775); err != nil {
		return nil, fmt.Errorf("failed to create root cache directory %s: %s", options.Path, err)
	}

	return &LocalCache{
		CacheOptions: options,
	}, nil
}

func (cache *LocalCache) HasArtifacts(rule *Rule) bool {
	ruleDir := filepath.Join(cache.Path, rule.Checksum())
	sourceChecksum := rule.SourceChecksum()
	outputsDir := filepath.Join(ruleDir, sourceChecksum)

	if _, err := os.Stat(outputsDir); err != nil {
		return false
	}

	return true
}

// @robustness This function "atomically" stores the newly created artifacts in the Cache, but it doesn't cleanup its temporary files
// in case of failure.
func (cache *LocalCache) StoreArtifacts(rule *Rule) error {
	ruleDir := filepath.Join(cache.Path, rule.Checksum())

	// Check if the Rule cache directory already exists, create it if it doesn't.
	if _, err := os.Stat(ruleDir); err != nil {
		if err = os.Mkdir(ruleDir, 0o775); err != nil {
			return fmt.Errorf("failed to create rule cache directory %s: %s", ruleDir, err)
		}
	}

	sourceChecksum := rule.SourceChecksum()
	outputsDir := filepath.Join(ruleDir, sourceChecksum)

	// Check that the Rule output directory does *not* exist yet.
	if _, err := os.Stat(outputsDir); err == nil {
		return fmt.Errorf("rule output directory already exists: %s", outputsDir)
	}

	// Create the temporary outputs directory in the cache, using the source checksum of the Rule.
	tmpOutputsDir := outputsDir + temporaryDirMarker
	if err := os.Mkdir(tmpOutputsDir, 0o775); err != nil {
		return fmt.Errorf("failed to create temporary rule output cache directory %s: %s", tmpOutputsDir, err)
	}

	// Create hardlinks to all of the outputs files.
	for _, tag := range rule.Outputs {
		linkPath := filepath.Join(tmpOutputsDir, string(tag))
		CreateBaseDir(linkPath)

		if err := CopyFile(string(tag), linkPath); err != nil {
			// @todo cleanup temporary directory.
			return fmt.Errorf("failed to copy output %s to the cache %s: %s", tag, linkPath, err)
		}
	}

	// Move the temporary directory to the real location.
	if err := os.Rename(tmpOutputsDir, outputsDir); err != nil {
		return fmt.Errorf("failed to rename temporary rule output cache directory %s to %s: %s", tmpOutputsDir, outputsDir, err)
	}

	return nil
}

func (cache *LocalCache) ProduceArtifacts(rule *Rule) error {
	ruleDir := filepath.Join(cache.Path, rule.Checksum())
	sourceChecksum := rule.SourceChecksum()
	outputsDir := filepath.Join(ruleDir, sourceChecksum)

	// Check that the Rule cache directory exists.
	if _, err := os.Stat(outputsDir); err != nil {
		return fmt.Errorf("failed to access rule output directory %s: %s", outputsDir, err)
	}

	// Create hardlinks from each cached output to their respective locations in the working tree directory.
	for _, tag := range rule.Outputs {
		cachedPath := filepath.Join(outputsDir, string(tag))

		// Check if the output already exists in the working tree directory.
		if info, err := os.Stat(string(tag)); err == nil {
			if cachedInfo, err := os.Stat(cachedPath); err == nil {
				if os.SameFile(info, cachedInfo) {
					fmt.Printf("  skipped: (%s)\n", tag)
					continue
				}
			}
			// The output already exists in the working tree directory, but is not up-to-date. We clean it.
			if err = os.Remove(string(tag)); err != nil {
				return fmt.Errorf("failed to remove outdated output %s: %s", tag, err)
			}
			fmt.Printf("  fetched: (%s)\n", tag)
		}
		if err := os.Link(cachedPath, string(tag)); err != nil {
			return fmt.Errorf("failed to create hardlink from %s to %s: %s", cachedPath, tag, err)
		}
	}

	return nil
}

// PrimordialCache

func (cache *LocalCache) RuleInLastRun(rule *Rule) bool {
	lastRunDir := filepath.Join(cache.Path, "last_run")
	if _, err := os.Stat(lastRunDir); err != nil {
		return false
	}

	ruleDir := filepath.Join(lastRunDir, rule.Checksum()+"_"+rule.SourceChecksum())
	if _, err := os.Stat(ruleDir); err == nil {
		return true
	}

	return false
}

func (cache *LocalCache) RecordRule(rule *Rule) {
	cache.recordedRules = append(cache.recordedRules, rule.Checksum()+"_"+rule.SourceChecksum())
}

func (cache *LocalCache) FlushRecords() error {
	// Create the `last_run` directory if it does not exist yet.
	lastRunDir := filepath.Join(cache.Path, "last_run")
	if _, err := os.Stat(lastRunDir); err != nil {
		if err = os.Mkdir(lastRunDir, 0o775); err != nil {
			return fmt.Errorf("failed to create `last_run` cache directory %s: %s", lastRunDir, err)
		}
	}

	// Delete all the previously recorded rules.
	lastRunDirEntries, err := os.ReadDir(lastRunDir)
	if err != nil {
		return fmt.Errorf("failed to read `last_run` cache directory %s: %s", lastRunDir, err)
	}
	for _, entry := range lastRunDirEntries {
		path := filepath.Join(lastRunDir, entry.Name())
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("failed to delete `last_run` cache entry %s: %s", path, err)
		}
	}

	// Write the new recorded rules
	for _, ruleID := range cache.recordedRules {
		ruleDir := filepath.Join(lastRunDir, ruleID)
		if err = os.Mkdir(ruleDir, 0o775); err != nil {
			return fmt.Errorf("failed to create `last_run` cache entry %s: %s", ruleDir, err)
		}
	}

	return nil
}

// PlanCache
func (cache *LocalCache) Plan() []string {
	return cache.plan
}

func (cache *LocalCache) ClearPlan() {
	cache.plan = cache.plan[:0]
}

func (cache *LocalCache) ScheduleRule(rule *Rule) {
	cache.plan = append(cache.plan, rule.Checksum())
}

const temporaryDirMarker = "="
