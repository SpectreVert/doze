package doze

import "strings"

type Artifact struct {
	creator   *Rule
	consumers []*Rule

	tag ArtifactTag
}

type ArtifactTag string

func CompareArtifactTags(ArtifactTag first, ArtifactTag second) int {
	return strings.Compare(first, second)
}
