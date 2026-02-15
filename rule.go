package doze

import (
	"crypto"
	_ "crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

// A Rule transforms input files into output files, using a Procedure.
// The checksum takes the input tags, output tags and procID into consideration
// and is used as a key to the Rule struct in the internal Doze context.
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
	vars Vars,
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

	// Interpolate variables into input and output tags, input and output dirs.
	if err := Interpolate(&inputs, &vars); err != nil {
		return err
	}
	if err := Interpolate(&outputs, &vars); err != nil {
		return err
	}
	if err := Interpolate(&inputDir, &vars); err != nil {
		return err
	}
	if err := Interpolate(&outputDir, &vars); err != nil {
		return err
	}

	// For passed input and output tags, fetch or create an Artifact and fill its Rule field.
	for _, tag := range inputs {
		tag = filepath.Join(inputDir, tag)
		_, ok := graph.artifacts[tag]
		if !ok {
			graph.artifacts[tag] = &Artifact{
				tag: ArtifactTag(tag),
			}
		}
		rule.Inputs = append(rule.Inputs, ArtifactTag(tag))
		graph.artifacts[tag].consumers = append(graph.artifacts[tag].consumers, rule)
	}

	for _, tag := range outputs {
		tag = filepath.Join(outputDir, tag)
		artifact, ok := graph.artifacts[tag]
		if !ok {
			graph.artifacts[tag] = &Artifact{
				tag:     ArtifactTag(tag),
				creator: rule,
			}
		} else if artifact.creator != nil {
			// @todo format the error message better e.g (<tag-name>)[artifact-checksum]
			return fmt.Errorf("artifact (%v) was registered as an output twice", tag)
		} else {
			artifact.creator = rule
		}
		rule.Outputs = append(rule.Outputs, ArtifactTag(tag))
	}

	checksum := rule.Checksum()
	_, ok := graph.rules[checksum]
	if ok {
		return fmt.Errorf("rule [%v] already exists with the same data", rule.Checksum())
	}

	graph.rules[checksum] = rule
	return nil
}

// The checksum of a Rule is the deterministic checksum built from a Rule's input tags, output tags and procID.
// The checksum is cached directly in the Rule structure.
// Caching of the Rule checksum is possible because its fields are constants.
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

	rule.checksum = hex.EncodeToString(checksum.Sum(nil))
	return rule.checksum
}

// The source checksum of a Rule is the deterministic checksum of the content of all the Rule's inputs.
// It is used to decide if a Rule is outdated. It cannot be cached.
func (rule *Rule) SourceChecksum() string {
	sourceChecksum := crypto.MD5.New()
	for _, tag := range rule.Inputs {
		content, err := os.ReadFile(string(tag))
		if err != nil {
			panic(fmt.Sprintf("could not open artifact (%s) for reading: %s", tag, err))
		}
		sourceChecksum.Write(content)
	}
	return hex.EncodeToString(sourceChecksum.Sum(nil))
}

func (rule *Rule) Execute() error {
	procInfo, err := GetProcedure(rule.procID)
	if err != nil {
		return err
	}

	fmt.Printf("  %s: %v -> %v\n", rule.procID, rule.Inputs, rule.Outputs)
	proc := procInfo.New()
	if err = proc.Execute(rule); err != nil {
		return err
	}

	return nil
}
