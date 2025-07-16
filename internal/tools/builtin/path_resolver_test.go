package builtin

import (
	"path/filepath"
	"testing"
)

func TestPathResolver_ResolvePath(t *testing.T) {
	workingDir := "/Users/test/project"
	resolver := NewPathResolver(workingDir)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "project relative path with leading slash - src",
			input:    "/src/core/ContainerTypes.ts",
			expected: filepath.Join(workingDir, "src/core/ContainerTypes.ts"),
		},
		{
			name:     "project relative path with leading slash - lib",
			input:    "/lib/utils/helpers.js",
			expected: filepath.Join(workingDir, "lib/utils/helpers.js"),
		},
		{
			name:     "project relative path with leading slash - components",
			input:    "/components/Button.tsx",
			expected: filepath.Join(workingDir, "components/Button.tsx"),
		},
		{
			name:     "project config file at root",
			input:    "/package.json",
			expected: filepath.Join(workingDir, "package.json"),
		},
		{
			name:     "project config file at root - tsconfig",
			input:    "/tsconfig.json",
			expected: filepath.Join(workingDir, "tsconfig.json"),
		},
		{
			name:     "relative path without leading slash",
			input:    "src/core/ContainerTypes.ts",
			expected: filepath.Join(workingDir, "src/core/ContainerTypes.ts"),
		},
		{
			name:     "current directory",
			input:    ".",
			expected: workingDir,
		},
		{
			name:     "parent directory",
			input:    "..",
			expected: filepath.Dir(workingDir),
		},
		{
			name:     "true system absolute path - usr bin",
			input:    "/usr/local/bin/node",
			expected: "/usr/local/bin/node",
		},
		{
			name:     "true system absolute path - etc",
			input:    "/etc/hosts",
			expected: "/etc/hosts",
		},
		{
			name:     "true system absolute path - var",
			input:    "/var/log/system.log",
			expected: "/var/log/system.log",
		},
		{
			name:     "empty path",
			input:    "",
			expected: workingDir,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolver.ResolvePath(tt.input)
			// Clean both paths for comparison
			expected := filepath.Clean(tt.expected)
			result = filepath.Clean(result)

			if result != expected {
				t.Errorf("ResolvePath(%q) = %q, want %q", tt.input, result, expected)
			}
		})
	}
}

func TestPathResolver_isProjectRelativePath(t *testing.T) {
	resolver := NewPathResolver("/test/project")

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "src directory",
			input:    "/src/main.ts",
			expected: true,
		},
		{
			name:     "lib directory",
			input:    "/lib/utils.js",
			expected: true,
		},
		{
			name:     "components directory",
			input:    "/components/Button.tsx",
			expected: true,
		},
		{
			name:     "package.json config file",
			input:    "/package.json",
			expected: true,
		},
		{
			name:     "tsconfig.json config file",
			input:    "/tsconfig.json",
			expected: true,
		},
		{
			name:     "README.md file",
			input:    "/README.md",
			expected: true,
		},
		{
			name:     "system path - usr",
			input:    "/usr/local/bin",
			expected: false,
		},
		{
			name:     "system path - etc",
			input:    "/etc/hosts",
			expected: false,
		},
		{
			name:     "system path - var",
			input:    "/var/log/app.log",
			expected: false,
		},
		{
			name:     "system path - home",
			input:    "/home/user/file.txt",
			expected: false,
		},
		{
			name:     "relative path without slash",
			input:    "src/main.ts",
			expected: false,
		},
		{
			name:     "empty path",
			input:    "",
			expected: false,
		},
		{
			name:     "root path only",
			input:    "/",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolver.isProjectRelativePath(tt.input)
			if result != tt.expected {
				t.Errorf("isProjectRelativePath(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
