package component

import (
	"os"
	"testing"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/goccy/go-yaml"
	"github.com/oscal-compass/oscal-sdk-go/extensions"
	"github.com/oscal-compass/oscal-sdk-go/validation"
	"github.com/ossf/gemara/layer2"
	"github.com/ossf/gemara/layer4"
	"github.com/stretchr/testify/require"
)

func TestDefinitionBuilder_Build(t *testing.T) {
	file, err := os.Open("./testdata/good-osps.yml")
	require.NoError(t, err)

	var parameters Parameters
	require.NoError(t, parameters.Load("./testdata/parameters.yml"))

	var catalog layer2.Catalog
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&catalog)
	require.NoError(t, err)

	plan := layer4.EvaluationPlan{
		Metadata: layer4.Metadata{
			Author: layer4.Author{
				Name: "myvalidator",
			},
		},
		Executors: []layer4.AssessmentExecutor{
			{
				Id:   "automated-executor",
				Name: "Automated Executor",
				Type: layer4.Automated,
			},
		},
		Plans: []layer4.AssessmentPlan{
			{
				Control: layer4.Mapping{
					EntryId: "OSPS-QA-07",
				},
				Assessments: []layer4.Assessment{
					{
						Requirement: layer4.Mapping{
							EntryId: "OSPS-QA-07.01",
						},
						Procedures: []layer4.AssessmentProcedure{
							{
								Id:          "my-check-id",
								Name:        "My Check",
								Description: "My Check",
								Executors: []layer4.ExecutorMapping{
									{
										Id: "automated-executor",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	builder := NewDefinitionBuilder("ComponentDefinition", "v0.1.0")
	componentDefinition := builder.AddTargetComponent("Example", "software", catalog, parameters).AddValidationComponent(plan).Build()
	require.Len(t, *componentDefinition.Components, 2)

	components := *componentDefinition.Components
	require.Len(t, *components[0].Props, 5)
	require.Len(t, *components[1].Props, 3)
	require.Equal(t, "Automated Executor", components[1].Title)
	require.Equal(t, "validation", components[1].Type)

	ci := *components[0].ControlImplementations
	require.Len(t, ci, 1)
	require.Equal(t, []oscalTypes.Property{{Name: extensions.FrameworkProp, Value: "800-161", Ns: extensions.TrestleNameSpace}}, *ci[0].Props)

	oscalModels := oscalTypes.OscalModels{
		ComponentDefinition: &componentDefinition,
	}

	validator := validation.NewSchemaValidator()
	err = validator.Validate(oscalModels)
	require.NoError(t, err)

	componentDefinition = builder.AddParameterModifiers("OSPS-B", []ParameterModifier{{
		TargetId: "main_branch_min_approvals",
		ModType:  "tighten",
		Value:    2,
	}}).Build()
	require.Len(t, *componentDefinition.Components, 2)
	ci = *components[0].ControlImplementations
	require.Len(t, ci, 1)
	require.Equal(t, []oscalTypes.SetParameter{{ParamId: "main_branch_min_approvals", Values: []string{"2"}}}, *ci[0].SetParameters)
}

func TestDefinitionBuilder_AddValidationComponent_MultipleExecutors(t *testing.T) {
	plan := layer4.EvaluationPlan{
		Metadata: layer4.Metadata{
			Author: layer4.Author{
				Name: "myvalidator",
			},
		},
		Executors: []layer4.AssessmentExecutor{
			{
				Id:   "executor-1",
				Name: "Executor One",
				Type: layer4.Automated,
			},
			{
				Id:   "executor-2",
				Name: "Executor Two",
				Type: layer4.Automated,
			},
		},
		Plans: []layer4.AssessmentPlan{
			{
				Control: layer4.Mapping{
					EntryId: "OSPS-QA-07",
				},
				Assessments: []layer4.Assessment{
					{
						Requirement: layer4.Mapping{
							EntryId: "OSPS-QA-07.01",
						},
						Procedures: []layer4.AssessmentProcedure{
							{
								Id:          "check-1",
								Name:        "Check One",
								Description: "First check",
								Executors: []layer4.ExecutorMapping{
									{Id: "executor-1"},
								},
							},
							{
								Id:          "check-2",
								Name:        "Check Two",
								Description: "Second check",
								Executors: []layer4.ExecutorMapping{
									{Id: "executor-2"},
								},
							},
						},
					},
				},
			},
		},
	}

	builder := NewDefinitionBuilder("ComponentDefinition", "v0.1.0")
	componentDefinition := builder.AddValidationComponent(plan).Build()
	require.Len(t, *componentDefinition.Components, 2)

	components := *componentDefinition.Components
	require.Equal(t, "Executor One", components[0].Title)
	require.Equal(t, "Executor Two", components[1].Title)
	require.Equal(t, "validation", components[0].Type)
	require.Equal(t, "validation", components[1].Type)
}
