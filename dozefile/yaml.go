package dozefile

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"

	"github.com/spectrevert/doze"
)

// An implementation of Dozefile targetting YAML configuration files.

type schema struct {
	Vars  doze.Vars `yaml:",omitempty"`
	Rules []struct {
		Inputs  []string `yaml:",flow"`
		Outputs []string `yaml:",flow"`
		ProcID  string   `yaml:"do"`
		InDir   string   `yaml:"in"`
		OutDir  string   `yaml:"out"`
	} `yaml:",omitempty"`
}

func ParseYAML(path string) (*doze.Graph, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("(%v): %v", path, err)
	}

	var dozefile schema
	if err = yaml.Unmarshal(content, &dozefile); err != nil {
		return nil, fmt.Errorf("(%v): %v", path, err)
	}
	if dozefile.Rules == nil {
		return nil, fmt.Errorf("(%v): must contain at least one rule", path)
	}

	g := doze.NewGraph()
	for _, rule := range dozefile.Rules {
		if _, err := doze.GetProcedure(doze.ProcedureID(rule.ProcID)); err != nil {
			return nil, fmt.Errorf("(%v): %v", path, err)
		}
		if err = doze.NewRule(g, rule.Inputs, rule.Outputs, rule.InDir, rule.OutDir, doze.ProcedureID(rule.ProcID), dozefile.Vars); err != nil {
			return nil, fmt.Errorf("(%v): %v", path, err)
		}
	}

	return g, nil
}
