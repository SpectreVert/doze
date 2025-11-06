package doze

import (
	"crypto"
	"encoding/hex"
	"maps"
	"slices"
)

// The Hash function of a Graph. Must be deterministic.
// Takes into account all its Rules.
func (graph *Graph) Hash() string {
	if graph.cachedHash != "" {
		return graph.cachedHash
	}

	hash := crypto.MD5.New()

	// Iterate over the hashes of Rules in the Graph, sorted by keys.
	for _, k := range slices.Sorted(maps.Keys(graph.rules)) {
		hash.Write([]byte(k))
	}

	graph.cachedHash = hex.EncodeToString(hash.Sum(nil))
	return graph.cachedHash
}

// The Hash function of a Rule. Obviously, must be deterministic.
// Takes into account the input and output ArtifactTags, and the ProcedureID.
// @unsure: Do we need to normalize the characters in the strings here?
func (rule *Rule) Hash() string {
	if rule.cachedHash != "" {
		return rule.cachedHash
	}

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

	rule.cachedHash = hex.EncodeToString(hash.Sum(nil))
	return rule.cachedHash
}
