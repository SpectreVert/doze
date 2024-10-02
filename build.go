package doze

// A Dozefile describes how a build should be performed.
type Dozefile struct {
	RuntimeRules     []Rule
	RuntimeArtifacts map[string]Artifact
	NamedRules       map[string][]Rule
}

type Rule struct {
	inputs  []Artifact
	outputs []Artifact
	proc    *Procedure

	isScheduled bool
}

type Artifact struct {
	tag  string
	path string

	isScheduled bool
	isFinal     bool

	creatorRule *Rule
}
