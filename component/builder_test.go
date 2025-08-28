package component

import (
	"os"
	"testing"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/goccy/go-yaml"
	"github.com/oscal-compass/oscal-sdk-go/extensions"
	"github.com/oscal-compass/oscal-sdk-go/validation"
	"github.com/ossf/gemara/layer2"
	"github.com/ossf/gemara/layer3"
	"github.com/ossf/gemara/layer4"
	"github.com/stretchr/testify/require"
)

func TestDefinitionBuilder_Build(t *testing.T) {
	file, err := os.Open("./testdata/good-osps.yml")
	require.NoError(t, err)

	var catalog layer2.Catalog
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&catalog)
	require.NoError(t, err)

	eval := layer4.AssessmentPlan{
		ControlId: "OSPS-QA-07",
		Assessments: []layer4.Assessment{
			{
				RequirementId: "OSPS-QA-07.01",
				Procedures: []layer4.AssessmentProcedure{
					{
						Id:          "my-check-id",
						Name:        "My Check",
						Description: "My Check",
					},
				},
			},
		},
	}

	builder := NewDefinitionBuilder("ComponentDefinition", "v0.1.0")
	componentDefinition := builder.AddTargetComponent("Example", "software", catalog).AddValidationComponent("myvalidator", []layer4.AssessmentPlan{eval}).Build()
	require.Len(t, *componentDefinition.Components, 2)

	components := *componentDefinition.Components
	require.Len(t, *components[0].Props, 5)
	require.Len(t, *components[1].Props, 3)

	ci := *components[0].ControlImplementations
	require.Len(t, ci, 1)
	require.Equal(t, []oscalTypes.Property{{Name: extensions.FrameworkProp, Value: "800-161", Ns: extensions.TrestleNameSpace}}, *ci[0].Props)
	implementReq := ci[0].ImplementedRequirements
	require.Len(t, implementReq, 5)

	oscalModels := oscalTypes.OscalModels{
		ComponentDefinition: &componentDefinition,
	}

	validator := validation.NewSchemaValidator()
	err = validator.Validate(oscalModels)
	require.NoError(t, err)

	componentDefinition = builder.AddParameterModifiers("OSPS-B", []layer3.ParameterModifier{{
		TargetId: "main_branch_min_approvals",
		ModType:  "tighten",
		Value:    2,
	}}).Build()
	require.Len(t, *componentDefinition.Components, 2)
	ci = *components[0].ControlImplementations
	require.Len(t, ci, 1)
	require.Equal(t, []oscalTypes.SetParameter{{ParamId: "main_branch_min_approvals", Values: []string{"2"}}}, *ci[0].SetParameters)
}
