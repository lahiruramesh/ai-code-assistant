package templates_manager

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// TemplateManager handles template operations
type TemplateManager struct {
	templatesPath string
	projectsPath  string
}

// Template represents available templates
type Template struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Description string `json:"description"`
	Type        string `json:"type"` // "react" or "nextjs"
}

// NewTemplateManager creates a new template manager
func NewTemplateManager(templatesPath, projectsPath string) *TemplateManager {
	return &TemplateManager{
		templatesPath: templatesPath,
		projectsPath:  projectsPath,
	}
}

// GetAvailableTemplates returns list of available templates
func (tm *TemplateManager) GetAvailableTemplates() ([]Template, error) {
	templates := []Template{
		{
			Name:        "react-basic",
			Path:        filepath.Join(tm.templatesPath, "react-basic"),
			Description: "Basic React application with Create React App setup",
			Type:        "react",
		},
		{
			Name:        "nextjs-shadcn",
			Path:        filepath.Join(tm.templatesPath, "nextjs-shadcn"),
			Description: "Next.js 14 with shadcn/ui components and Tailwind CSS",
			Type:        "nextjs",
		},
	}

	// Verify templates exist
	var validTemplates []Template
	for _, template := range templates {
		if _, err := os.Stat(template.Path); err == nil {
			validTemplates = append(validTemplates, template)
		}
	}

	return validTemplates, nil
}

// CopyTemplate copies a template to the project directory
func (tm *TemplateManager) CopyTemplate(templateName, projectName string) error {
	templatePath := filepath.Join(tm.templatesPath, templateName)
	projectPath := filepath.Join(tm.projectsPath, projectName)

	// Check if template exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return fmt.Errorf("template %s does not exist", templateName)
	}

	// Check if project directory already exists
	if _, err := os.Stat(projectPath); err == nil {
		return fmt.Errorf("project %s already exists", projectName)
	}

	// Create project directory
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %v", err)
	}

	// Copy template files
	err := tm.copyDir(templatePath, projectPath)
	if err != nil {
		// Clean up on error
		os.RemoveAll(projectPath)
		return fmt.Errorf("failed to copy template: %v", err)
	}

	// Update package.json with project name
	if err := tm.updatePackageJSON(projectPath, projectName); err != nil {
		return fmt.Errorf("failed to update package.json: %v", err)
	}

	return nil
}

// copyDir recursively copies a directory
func (tm *TemplateManager) copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			// Create directory
			return os.MkdirAll(destPath, info.Mode())
		} else {
			// Copy file
			return tm.copyFile(path, destPath)
		}
	})
}

// copyFile copies a single file
func (tm *TemplateManager) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// updatePackageJSON updates the package.json with project-specific information
func (tm *TemplateManager) updatePackageJSON(projectPath, projectName string) error {
	packageJSONPath := filepath.Join(projectPath, "package.json")

	// Check if package.json exists
	if _, err := os.Stat(packageJSONPath); os.IsNotExist(err) {
		return nil // No package.json to update
	}

	// Read package.json
	content, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return err
	}

	// Replace template name with project name
	updatedContent := strings.ReplaceAll(string(content), "react-basic-template", projectName)
	updatedContent = strings.ReplaceAll(updatedContent, "nextjs-shadcn-template", projectName)

	// Write updated content
	return os.WriteFile(packageJSONPath, []byte(updatedContent), 0644)
}

// GenerateProjectName generates a unique project name
func (tm *TemplateManager) GenerateProjectName(baseName string) string {
	// Sanitize base name (replace spaces with dashes, lowercase)
	sanitized := strings.ToLower(strings.ReplaceAll(baseName, " ", "-"))

	// Remove special characters except dashes
	var result strings.Builder
	for _, r := range sanitized {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	sanitized = result.String()

	// Add timestamp suffix for uniqueness
	timestamp := fmt.Sprintf("%03d", os.Getpid()%1000)

	return fmt.Sprintf("%s-%s", sanitized, timestamp)
}

// GetProjectPath returns the full path for a project
func (tm *TemplateManager) GetProjectPath(projectName string) string {
	return filepath.Join(tm.projectsPath, projectName)
}
