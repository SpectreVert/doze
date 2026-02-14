package dozecmd

import (
	"fmt"

	"github.com/spectrevert/doze"
	"github.com/spectrevert/doze/dozefile"
)

// The entrypoint of the Doze CLI.
func Main() int {
	// Parse the dozefile.
	var dfPath = "Dozefile.yml"
	graph, err := dozefile.ParseYAML(dfPath)
	if err != nil {
		fmt.Println(err)
		return 1
	}

	// Start the cache.
	opts := doze.CacheOptions{Path: ".doze"}
	cache, err := doze.NewLocalCache(opts)

	// @hardcoded the execution mode. Should think about the CLI.
	doze.Resolve(graph, cache, doze.TerseMode)

	// Exexute the graph and display some information.
	var executedRules, fetchedRules int
	if executedRules, fetchedRules, err = doze.Execute(graph, cache); err != nil {
		fmt.Printf("doze: %s\n", err)
	} else {
		if executedRules > 0 || fetchedRules > 0 {
			fmt.Printf("Done. Executed: %d Fetched: %d\n", executedRules, fetchedRules)
		} else {
			fmt.Printf("Nothing to do. Dozing off...\n")
		}
	}

	return 0
}
