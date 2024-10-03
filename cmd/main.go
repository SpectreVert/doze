package main

import (
	"fmt"
	"os"

	"github.com/spectrevert/doze"

	// Load modules below
	_ "github.com/spectrevert/doze/procedures/myproc" // this one imports two empty procedures-- with real names
)

func main() {
	if err := doze.Run(os.Args[1:]); err != nil {
		fmt.Println("doze:", err)
		os.Exit(1)
	}
}
