package main

import (
	"github.com/spectrevert/doze"
)

func main() {
	graph := doze.NewGraph()

	graph.AddRule(
		[]string{"main.c", "parse.h"},
		[]string{"main.o"},
		"lang:c:object_file",
		"", "",
	)
	graph.AddRule(
		[]string{"parse.c", "parse.h"},
		[]string{"parse.o"},
		"lang:c:object_file",
		"", "",
	)
	graph.AddRule(
		[]string{"parse.y"},
		[]string{"parse.h", "parse.c"},
		"lang:c:yacc",
		"", "",
	)
	graph.AddRule(
		[]string{"parse.o", "main.o"},
		[]string{"exe"},
		"lang:c:executable",
		"", "",
	)

	graph.MarkArtifactsAsExisting()

	ruleHashes := graph.Resolve()
	graph.Execute(ruleHashes)
}
