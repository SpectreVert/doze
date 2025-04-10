package dozefile

import (
	"fmt"
	"io/ioutil"

	"github.com/spectrevert/doze"

	"gopkg.in/yaml.v2"
)

type Rule struct {
	Procedure string   `yaml:"do"`
	Inputs    []string `yaml:",flow"`
	Outputs   []string `yaml:",flow"`
}

type Schema struct {
	Rules []Rule `yaml:",omitempty"`
}

// A Dozefile is a yaml file expressing a build graph.
// Return a Graph constructed from the dozefile, or an error.
func ParseDozefileYAML(dozefilePath string) (*doze.Graph, error) {
	content, err := ioutil.ReadFile(dozefilePath)
	if err != nil {
		return nil, fmt.Errorf("parseDozefileYAML: %v", err)
	}

	var dozefileYAML Schema
	if err := yaml.Unmarshal(content, &dozefileYAML); err != nil {
		return nil, fmt.Errorf("parseDozefileYAML: %v", err)
	}

	// Transform the Schema into a Graph.
	g := doze.NewGraph()

	if dozefileYAML.Rules == nil {
		return nil, fmt.Errorf("parseDozefileYAML: provide at least one element in 'rules'")
	}
	for _, rule := range dozefileYAML.Rules {
		if _, err := doze.GetProcedure(doze.ProcedureID(rule.Procedure)); err != nil {
			return nil, fmt.Errorf("parseDozefileYAML: %v", err)
		}
		if err := g.AddRule(rule.Inputs, rule.Outputs, doze.ProcedureID(rule.Procedure), "", ""); err != nil {
			return nil, fmt.Errorf("parseDozefileYAML: %v", err)
		}
	}

	return g, nil
}
