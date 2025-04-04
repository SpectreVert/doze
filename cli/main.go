package main

import (
	"github.com/spectrevert/doze"
)

func main() {
	graph := doze.NewGraph()

	graph.AddRule([]string{"parse.y"}, []string{"parse.h", "parse.c"}, "", "")
	graph.AddRule([]string{"parse.c", "parse.h"}, []string{"parse.o"}, "", "")
	graph.AddRule([]string{"main.c", "parse.h"}, []string{"main.o"}, "", "")
	graph.AddRule([]string{"parse.o", "main.o"}, []string{"exe"}, "", "")

	graph.MarkArtifactsAsExisting()

	ruleHashes := graph.Resolve()
	graph.Execute(ruleHashes)
}
