package controls

import (
	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"

	"github.com/ossf/gemara/layer1"
)

// Thinly wrapping the upstream features here to minimize breaking changes

// ToOSCALProfile creates an OSCAL Profile from the shared guidelines in a given
// Layer 1 Guidance Document.
func ToOSCALProfile(guidance layer1.GuidanceDocument, guidanceDocHref string) (oscalTypes.Profile, error) {
	return guidance.ToOSCALProfile(guidanceDocHref)
}

// ToOSCALCatalog creates an OSCAL Catalog from the locally defined guidelines in a given
// Layer 1 Guidance Document.
func ToOSCALCatalog(guidance layer1.GuidanceDocument) (oscalTypes.Catalog, error) {
	return guidance.ToOSCALCatalog()
}
