package main

import (
	"fmt"
	"os"

	"github.com/spectrevert/doze"

	// plug in Doze modules under this comment
	_ "github.com/spectrevert/doze/procedures/myTestProcedures"
)

func main() {
	if err := doze.Run(os.Args[1:]); err != nil {
		fmt.Println("doze:", err)
		os.Exit(1)
	}
}
