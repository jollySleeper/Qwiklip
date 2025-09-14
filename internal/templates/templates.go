package templates

import (
	"embed"
	"fmt"
	"html/template"
)

//go:embed *.html
var FS embed.FS

// TemplateSet holds the parsed HTML templates
type TemplateSet struct {
	Index *template.Template
	Error *template.Template
}

// Load parses and validates all required templates
func Load() (*TemplateSet, error) {
	// Parse all templates from embedded filesystem
	tmpl, err := template.ParseFS(FS, "index.html", "error.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse embedded templates: %w", err)
	}

	// Extract and validate individual templates
	indexTemplate := tmpl.Lookup("index.html")
	errorTemplate := tmpl.Lookup("error.html")

	if indexTemplate == nil {
		return nil, fmt.Errorf("index.html template not found in embedded filesystem")
	}
	if errorTemplate == nil {
		return nil, fmt.Errorf("error.html template not found in embedded filesystem")
	}

	return &TemplateSet{
		Index: indexTemplate,
		Error: errorTemplate,
	}, nil
}
