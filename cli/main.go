package main

import (
	"fmt"
	"os"

	"github.com/spectrevert/doze"
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

	cache := doze.NewLocalCache()

	graph.Resolve(cache)
	if _, _, err := graph.Execute(cache); err != nil {
		fmt.Println("execute: %s", err)
	}
}
