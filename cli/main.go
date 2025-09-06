package main

import (
	"fmt"
	"os"

	"github.com/spectrevert/doze/dozefile"

	// Plug in Doze procedures under this comment.
	_ "github.com/spectrevert/doze/procedures/lang_c"
)

func main() {
	var dozefilePath = "Dozefile.yaml"

	graph, err := dozefile.ParseDozefileYAML(dozefilePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var plan = graph.Resolve()
	graph.Execute(plan)
}
