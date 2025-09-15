module github.com/complytime/gemara2oscal

go 1.24.4

require (
	github.com/oscal-compass/oscal-sdk-go v0.0.5
	github.com/ossf/gemara v0.4.0
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.2 // indirect
	golang.org/x/text v0.25.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	github.com/defenseunicorns/go-oscal v0.7.0
	github.com/goccy/go-yaml v1.18.0
	github.com/stretchr/testify v1.11.1
)

// Points to experiments/oscal-transformation branch (temporarily)
replace github.com/ossf/gemara => github.com/jpower432/sci v0.0.0-20250821140246-72e7d4e3fb1e
