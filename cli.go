package doze

import (
	"fmt"
	"runtime"

	"github.com/alexflint/go-arg"

	"github.com/spectrevert/doze/internal"
)

var LinkerVersion = "unknown"

type ListProcsCmd struct{}
type UpdateArtifactsCmd struct {
	TargetArtifacts  []string
	TargetNamedRules []string
}

type options struct {
	ArtifactsTag  []string `arg:"-a,--artifact,separate" help:"artifacts to bring up to date"`
	NamedRulesTag []string `arg:"-r,--rule,separate" help:"named rules to execute"`

	// ** Commands
	// * Modules
	ListProcs *ListProcsCmd `arg:"subcommand:ls-procs" help:"list the procedures loaded into this doze binary"`
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
			TargetArtifacts:  opts.ArtifactsTag,
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
	fmt.Println("cmd.TargetArtifacts", cmd.TargetArtifacts)
	fmt.Println("cmd.TargetNamedRules", cmd.TargetNamedRules)

	df := NewDozefile()
	ins0 := []string{"parse.o", "main.o"}
	outs0 := []string{"ykz"}

	ins1 := []string{"parse.h", "parse.c"}
	outs1 := []string{"parse.o"}

	ins2 := []string{"parse.h", "main.c"}
	outs2 := []string{"main.o"}

	ins3 := []string{"parse.y"}
	outs3 := []string{"parse.c", "parse.h"}

	if err := df.createRule(ins1, outs1, "somenamespace:bigzbi"); err != nil {
		return err
	}
	if err := df.createRule(ins0, outs0, "somenamespace:bigzbi"); err != nil {
		return err
	}
	if err := df.createRule(ins2, outs2, "somenamespace:bigzben"); err != nil {
		return err
	}
	if err := df.createRule(ins3, outs3, "somenamespace:bigzben"); err != nil {
		return err
	}

	plan, err := df.Rebuild(cmd.TargetArtifacts)
	if err != nil {
		return err
	}

	for _, planned := range plan {
		fmt.Println(planned.procInfo.ID)
		for _, inputTag := range planned.inputs {
			fmt.Println("  ", inputTag)
		}
		for _, outputTag := range planned.outputs {
			fmt.Println("  ", outputTag)
		}
	}

	return nil
}
