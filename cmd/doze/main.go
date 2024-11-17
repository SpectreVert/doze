package main

import (
	"os"

	cmd "github.com/spectrevert/doze/cmd"
	// plug in Doze modules under this comment
)

func main() {
	os.Exit(cmd.Main())
}
