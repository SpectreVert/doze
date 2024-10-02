package main

import (
	"fmt"
	"os"

	"github.com/spectrevert/doze/pkg/doze"
)

func main() {
	if err := doze.Run(os.Args[1:]); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
