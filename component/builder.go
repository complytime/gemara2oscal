package component

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/defenseunicorns/go-oscal/src/pkg/uuid"
	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/oscal-compass/oscal-sdk-go/extensions"
	"github.com/oscal-compass/oscal-sdk-go/models"
	"github.com/ossf/gemara/layer2"
	"github.com/ossf/gemara/layer3"
	"github.com/ossf/gemara/layer4"

	"github.com/complytime/gemara2oscal/internal/utils"
)

// DefinitionBuilder constructs an OSCAL Component Definition from Gemara
// inputs.
type DefinitionBuilder struct {
	title               string
	version             string
	targetComponents    map[string]oscalTypes.DefinedComponent
	validationComponent []oscalTypes.DefinedComponent
}

func NewDefinitionBuilder(title, version string) *DefinitionBuilder {
	return &DefinitionBuilder{
		title:            title,
		version:          version,
		targetComponents: make(map[string]oscalTypes.DefinedComponent),
	}
}

func (c *DefinitionBuilder) AddTargetComponent(targetComponent, componentType string, catalog layer2.Catalog) *DefinitionBuilder {
	mappingSet := make(map[string]oscalTypes.ControlImplementationSet)
	for _, mappingRef := range catalog.Metadata.MappingReferences {
		mappingSet[mappingRef.Id] = oscalTypes.ControlImplementationSet{
			UUID:        uuid.NewUUID(),
			Description: mappingRef.Description,
			Source:      mappingRef.Url,
			Props: &[]oscalTypes.Property{
				{
					Name:  extensions.FrameworkProp,
					Value: mappingRef.Id,
					Ns:    extensions.TrestleNameSpace,
				},
			},
		}
	}

	var componentProps []oscalTypes.Property
	var groupNumber = 00

	for _, family := range catalog.ControlFamilies {
		for _, control := range family.Controls {
			for _, assessment := range control.AssessmentRequirements {
				ruleProps := makeRule(assessment, groupNumber)
				groupNumber += 1
				mapRule(assessment.Id, control.GuidelineMappings, mappingSet)
				componentProps = append(componentProps, ruleProps...)
			}
		}
	}

	controlImplementations := make([]oscalTypes.ControlImplementationSet, 0, len(mappingSet))
	for _, ciSet := range mappingSet {
		controlImplementations = append(controlImplementations, ciSet)
	}

	component := oscalTypes.DefinedComponent{
		UUID:                   uuid.NewUUID(),
		Title:                  targetComponent,
		Type:                   componentType,
		Props:                  utils.NilIfEmpty(&componentProps),
		ControlImplementations: utils.NilIfEmpty(&controlImplementations),
	}
	c.targetComponents[catalog.Metadata.Id] = component
	return c
}

func (c *DefinitionBuilder) AddValidationComponent(source string, evaluations []layer4.AssessmentPlan) *DefinitionBuilder {
	var componentProps []oscalTypes.Property
	var groupNumber = 00

	for _, eval := range evaluations {
		for _, assessment := range eval.Assessments {
			for _, procedure := range assessment.Procedures {
				checkProps := makeCheck(assessment.RequirementId, procedure, groupNumber)
				groupNumber += 1
				componentProps = append(componentProps, checkProps...)
			}

		}
	}

	component := oscalTypes.DefinedComponent{
		UUID:  uuid.NewUUID(),
		Type:  "validation",
		Title: source,
		Props: utils.NilIfEmpty(&componentProps),
	}
	c.validationComponent = append(c.validationComponent, component)
	return c
}

// AddParameterModifiers takes parameter modifications for a given Layer 2 reference and creates OSCAL set-parameters
// on the associated control set implementations. This will only take effect if the Layer 2 Catalogs has been added
// through AddTargetComponent.
func (c *DefinitionBuilder) AddParameterModifiers(referenceId string, modifiers []layer3.ParameterModifier) *DefinitionBuilder {
	component, found := c.targetComponents[referenceId]
	if found {
		// Create set parameters
		setParams := make([]oscalTypes.SetParameter, 0, len(modifiers))
		for _, param := range modifiers {
			setParameter := oscalTypes.SetParameter{
				ParamId: param.TargetId,
				Values:  []string{convertToString(param.Value)},
			}
			setParams = append(setParams, setParameter)
		}

		// Turn params modifiers into set parameters
		if component.ControlImplementations != nil {
			for i := range *component.ControlImplementations {
				ci := &(*component.ControlImplementations)[i]
				if ci.SetParameters == nil {
					ci.SetParameters = &setParams
				} else {
					*ci.SetParameters = append(*ci.SetParameters, setParams...)
				}
			}
		}
	}
	return c
}

func (c *DefinitionBuilder) Build() oscalTypes.ComponentDefinition {
	metadata := models.NewSampleMetadata()
	metadata.Title = c.title
	metadata.Version = c.version

	var allComponent []oscalTypes.DefinedComponent
	for _, comp := range c.targetComponents {
		allComponent = append(allComponent, comp)
	}
	allComponent = append(allComponent, c.validationComponent...)

	return oscalTypes.ComponentDefinition{
		UUID:       uuid.NewUUID(),
		Metadata:   metadata,
		Components: utils.NilIfEmpty(&allComponent),
	}
}

func makeRule(requirement layer2.AssessmentRequirement, groupNumber int) []oscalTypes.Property {
	remark := fmt.Sprintf("rule_set_%d", groupNumber)

	ruleIdProp := oscalTypes.Property{
		Name:    extensions.RuleIdProp,
		Value:   requirement.Id,
		Ns:      extensions.TrestleNameSpace,
		Remarks: remark,
	}

	ruleDescProp := oscalTypes.Property{
		Name:    extensions.RuleDescriptionProp,
		Value:   strings.ReplaceAll(requirement.Text, "\n", "\\n"),
		Ns:      extensions.TrestleNameSpace,
		Remarks: remark,
	}

	props := []oscalTypes.Property{
		ruleIdProp,
		ruleDescProp,
	}

	if len(requirement.RecommendedParameters) > 0 {
		for i, parameter := range requirement.RecommendedParameters {
			paramIdProp := oscalTypes.Property{
				Name:    fmt.Sprintf("%s_%d", extensions.ParameterIdProp, i),
				Value:   parameter.Id,
				Ns:      extensions.TrestleNameSpace,
				Remarks: remark,
			}

			paramDescProp := oscalTypes.Property{
				Name:    fmt.Sprintf("%s_%d", extensions.ParameterDescriptionProp, i),
				Value:   strings.ReplaceAll(parameter.Description, "\n", "\\n"),
				Ns:      extensions.TrestleNameSpace,
				Remarks: remark,
			}

			if parameter.Default != nil {
				parameterDefaultProp := oscalTypes.Property{
					Name:    fmt.Sprintf("%s_%d", extensions.ParameterDefaultProp, i),
					Value:   convertToString(parameter.Default),
					Ns:      extensions.TrestleNameSpace,
					Remarks: remark,
				}
				props = append(props, parameterDefaultProp)
			}

			props = append(props, paramDescProp, paramIdProp)
		}
	}

	return props
}

func makeCheck(ruleId string, procedure layer4.AssessmentProcedure, groupNumber int) []oscalTypes.Property {
	remark := fmt.Sprintf("rule_set_%d", groupNumber)
	ruleIdProp := oscalTypes.Property{
		Name:    extensions.RuleIdProp,
		Value:   ruleId,
		Ns:      extensions.TrestleNameSpace,
		Remarks: remark,
	}

	checkIdProp := oscalTypes.Property{
		Name:    extensions.CheckIdProp,
		Value:   procedure.Id,
		Ns:      extensions.TrestleNameSpace,
		Remarks: remark,
	}

	checkDescProp := oscalTypes.Property{
		Name:    extensions.CheckDescriptionProp,
		Value:   procedure.Description,
		Ns:      extensions.TrestleNameSpace,
		Remarks: remark,
	}
	return []oscalTypes.Property{
		ruleIdProp,
		checkIdProp,
		checkDescProp,
	}
}

func mapRule(ruleId string, mappings []layer2.Mapping, ciSets map[string]oscalTypes.ControlImplementationSet) {
	ruleIdProp := oscalTypes.Property{
		Name:  extensions.RuleIdProp,
		Value: ruleId,
		Ns:    extensions.TrestleNameSpace,
	}

	for _, mapping := range mappings {
		targetCI, ok := ciSets[mapping.ReferenceId]
		if !ok {
			continue
		}
		for _, entry := range mapping.Entries {
			createOrUpdateImplementedRequirement(ruleIdProp, entry.ReferenceId, &targetCI)
		}
		ciSets[mapping.ReferenceId] = targetCI
	}
}

func createOrUpdateImplementedRequirement(ruleIdProp oscalTypes.Property, identifier string, controlImplementation *oscalTypes.ControlImplementationSet) {
	var found bool
	for i := range controlImplementation.ImplementedRequirements {
		if controlImplementation.ImplementedRequirements[i].ControlId == identifier {
			if controlImplementation.ImplementedRequirements[i].Props == nil {
				controlImplementation.ImplementedRequirements[i].Props = &[]oscalTypes.Property{}
			}
			*controlImplementation.ImplementedRequirements[i].Props = append(*controlImplementation.ImplementedRequirements[i].Props, ruleIdProp)
			found = true
			break
		}
	}

	// Check if it is set, this means create a new one
	if !found {
		implRequirement := oscalTypes.ImplementedRequirementControlImplementation{
			UUID:      uuid.NewUUID(),
			ControlId: utils.NormalizeControl(identifier),
			Props:     &[]oscalTypes.Property{ruleIdProp},
		}
		controlImplementation.ImplementedRequirements = append(controlImplementation.ImplementedRequirements, implRequirement)
	}
}

func convertToString(val any) string {
	if val == nil {
		return ""
	}
	switch v := val.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	default:
		return fmt.Sprint(v)
	}
}
