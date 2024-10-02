package doze

import (
	"fmt"
	"runtime"

	"github.com/alexflint/go-arg"

	"github.com/spectrevert/doze/internal"
)

var LinkerVersion = "unknown"

type ListProcsCmd struct {}
type UpdateArtifactsCmd struct {
	TargetArtifacts  []string
	TargetNamedRules []string
}

type options struct {
	ArtifactsTag  []string `arg:"-a,--artifact" help:"artifacts to bring up to date"`
	NamedRulesTag []string `arg:"-r,--rule" help:"named rules to execute"`

	// ** Commands
	// * Modules
	ListProcs       *ListProcsCmd        `arg:"subcommand:ls-procs" help:"list the procedures loaded into this doze binary"`
}

func (options) Description() string {
	return "An opinionated, predictable and lazy build system\n"
}

func (options) Epilogue() string {
	return "Made with " + runtime.Version() + ". For more information visit https://github.com/spectrevert/doze"
}

func (options) Version() string {
	return internal.Version(LinkerVersion)
}

func Run(cmdLine []string) error {
	var opts options
	cli, err := arg.NewParser(arg.Config{}, &opts)
	if err != nil {
		return fmt.Errorf("init cli parsing: %s", err)
	}
	cli.MustParse(cmdLine)

	switch {
	case opts.ListProcs != nil:
		return opts.ListProcs.Run()
	default:
		// this is the default build path
		return UpdateArtifactsCmd{
			TargetArtifacts: opts.ArtifactsTag,
			TargetNamedRules: opts.NamedRulesTag,
		}.Run()
	}
}

func (cmd ListProcsCmd) Run() error {
	procs := GetProcedures("")
	if procs == nil {
		return fmt.Errorf("no procedures available")
	}

	for _, proc := range procs {
		fmt.Println(proc.ID)
	}

	return nil
}

func (cmd UpdateArtifactsCmd) Run() error {
	fmt.Println(cmd.TargetArtifacts)
	fmt.Println(cmd.TargetNamedRules)

	df := NewDozefile()
	ins1 := []string{"parse.o", "main.o"}
	outs1 := []string{"ykz"}

	df.createRule(ins1, outs1, "somenamespace:bigzbi")
	if err := df.createRule(ins1, outs1, "makismusam"); err != nil {
		return err
	}

	return nil
}