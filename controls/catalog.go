package controls

import (
	"fmt"
	"strings"
	"time"

	"github.com/defenseunicorns/go-oscal/src/pkg/uuid"
	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/oscal-compass/oscal-sdk-go/extensions"
	"github.com/oscal-compass/oscal-sdk-go/models"
	"github.com/ossf/gemara/layer1"

	"github.com/complytime/cac-transpiler/internal/utils"
)

// ToOSCALProfile creates an OSCAL Profile from the shared guidelines in a given
// Layer 1 Guidance Document.
func ToOSCALProfile(guidance layer1.GuidanceDocument) (oscalTypes.Profile, error) {
	metadata, err := createMetadata(guidance)
	if err != nil {
		return oscalTypes.Profile{}, fmt.Errorf("error creating profile metadata: %w", err)
	}

	mappingSet := make(map[string]oscalTypes.Import)
	for _, mappingRef := range guidance.Metadata.MappingReferences {
		mappingSet[mappingRef.Id] = oscalTypes.Import{
			Href: mappingRef.Url,
		}
	}

	for _, mapping := range guidance.SharedGuidelines {
		targetImport, ok := mappingSet[mapping.ReferenceId]
		if !ok {
			continue
		}

		withIds := make([]string, 0, len(mapping.Identifiers))
		for _, identifier := range mapping.Identifiers {
			withIds = append(withIds, utils.NormalizeControl(identifier))
		}

		selector := oscalTypes.SelectControlById{
			WithIds: &withIds,
		}
		targetImport.IncludeControls = &[]oscalTypes.SelectControlById{selector}
		mappingSet[mapping.ReferenceId] = targetImport
	}

	imports := make([]oscalTypes.Import, 0, len(mappingSet))
	for _, imp := range mappingSet {
		imports = append(imports, imp)
	}

	profile := oscalTypes.Profile{
		UUID:     uuid.NewUUID(),
		Imports:  imports,
		Metadata: metadata,
	}
	return profile, nil
}

// ToOSCALCatalog creates an OSCAL Catalog from the locally defined guidelines in a given
// Layer 1 Guidance Document.
func ToOSCALCatalog(guidance layer1.GuidanceDocument) (oscalTypes.Catalog, error) {
	metadata, err := createMetadata(guidance)
	if err != nil {
		return oscalTypes.Catalog{}, fmt.Errorf("error creating catalog metadata: %w", err)
	}

	// Create a resource map for control linking
	resourcesMap := make(map[string]string)
	backmatter := resourcesToBackMatter(guidance.Metadata.Resources)
	if backmatter != nil {
		for _, resource := range *backmatter.Resources {
			// Extract the id from the props
			props := *resource.Props
			id := props[0].Value
			resourcesMap[id] = resource.UUID
		}
	}

	var groups []oscalTypes.Group
	for _, category := range guidance.Categories {
		groups = append(groups, createControlGroup(category, resourcesMap))
	}

	catalog := oscalTypes.Catalog{
		UUID:       uuid.NewUUID(),
		Metadata:   metadata,
		Groups:     utils.NilIfEmpty(&groups),
		BackMatter: backmatter,
	}
	return catalog, nil
}

func createMetadata(guidance layer1.GuidanceDocument) (oscalTypes.Metadata, error) {
	metadata := models.NewSampleMetadata()
	metadata.Title = guidance.Metadata.Title

	published, err := time.Parse(time.DateOnly, guidance.Metadata.PublicationDate)
	if err != nil {
		return oscalTypes.Metadata{}, err
	}
	metadata.Published = &published

	lastModified, err := time.Parse(time.DateTime, guidance.Metadata.LastModified)
	if err != nil {
		return oscalTypes.Metadata{}, err
	}

	metadata.LastModified = lastModified
	metadata.Version = guidance.Metadata.Version

	authorRole := oscalTypes.Role{
		ID:          "author",
		Description: "Author of the guidance document",
		Title:       "Author",
	}

	author := oscalTypes.Party{
		UUID: uuid.NewUUID(),
		Type: "person",
		Name: guidance.Metadata.Author,
	}

	responsibleParty := oscalTypes.ResponsibleParty{
		PartyUuids: []string{author.UUID},
		RoleId:     authorRole.ID,
	}

	metadata.Parties = &[]oscalTypes.Party{author}
	metadata.Roles = &[]oscalTypes.Role{authorRole}
	metadata.ResponsibleParties = &[]oscalTypes.ResponsibleParty{responsibleParty}
	return metadata, nil
}

func createControlGroup(category layer1.Category, resourcesMap map[string]string) oscalTypes.Group {
	group := oscalTypes.Group{
		ID:    category.Id,
		Title: category.Title,
	}

	controlMap := make(map[string]oscalTypes.Control)
	for _, guideline := range category.Guidelines {
		control, parent := guidelineToControl(guideline, resourcesMap)

		if parent == "" {
			controlMap[control.ID] = control
		} else {
			parentControl := controlMap[parent]
			if parentControl.Controls == nil {
				parentControl.Controls = &[]oscalTypes.Control{}
			}
			*parentControl.Controls = append(*parentControl.Controls, control)
			controlMap[parent] = parentControl
		}
	}

	controls := make([]oscalTypes.Control, 0, len(controlMap))
	for _, control := range controlMap {
		controls = append(controls, control)
	}

	group.Controls = utils.NilIfEmpty(&controls)
	return group
}

func resourcesToBackMatter(resourceRefs []layer1.ResourceReference) *oscalTypes.BackMatter {
	var resources []oscalTypes.Resource
	for _, ref := range resourceRefs {
		resource := oscalTypes.Resource{
			UUID:        uuid.NewUUID(),
			Title:       ref.Title,
			Description: ref.Description,
			Props: &[]oscalTypes.Property{
				{
					Name:  "id",
					Value: ref.Id,
					Ns:    extensions.TrestleNameSpace,
				},
			},
			Rlinks: &[]oscalTypes.ResourceLink{
				{
					Href: ref.Url,
				},
			},
			Citation: &oscalTypes.Citation{
				Text: fmt.Sprintf(
					"%s. (%s). *%s*. %s",
					ref.IssuingBody,
					ref.PublicationDate,
					ref.Title,
					ref.Url),
			},
		}
		resources = append(resources, resource)
	}

	if len(resources) == 0 {
		return nil
	}

	backmatter := oscalTypes.BackMatter{
		Resources: &resources,
	}
	return &backmatter
}

func guidelineToControl(guideline layer1.Guideline, resourcesMap map[string]string) (oscalTypes.Control, string) {
	controlId := utils.NormalizeControl(guideline.Id)

	control := oscalTypes.Control{
		ID:    controlId,
		Title: guideline.Title,
	}

	var links []oscalTypes.Link
	for _, also := range guideline.SeeAlso {
		relatedLink := oscalTypes.Link{
			Href: fmt.Sprintf("#%s", also),
			Rel:  "related",
		}
		links = append(links, relatedLink)
	}

	for _, external := range guideline.ExternalReferences {
		ref, found := resourcesMap[external]
		if !found {
			continue
		}
		externalLink := oscalTypes.Link{
			Href: fmt.Sprintf("#%s", ref),
			Rel:  "reference",
		}
		links = append(links, externalLink)
	}

	// Top-level statements are required for controls
	smtPart := oscalTypes.Part{
		Name: "statement",
		ID:   fmt.Sprintf("%s_smt", controlId),
	}
	var subSmts []oscalTypes.Part
	for _, part := range guideline.GuidelineParts {
		subSmt := oscalTypes.Part{
			ID:    fmt.Sprintf("%s_smt.%s", controlId, part.Id),
			Prose: part.Prose,
			Title: part.Title,
		}

		if len(part.Recommendations) > 0 {
			gdnSubPart := oscalTypes.Part{
				Name:  "guidance",
				ID:    fmt.Sprintf("%s_smt.%s_gdn", controlId, part.Id),
				Prose: strings.Join(part.Recommendations, " "),
			}
			subSmt.Parts = &[]oscalTypes.Part{
				gdnSubPart,
			}
		}

		subSmts = append(subSmts, subSmt)
	}
	smtPart.Parts = utils.NilIfEmpty(&subSmts)
	control.Parts = &[]oscalTypes.Part{smtPart}

	if guideline.Objective != "" {
		// objective part
		objPart := oscalTypes.Part{
			Name:  "assessment-objective",
			ID:    fmt.Sprintf("%s_obj", controlId),
			Prose: guideline.Objective,
		}
		*control.Parts = append(*control.Parts, objPart)
	}

	if len(guideline.Recommendations) > 0 {
		// gdn part
		gdnPart := oscalTypes.Part{
			Name:  "guidance",
			ID:    fmt.Sprintf("%s_gdn", controlId),
			Prose: strings.Join(guideline.Recommendations, " "),
		}
		*control.Parts = append(*control.Parts, gdnPart)
	}

	return control, utils.NormalizeControl(guideline.BaseGuidelineID)
}
