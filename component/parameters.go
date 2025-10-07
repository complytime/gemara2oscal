package component

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/ossf/gemara/layer3"
)

// Parameter define a "knob" that can be tuned on an Assessment Requirement
type Parameter struct {
	Id string `json:"id" yaml:"id"`

	Description string `yaml:"description,omitempty"`

	Default any `yaml:"default,omitempty"`
}

// Parameters maps a slice of parameters to an assessment requirement id.
type Parameters map[string][]Parameter

// Load populates parameters from a file path
func (p *Parameters) Load(filePath string) error {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	return yaml.Unmarshal(file, p)
}

// ParameterModifier defines a modification to the default parameter value
type ParameterModifier struct {
	TargetId string `yaml:"target-id"`

	ModType layer3.ModType `yaml:"modification-type"`

	ModificationRationale string `yaml:"modification-rationale"`

	Description string `yaml:"description,omitempty"`

	Value any `yaml:"value"`
}
