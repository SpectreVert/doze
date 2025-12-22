package doze

import (
	"crypto"
	_ "crypto/md5"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"slices"
)

type Rule struct {
	Inputs, Outputs []ArtifactTag
	procID          ProcedureID

	checksum string
}

func NewRule(
	graph *Graph,
	inputs, outputs []string,
	inputDir, outputDir string,
	procID ProcedureID,
) error {
	if len(inputs) == 0 {
		return fmt.Errorf("no inputs provided")
	}
	if len(outputs) == 0 {
		return fmt.Errorf("no outputs provided")
	}
	if _, err := GetProcedure(procID); err != nil {
		return err
	}

	rule := &Rule{
		procID: procID,
	}

	// For passed input and output tags, fetch or create an Artifact and fill its Rule field.
	for _, tag := range inputs {
		tag = filepath.Join(inputDir, tag)
		_, ok := graph.artifacts[tag]
		if !ok {
			graph.artifacts[tag] = &Artifact{
				tag: tag,
			}
		}
		rule.Inputs = append(rule.Inputs, tag)
		graph.artifacts[tag].consumers = append(graph.artifacts[tag].consumers, rule)
	}

	for _, tag := range outputs {
		tag = filepath.Join(outputDir, tag)
		artifact, ok := graph.artifacts[tag]
		if !ok {
			graph.artifacts[tag] = &Artifact{
				tag:     tag,
				creator: rule,
			}
		} else if artifact.creator != nil {
			return fmt.Errorf("artifact (%v) was registered as an output twice", tag) // TODO: format the error message better e.g (<tag-name>)[artifact-checksum]
		} else {
			artifact.creator = rule
		}
		rule.Outputs = append(rule.Outputs, tag)
	}

	checksum := rule.Checksum()
	_, ok := graph.rules[checksum]
	if ok {
		return fmt.Errorf("rule [%v] already exists with the same data", rule.Checksum())
	}

	graph.rules[checksum] = rule
	return nil
}

func (rule *Rule) Checksum() string {
	if rule.checksum != "" {
		return rule.checksum
	}

	checksum := crypto.MD5.New()
	slices.SortFunc(rule.Inputs, CompareArtifactTags)
	for _, tag := range rule.Inputs {
		checksum.Write([]byte(tag))
	}
	slices.SortFunc(rule.Outputs, CompareArtifactTags)
	for _, tag := range rule.Outputs {
		checksum.Write([]byte(tag))
	}
	checksum.Write([]byte(rule.procID))

	rule.checksum = hex.EncodeToString(hash.Sum(nil))
	return rule.checksum
}

func (rule *Rule) Execute() error {
	if procInfo, err := GetProcedure(rule.procID); err != nil {
		return err
	}

	proc := procInfo.New()
	if err = proc.Execute(rule); err != nil {
		return err
	}

	return nil
}
