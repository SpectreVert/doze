package main

import (
	"os"

	dozecmd "github.com/spectrevert/doze/cmd"

	// Plug-in your own Doze procedures here.
	_ "github.com/spectrevert/doze/procedures/standard"
)

func main() {
	os.Exit(dozecmd.Main())
}
