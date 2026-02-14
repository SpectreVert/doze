package doze

import (
	"os"
	"strings"
)

type Artifact struct {
	creator   *Rule
	consumers []*Rule

	tag ArtifactTag
}

type ArtifactTag string

func (tag ArtifactTag) String() string {
	return string(tag)
}

func (tag ArtifactTag) Exists() bool {
	_, err := os.Stat(string(tag))
	return (err == nil)
}

func CompareArtifactTags(first, second ArtifactTag) int {
	return strings.Compare(string(first), string(second))
}
