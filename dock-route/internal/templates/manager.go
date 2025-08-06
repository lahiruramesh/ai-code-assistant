package templates

import (
    "embed"
    "fmt"
    "path/filepath"
    
    "gopkg.in/yaml.v3"
)
//go:embed data/*
var templatesFS embed.FS

type Manager struct {
    templates map[string]*Template
}

func NewManager() *Manager {
    return &Manager{
        templates: make(map[string]*Template),
    }
}

func (m *Manager) GetTemplate(appType string) (*Template, error) {
    if template, exists := m.templates[appType]; exists {
        return template, nil
    }
    
    // Load template from embedded filesystem
    templatePath := filepath.Join("data", appType, "template.yaml")
    data, err := templatesFS.ReadFile(templatePath)
    if err != nil {
		fmt.Println(err)
        return nil, fmt.Errorf("template not found for app type: %s", appType)
    }
    
    var template Template
    if err := yaml.Unmarshal(data, &template); err != nil {
        return nil, fmt.Errorf("failed to parse template: %w", err)
    }
    
    // Load Dockerfile content
    dockerfilePath := filepath.Join("data", appType, "Dockerfile")
    dockerfileContent, err := templatesFS.ReadFile(dockerfilePath)
    if err != nil {
        return nil, fmt.Errorf("failed to load Dockerfile: %w", err)
    }
    
    template.Dockerfile = string(dockerfileContent)
    
    // Cache the template
    m.templates[appType] = &template
    
    return &template, nil
}

func (m *Manager) ListTemplates() []string {
    var types []string
    
    entries, err := templatesFS.ReadDir("data")
    if err != nil {
        return types
    }
    
    for _, entry := range entries {
        if entry.IsDir() {
            types = append(types, entry.Name())
        }
    }
    
    return types
}
