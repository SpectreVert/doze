package doze

import (
	"fmt"
	"runtime"

	"github.com/alexflint/go-arg"
)

// set by the linker (see Taskfile)
var LinkerVersion = "unknown"

type options struct {
	ID int
}

func (options) Description() string {
	return "A simple, predictable build system"
}

func (options) Epilogue() string {
	return "Made with " + runtime.Version() + ". For more information visit https://github.com/spectrevert/doze"
}

func (options) Version() string {
	return LinkerVersion
}

func Run(args []string) error {
	var opts options
	argParser, err := arg.NewParser(arg.Config{}, &opts)
	if err != nil {
		return fmt.Errorf("init cli parsing: %s", err)
	}
	argParser.MustParse(args)

	return err
}
