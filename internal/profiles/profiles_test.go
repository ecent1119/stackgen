package profiles

import (
	"testing"

	"github.com/stackgen-cli/stackgen/internal/models"
)

func TestAvailableProfiles(t *testing.T) {
	profiles := AvailableProfiles()
	
	if len(profiles) < 5 {
		t.Errorf("Expected at least 5 profiles, got %d", len(profiles))
	}

	// Check web-app profile exists
	found := false
	for _, p := range profiles {
		if p.Name == "web-app" {
			found = true
			break
		}
	}
	if !found {
		t.Error("web-app profile should exist")
	}
}

func TestGetProfile(t *testing.T) {
	profile := GetProfile("api")
	
	if profile == nil {
		t.Fatal("api profile should exist")
	}

	if profile.Name != "api" {
		t.Errorf("Expected api, got %s", profile.Name)
	}

	// API profile should have Go runtime
	found := false
	for _, rt := range profile.Runtimes {
		if rt.Type == models.RuntimeGo {
			found = true
			break
		}
	}
	if !found {
		t.Error("api profile should include Go runtime")
	}
}

func TestGetProfileNotFound(t *testing.T) {
	profile := GetProfile("nonexistent")
	
	if profile != nil {
		t.Error("Non-existent profile should return nil")
	}
}

func TestBuildProjectFromProfile(t *testing.T) {
	profile := GetProfile("web-app")
	if profile == nil {
		t.Fatal("web-app profile should exist")
	}

	project := BuildProjectFromProfile(profile, "myproject", ".")

	if project.Name != "myproject" {
		t.Errorf("Expected myproject, got %s", project.Name)
	}

	if len(project.Datastores) == 0 {
		t.Error("Project should have datastores")
	}

	if len(project.Runtimes) == 0 {
		t.Error("Project should have runtimes")
	}
}

func TestFullstackProfile(t *testing.T) {
	profile := GetProfile("fullstack")
	if profile == nil {
		t.Fatal("fullstack profile should exist")
	}

	// Should have multiple datastores
	if len(profile.Datastores) < 3 {
		t.Error("fullstack profile should have at least 3 datastores")
	}

	// Should have multiple runtimes
	if len(profile.Runtimes) < 2 {
		t.Error("fullstack profile should have at least 2 runtimes")
	}
}

func TestDotnetProfile(t *testing.T) {
	profile := GetProfile("dotnet")
	if profile == nil {
		t.Fatal("dotnet profile should exist")
	}

	// Should use MSSQL
	found := false
	for _, ds := range profile.Datastores {
		if ds == models.DatastoreMSSQL {
			found = true
			break
		}
	}
	if !found {
		t.Error("dotnet profile should include MSSQL")
	}

	// Should have C# runtime
	found = false
	for _, rt := range profile.Runtimes {
		if rt.Type == models.RuntimeCSharp {
			found = true
			break
		}
	}
	if !found {
		t.Error("dotnet profile should include C# runtime")
	}
}
