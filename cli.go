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

// Outputs the list of procedures loaded into this doze binary.
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

// Updates the artifacts as per the build configuration.
//
// In the future, the build config will be specified in
// a build file, the Dozefile. I haven't settled yet on
// a syntax/language to use for the Dozefile so for now
// we're hardcoding structs :D
//
// Takes in the command-line arguments for the command inside the UpdateArtifactsCmd struct.
func (cmd UpdateArtifactsCmd) Run() error {
	fmt.Println("cmd.TargetArtifacts", cmd.TargetArtifacts)
	fmt.Println("cmd.TargetNamedRules", cmd.TargetNamedRules)

	df := NewDozefile()

	// a hardcoded build
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

	plan, err := df.MakePlan(cmd.TargetArtifacts)
	if err != nil {
		return err
	}

	// Debug output some more...
	for _, planned := range plan {
		fmt.Println(planned.procInfo.ID)
		fmt.Println(" inputs:")
		for _, inputTag := range planned.inputs {
			fmt.Println("  ", inputTag)
		}
		fmt.Println(" outputs:")
		for _, outputTag := range planned.outputs {
			fmt.Println("  ", outputTag)
		}
	}

	return nil
}
