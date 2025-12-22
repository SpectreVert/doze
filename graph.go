package doze

type Graph struct {
	rules     map[string]*Rule
	artifacts map[string]*Artifact

	checksum string
}

type ResolveMode int

func NewGraph() *Graph {
	return &Graph{
		rules:     make(map[string]*Rule),
		artifacts: make(map[string]*Artifact),
	}
}

func Execute(graph *Graph, cache *Cache) (int, int, error) {
}

func Resolve(graph *Graph, cache *Cache, mode ResolveMode) (int, error) {
}

const (
	Terse ResolveMode = iota // Terse mode checks if the primordial inputs of the Graph have changed and schedules only rules that descend from modified primordial inputs.
	Full                     // Full mode shedules all rules that Terse mode would schedule and in addition schedules any rule that is missing an output in the build directory.
	//							Full mode can be useful when outputs have potentially been removed from the build directory.
)
