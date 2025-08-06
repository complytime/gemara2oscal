package controls

import (
	"os"
	"testing"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/goccy/go-yaml"
	"github.com/oscal-compass/oscal-sdk-go/validation"
	"github.com/ossf/gemara/layer1"
	"github.com/stretchr/testify/require"
)

func TestToOSCALCatalog(t *testing.T) {
	file, err := os.Open("./testdata/800-161.yml")
	require.NoError(t, err)

	var guidance layer1.GuidanceDocument
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&guidance)
	require.NoError(t, err)

	catalog, err := ToOSCALCatalog(guidance)
	require.NoError(t, err)

	oscalModels := oscalTypes.OscalModels{
		Catalog: &catalog,
	}

	validator := validation.NewSchemaValidator()
	err = validator.Validate(oscalModels)
	require.NoError(t, err)
}

func TestToOSCALProfile(t *testing.T) {
	file, err := os.Open("./testdata/800-161.yml")
	require.NoError(t, err)

	var guidance layer1.GuidanceDocument
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&guidance)
	require.NoError(t, err)

	// Add some shared guidelines
	mapping := layer1.MappingReference{
		Id:          "EXP",
		Description: "Example",
		Version:     "0.1.0",
		Url:         "https://example.com",
	}

	sharedGuidelines := layer1.Mapping{
		ReferenceId: "EXP",
		Identifiers: []string{
			"EX-1",
			"EX-1(2)",
			"EX-2",
		},
	}

	guidance.Metadata.MappingReferences = append(guidance.Metadata.MappingReferences, mapping)
	guidance.SharedGuidelines = append(guidance.SharedGuidelines, sharedGuidelines)

	profile, err := ToOSCALProfile(guidance)
	require.NoError(t, err)

	oscalModels := oscalTypes.OscalModels{
		Profile: &profile,
	}

	validator := validation.NewSchemaValidator()
	err = validator.Validate(oscalModels)
	require.NoError(t, err)

	wantImports := []oscalTypes.Import{
		{
			Href: "https://example.com",
			IncludeControls: &[]oscalTypes.SelectControlById{
				{
					WithIds: &[]string{
						"ex-1",
						"ex-1.2",
						"ex-2",
					},
				},
			},
		},
	}
	require.Equal(t, wantImports, profile.Imports)
}
