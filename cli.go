package doze

import (
	"fmt"
	"runtime"

	"github.com/alexflint/go-arg"
)

// set by the linker (see Taskfile)
var LinkerVersion = "unknown"

type options struct {
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

	// ephemeral graph
	graph := NewGraph()

	/* 	// a hardcoded build
	   	ins0 := []ArtifactTag{"parse.o", "main.o"}
	   	outs0 := []ArtifactTag{"ykz"}

	   	ins1 := []ArtifactTag{"parse.h", "parse.c"}
	   	outs1 := []ArtifactTag{"parse.o"}

	   	ins2 := []ArtifactTag{"main.c"}
	   	outs2 := []ArtifactTag{"main.o"}

	   	if err := graph.createRule(ins1, outs1, "lang.c.objectFile"); err != nil {
	   		return err
	   	}

	   	if err := graph.createRule(ins2, outs2, "lang.c.objectFile"); err != nil {
	   		return err
	   	}

	   	if err := graph.createRule(ins0, outs0, "lang.c.linker"); err != nil {
	   		return err
	   	}

	   	// harcoded target Artifacts
	   	targets := []ArtifactTag{"main.o"} */

	// a hardcoded build

	i0 := []ArtifactTag{"primordial_1", "primordial_2"}
	o0 := []ArtifactTag{"inter_1", "inter_2"}

	i1 := []ArtifactTag{"primordial_3"}
	o1 := []ArtifactTag{"inter_3"}

	i2 := []ArtifactTag{"inter_1", "inter_2", "inter_3"}
	o2 := []ArtifactTag{"inter_4", "inter_5"}

	i3 := []ArtifactTag{"inter_3", "inter_4"}
	o3 := []ArtifactTag{"final_1"}

	i4 := []ArtifactTag{"inter_4", "inter_5"}
	o4 := []ArtifactTag{"final_2"}

	if err := graph.createRule(i0, o0, "lang.c.objectFile"); err != nil {
		return err
	}

	if err := graph.createRule(i1, o1, "lang.c.objectFile"); err != nil {
		return err
	}

	if err := graph.createRule(i2, o2, "lang.c.objectFile"); err != nil {
		return err
	}

	if err := graph.createRule(i3, o3, "lang.c.objectFile"); err != nil {
		return err
	}

	if err := graph.createRule(i4, o4, "lang.c.objectFile"); err != nil {
		return err
	}

	// using SimpleRebuilder to create a build plan
	dr := new(DeepRebuilder)

	// build targets
	targets := []ArtifactTag{"final_1", "final_2"}

	plan, err := dr.rebuild(targets, 0, graph)
	if err != nil {
		return err
	}

	// some debug output
	for _, planned := range plan {
		fmt.Println("proc:", planned.procInfo.ID)
		fmt.Println("inputs:")
		for _, input := range planned.inputs {
			fmt.Println(" ", input)
		}
		fmt.Println("outputs:")
		for _, output := range planned.outputs {
			fmt.Println(" ", output)
		}
	}

	return err
}
