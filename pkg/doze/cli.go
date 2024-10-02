package doze

import (
	"fmt"

	"github.com/alexflint/go-arg"

	"github.com/spectrevert/doze/internal"
)

type Opts struct {
	ArtifactsTag  []string `arg:"-a,--artifact" help:"artifacts to bring up to date"`
	NamedRulesTag []string `arg:"-r,--rule" help:"named rules to execute"`
}

func (Opts) Version() string {
	return internal.Version(LinkerVersion)
}

func Run(cmdLine []string) error {
	var opts Opts
	cli, err := arg.NewParser(arg.Config{}, &opts)
	if err != nil {
		return fmt.Errorf("init cli parsing: %s", err)
	}
	cli.MustParse(cmdLine)

	return nil
}
