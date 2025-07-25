// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright (c) 2025 Stefan Wold <ratler@stderr.eu>

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ratler/airuler/internal/compiler"
)

func TestIsValidTarget(t *testing.T) {
	tests := []struct {
		target   compiler.Target
		expected bool
	}{
		{compiler.TargetCursor, true},
		{compiler.TargetClaude, true},
		{compiler.TargetCline, true},
		{compiler.TargetCopilot, true},
		{compiler.Target("invalid"), false},
		{compiler.Target(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.target), func(t *testing.T) {
			result := isValidTarget(tt.target)
			if result != tt.expected {
				t.Errorf("isValidTarget(%s) = %v, expected %v", tt.target, result, tt.expected)
			}
		})
	}
}

func TestGetTargetNames(t *testing.T) {
	names := getTargetNames()

	expectedNames := []string{"cursor", "claude", "cline", "copilot", "roo"}

	if len(names) != len(expectedNames) {
		t.Errorf("getTargetNames() returned %d names, expected %d", len(names), len(expectedNames))
	}

	// Check that all expected names are present
	nameMap := make(map[string]bool)
	for _, name := range names {
		nameMap[name] = true
	}

	for _, expectedName := range expectedNames {
		if !nameMap[expectedName] {
			t.Errorf("getTargetNames() missing expected name: %s", expectedName)
		}
	}
}

func TestLoadTemplatesFromDirs(t *testing.T) {
	// Create temporary directory structure
	tempDir := t.TempDir()

	// Create template directories
	templatesDir := filepath.Join(tempDir, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("Failed to create templates directory: %v", err)
	}

	// Create test template files
	templates := map[string]string{
		"simple.tmpl":        "Hello {{.Name}}!",
		"cursor.tmpl":        "{{if eq .Target \"cursor\"}}Cursor rule{{end}}",
		"subdir/nested.tmpl": "Nested template content",
	}

	for path, content := range templates {
		fullPath := filepath.Join(templatesDir, path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write template %s: %v", path, err)
		}
	}

	// Also create a non-template file that should be ignored
	if err := os.WriteFile(filepath.Join(templatesDir, "ignore.txt"), []byte("ignore"), 0644); err != nil {
		t.Fatalf("Failed to write ignore file: %v", err)
	}

	// Test loading templates
	result, partials, err := loadTemplatesFromDirs([]string{templatesDir})
	if err != nil {
		t.Errorf("loadTemplatesFromDirs() failed: %v", err)
	}

	// Check that correct templates were loaded (main templates only, no partials)
	expectedTemplates := map[string]string{
		"simple":        "Hello {{.Name}}!",
		"cursor":        "{{if eq .Target \"cursor\"}}Cursor rule{{end}}",
		"subdir/nested": "Nested template content",
	}

	if len(result) != len(expectedTemplates) {
		t.Errorf("loadTemplatesFromDirs() returned %d templates, expected %d", len(result), len(expectedTemplates))
	}

	// Partials should be empty in this test since we don't have any partials/ directories
	if len(partials) != 0 {
		t.Errorf("loadTemplatesFromDirs() returned %d partials, expected 0", len(partials))
	}

	for name, expectedContent := range expectedTemplates {
		if templateSource, exists := result[name]; !exists {
			t.Errorf("loadTemplatesFromDirs() missing template: %s", name)
		} else if templateSource.Content != expectedContent {
			t.Errorf("loadTemplatesFromDirs() template %s content = %q, expected %q", name, templateSource.Content, expectedContent)
		}
	}

	// Check that non-template files were ignored
	if _, exists := result["ignore"]; exists {
		t.Error("loadTemplatesFromDirs() should not load non-.tmpl files")
	}
}

func TestLoadTemplatesFromDirsNonExistent(t *testing.T) {
	// Test with non-existent directory
	result, partials, err := loadTemplatesFromDirs([]string{"/path/that/does/not/exist"})
	if err != nil {
		t.Errorf("loadTemplatesFromDirs() with non-existent dir should not error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("loadTemplatesFromDirs() with non-existent dir should return empty map, got %d templates", len(result))
	}

	if len(partials) != 0 {
		t.Errorf("loadTemplatesFromDirs() with non-existent dir should return empty partials map, got %d partials", len(partials))
	}
}

func TestLoadTemplatesFromMultipleDirs(t *testing.T) {
	// Create temporary directory structure
	tempDir := t.TempDir()

	// Create two template directories
	dir1 := filepath.Join(tempDir, "templates1")
	dir2 := filepath.Join(tempDir, "templates2")

	for _, dir := range []string{dir1, dir2} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create templates in first directory
	template1 := filepath.Join(dir1, "template1.tmpl")
	if err := os.WriteFile(template1, []byte("Template 1 content"), 0644); err != nil {
		t.Fatalf("Failed to write template1: %v", err)
	}

	// Create templates in second directory
	template2 := filepath.Join(dir2, "template2.tmpl")
	if err := os.WriteFile(template2, []byte("Template 2 content"), 0644); err != nil {
		t.Fatalf("Failed to write template2: %v", err)
	}

	// Create overlapping template (same name in both dirs - should be overwritten)
	overlap1 := filepath.Join(dir1, "overlap.tmpl")
	overlap2 := filepath.Join(dir2, "overlap.tmpl")
	if err := os.WriteFile(overlap1, []byte("Overlap 1"), 0644); err != nil {
		t.Fatalf("Failed to write overlap1: %v", err)
	}
	if err := os.WriteFile(overlap2, []byte("Overlap 2"), 0644); err != nil {
		t.Fatalf("Failed to write overlap2: %v", err)
	}

	// Test loading from multiple directories
	result, partials, err := loadTemplatesFromDirs([]string{dir1, dir2})
	if err != nil {
		t.Errorf("loadTemplatesFromDirs() failed: %v", err)
	}

	// Check that templates from both directories were loaded
	expectedTemplates := map[string]string{
		"template1": "Template 1 content",
		"template2": "Template 2 content",
		"overlap":   "Overlap 2", // Second directory should override first
	}

	if len(result) != len(expectedTemplates) {
		t.Errorf("loadTemplatesFromDirs() returned %d templates, expected %d", len(result), len(expectedTemplates))
	}

	// No partials in this test
	if len(partials) != 0 {
		t.Errorf("loadTemplatesFromDirs() returned %d partials, expected 0", len(partials))
	}

	for name, expectedContent := range expectedTemplates {
		if templateSource, exists := result[name]; !exists {
			t.Errorf("loadTemplatesFromDirs() missing template: %s", name)
		} else if templateSource.Content != expectedContent {
			t.Errorf("loadTemplatesFromDirs() template %s content = %q, expected %q", name, templateSource.Content, expectedContent)
		}
	}
}

func TestLoadTemplatesWithPartials(t *testing.T) {
	// Create temporary directory with partials
	tempDir := t.TempDir()
	templatesDir := filepath.Join(tempDir, "templates")
	partialsDir := filepath.Join(templatesDir, "partials")

	if err := os.MkdirAll(partialsDir, 0755); err != nil {
		t.Fatalf("Failed to create partials directory: %v", err)
	}

	// Create main template
	mainTemplate := filepath.Join(templatesDir, "main.tmpl")
	if err := os.WriteFile(mainTemplate, []byte("{{template \"partials/header\" .}}"), 0644); err != nil {
		t.Fatalf("Failed to write main template: %v", err)
	}

	// Create partial template
	partialTemplate := filepath.Join(partialsDir, "header.tmpl")
	if err := os.WriteFile(partialTemplate, []byte("# {{.Name}}"), 0644); err != nil {
		t.Fatalf("Failed to write partial template: %v", err)
	}

	// Test loading templates
	templates, partials, err := loadTemplatesFromDirs([]string{templatesDir})
	if err != nil {
		t.Errorf("loadTemplatesFromDirs() failed: %v", err)
	}

	// Check that main template is in templates
	if len(templates) != 1 {
		t.Errorf("Expected 1 main template, got %d", len(templates))
	}
	if _, exists := templates["main"]; !exists {
		t.Error("Main template not found in templates")
	}

	// Check that partial is in partials
	if len(partials) != 1 {
		t.Errorf("Expected 1 partial, got %d", len(partials))
	}
	if _, exists := partials["partials/header"]; !exists {
		t.Error("Partial template not found in partials")
	}
}

func TestCompileTemplatesIntegration(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()

	// Save current directory and change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create templates directory and a simple template
	templatesDir := "templates"
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("Failed to create templates directory: %v", err)
	}

	templateContent := `# {{.Name}} Rule

{{if eq .Target "cursor"}}---
description: Test rule for {{.Target}}
globs: "**/*"
alwaysApply: true
---
{{end}}

This is a test rule for {{.Target}}.`

	templatePath := filepath.Join(templatesDir, "test.tmpl")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template: %v", err)
	}

	// Test compilation for a single target
	targets := []compiler.Target{compiler.TargetCursor}
	err = compileTemplates(targets)
	if err != nil {
		t.Errorf("compileTemplates() failed: %v", err)
	}

	// Check that output file was created
	outputPath := filepath.Join("compiled", "cursor", "test.mdc")
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Expected output file %s was not created", outputPath)
	}

	// Check output content
	outputContent, err := os.ReadFile(outputPath)
	if err != nil {
		t.Errorf("Failed to read output file: %v", err)
	}

	outputStr := string(outputContent)
	expectedParts := []string{
		"# test Rule",
		"---",
		"description: Test rule for cursor",
		"This is a test rule for cursor",
	}

	for _, part := range expectedParts {
		if !containsSubstring(outputStr, part) {
			t.Errorf("Output missing expected part: %s\nFull output:\n%s", part, outputStr)
		}
	}
}

func TestParseTemplateFrontMatter(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected *TemplateFrontMatter
		hasError bool
	}{
		{
			name: "Basic front matter",
			content: `---
description: "Test template"
claude_mode: "memory"
globs: "**/*.go"
---
Template content here`,
			expected: &TemplateFrontMatter{
				Description: "Test template",
				ClaudeMode:  "memory",
				Globs:       "**/*.go",
			},
			hasError: false,
		},
		{
			name: "Extended fields",
			content: `---
description: "Test template with extended fields"
claude_mode: "both"
globs: "**/*.{js,ts}"
project_type: "web-application"
language: "TypeScript"
framework: "React"
tags:
  - "frontend"
  - "spa"
always_apply: "true"
documentation: "docs/frontend.md"
style_guide: "Airbnb JavaScript style guide"
examples: "examples/react/"
custom:
  build_tool: "Vite"
  testing_framework: "Jest"
---
Template with extended fields`,
			expected: &TemplateFrontMatter{
				Description:   "Test template with extended fields",
				ClaudeMode:    "both",
				Globs:         "**/*.{js,ts}",
				ProjectType:   "web-application",
				Language:      "TypeScript",
				Framework:     "React",
				Tags:          []string{"frontend", "spa"},
				AlwaysApply:   "true",
				Documentation: "docs/frontend.md",
				StyleGuide:    "Airbnb JavaScript style guide",
				Examples:      "examples/react/",
				Custom: map[string]interface{}{
					"build_tool":        "Vite",
					"testing_framework": "Jest",
				},
			},
			hasError: false,
		},
		{
			name: "No front matter",
			content: `Template without front matter
Just content here`,
			expected: &TemplateFrontMatter{},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseTemplateFrontMatter(tt.content)

			if tt.hasError {
				if err == nil {
					t.Errorf("parseTemplateFrontMatter() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("parseTemplateFrontMatter() unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Errorf("parseTemplateFrontMatter() returned nil result")
				return
			}

			// Compare basic fields
			if result.Description != tt.expected.Description {
				t.Errorf("Description = %q, expected %q", result.Description, tt.expected.Description)
			}
			if result.ClaudeMode != tt.expected.ClaudeMode {
				t.Errorf("ClaudeMode = %q, expected %q", result.ClaudeMode, tt.expected.ClaudeMode)
			}
			if result.Globs != tt.expected.Globs {
				t.Errorf("Globs = %q, expected %q", result.Globs, tt.expected.Globs)
			}

			// Compare extended fields
			if result.ProjectType != tt.expected.ProjectType {
				t.Errorf("ProjectType = %q, expected %q", result.ProjectType, tt.expected.ProjectType)
			}
			if result.Language != tt.expected.Language {
				t.Errorf("Language = %q, expected %q", result.Language, tt.expected.Language)
			}
			if result.Framework != tt.expected.Framework {
				t.Errorf("Framework = %q, expected %q", result.Framework, tt.expected.Framework)
			}
			if result.AlwaysApply != tt.expected.AlwaysApply {
				t.Errorf("AlwaysApply = %q, expected %q", result.AlwaysApply, tt.expected.AlwaysApply)
			}
			if result.Documentation != tt.expected.Documentation {
				t.Errorf("Documentation = %q, expected %q", result.Documentation, tt.expected.Documentation)
			}
			if result.StyleGuide != tt.expected.StyleGuide {
				t.Errorf("StyleGuide = %q, expected %q", result.StyleGuide, tt.expected.StyleGuide)
			}
			if result.Examples != tt.expected.Examples {
				t.Errorf("Examples = %q, expected %q", result.Examples, tt.expected.Examples)
			}

			// Compare tags slice
			if len(result.Tags) != len(tt.expected.Tags) {
				t.Errorf("Tags length = %d, expected %d", len(result.Tags), len(tt.expected.Tags))
			} else {
				for i, tag := range result.Tags {
					if tag != tt.expected.Tags[i] {
						t.Errorf("Tags[%d] = %q, expected %q", i, tag, tt.expected.Tags[i])
					}
				}
			}

			// Compare custom map
			if tt.expected.Custom == nil && result.Custom != nil {
				t.Errorf("Custom should be nil but got %v", result.Custom)
			} else if tt.expected.Custom != nil {
				if result.Custom == nil {
					t.Errorf("Custom should not be nil")
				} else {
					if len(result.Custom) != len(tt.expected.Custom) {
						t.Errorf("Custom map length = %d, expected %d", len(result.Custom), len(tt.expected.Custom))
					}
					for key, expectedValue := range tt.expected.Custom {
						if actualValue, exists := result.Custom[key]; !exists {
							t.Errorf("Custom map missing key %q", key)
						} else if actualValue != expectedValue {
							t.Errorf("Custom[%q] = %v, expected %v", key, actualValue, expectedValue)
						}
					}
				}
			}
		})
	}
}
