package cmd

import (
	"fmt"
	"os"

	"github.com/spectrevert/doze"
)

func Main() int {
	if err := doze.Run(os.Args[1:]); err != nil {
		fmt.Println("doze:", err)
		return 1
	}
	return 0
}
