package middleware

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// OpenAPISpec represents the structure of OpenAPI specification
type OpenAPISpec struct {
	Paths map[string]PathItem `json:"paths"`
}

// PathItem represents a path item in OpenAPI spec
type PathItem struct {
	Get    *Operation `json:"get,omitempty"`
	Post   *Operation `json:"post,omitempty"`
	Put    *Operation `json:"put,omitempty"`
	Delete *Operation `json:"delete,omitempty"`
	Patch  *Operation `json:"patch,omitempty"`
}

// Operation represents an operation in OpenAPI spec
type Operation struct {
	Security []map[string][]string `json:"security,omitempty"`
}

// SecurityRequirement represents a security requirement
type SecurityRequirement struct {
	Method       string
	Path         string
	RequiresAuth bool
}

// ParseOpenAPISpec parses OpenAPI specification files and extracts security requirements
func ParseOpenAPISpec(specDir string) ([]SecurityRequirement, error) {
	var requirements []SecurityRequirement

	// Read all YAML files in the spec directory
	files, err := filepath.Glob(filepath.Join(specDir, "*.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to read spec directory: %w", err)
	}

	for _, file := range files {
		// Convert YAML to JSON first (simplified approach)
		// In production, you might want to use a proper YAML parser
		yamlContent, err := ioutil.ReadFile(file)
		if err != nil {
			continue // Skip files that can't be read
		}

		// For now, let's manually define the security requirements
		// This is a simplified approach - in production you'd parse YAML properly
		reqs := extractSecurityFromYAML(string(yamlContent))
		requirements = append(requirements, reqs...)
	}

	return requirements, nil
}

// extractSecurityFromYAML extracts security requirements from YAML content
// This is a simplified implementation - in production use proper YAML parsing
func extractSecurityFromYAML(content string) []SecurityRequirement {
	var requirements []SecurityRequirement

	// Simple pattern matching for security requirements
	lines := strings.Split(content, "\n")
	var currentPath string
	var currentMethod string

	fmt.Printf("DEBUG: Parsing YAML content with %d lines\n", len(lines))

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Detect path
		if strings.HasPrefix(line, "  /") && strings.HasSuffix(line, ":") {
			currentPath = strings.TrimSpace(strings.TrimSuffix(line, ":"))
			fmt.Printf("DEBUG: Found path: %s\n", currentPath)
		}

		// Detect HTTP method
		if strings.HasPrefix(line, "    ") && (strings.Contains(line, "get:") ||
			strings.Contains(line, "post:") || strings.Contains(line, "put:") ||
			strings.Contains(line, "delete:") || strings.Contains(line, "patch:")) {
			currentMethod = strings.TrimSuffix(strings.TrimSpace(line), ":")
			fmt.Printf("DEBUG: Found method: %s\n", currentMethod)
		}

		// Detect security requirement
		if strings.Contains(line, "bearerAuth:") {
			fmt.Printf("DEBUG: Found bearerAuth at line %d: %s\n", i+1, line)
			if currentPath != "" && currentMethod != "" {
				requirements = append(requirements, SecurityRequirement{
					Method:       strings.ToUpper(currentMethod),
					Path:         currentPath,
					RequiresAuth: true,
				})
				fmt.Printf("DEBUG: Added requirement: %s %s\n", strings.ToUpper(currentMethod), currentPath)
				// Reset for next endpoint
				currentPath = ""
				currentMethod = ""
			}
		}
	}

	return requirements
}

// LoadSecurityRequirements loads security requirements from OpenAPI specs
func LoadSecurityRequirements(specDir string) (map[string]bool, error) {
	requirements, err := ParseOpenAPISpec(specDir)
	if err != nil {
		return nil, err
	}

	securityMap := make(map[string]bool)

	// Debug: print all requirements
	fmt.Printf("DEBUG: Found %d security requirements:\n", len(requirements))
	for _, req := range requirements {
		key := fmt.Sprintf("%s %s", req.Method, req.Path)
		securityMap[key] = req.RequiresAuth
		fmt.Printf("  - %s: %t\n", key, req.RequiresAuth)
	}

	return securityMap, nil
}
