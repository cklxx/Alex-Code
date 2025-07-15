package version

import (
	"strings"
	"testing"
)

func TestGetVersion(t *testing.T) {
	version := GetVersion()

	// Should return just the version string (no prefix)
	if version == "" {
		t.Error("GetVersion should not return empty string")
	}

	// Should be the same as the Version variable
	if version != Version {
		t.Errorf("GetVersion should return Version variable, got: %s, expected: %s", version, Version)
	}
}

func TestVersionVariables(t *testing.T) {
	// Check that version variables are at least set to defaults
	if Version == "" {
		t.Error("Version should not be empty")
	}

	if GitCommit == "" {
		t.Error("GitCommit should not be empty")
	}

	if BuildTime == "" {
		t.Error("BuildTime should not be empty")
	}

	if GoVersion == "" {
		t.Error("GoVersion should not be empty")
	}

	if Platform == "" {
		t.Error("Platform should not be empty")
	}

	// GoVersion should start with "go"
	if !strings.HasPrefix(GoVersion, "go") {
		t.Errorf("GoVersion should start with 'go', got: %s", GoVersion)
	}

	// Platform should contain a slash
	if !strings.Contains(Platform, "/") {
		t.Errorf("Platform should contain '/', got: %s", Platform)
	}
}
