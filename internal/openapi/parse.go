// Package openapi parses the parts of an OpenAPI document needed by ContractLens.
package openapi

import (
	"fmt"

	"go.yaml.in/yaml/v3"
)

const MaxDocumentBytes = 10 << 20 // 10 MiB

type Document struct {
	Version string         `yaml:"openapi"`
	Paths   map[string]any `yaml:"paths"`
}

// Parse reads an OpenAPI document encoded as JSON or YAML and checks that it
func Parse(data []byte) (*Document, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("OpenAPI document is empty")
	}

	if len(data) > MaxDocumentBytes {
		return nil, fmt.Errorf("OpenAPI document exceeds the %d-byte limit", MaxDocumentBytes)
	}

	var doc Document
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse OpenAPI document: %w", err)
	}

	if doc.Version == "" {
		return nil, fmt.Errorf("OpenAPI document is missing the openapi version")
	}

	if doc.Paths == nil {
		return nil, fmt.Errorf("OpenAPI document is missing the paths object")
	}

	return &doc, nil
}
