package config

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultRespectsXDG(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "/tmp/xdg-fake")
	c, err := Default()
	if err != nil {
		t.Fatalf("Default: %v", err)
	}
	want := filepath.Join("/tmp/xdg-fake", "todo", "tasks.json")
	if c.StorePath != want {
		t.Errorf("StorePath = %q want %q", c.StorePath, want)
	}
}

func TestDefaultFallsBackToHome(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("HOME", "/tmp/home-fake")
	c, err := Default()
	if err != nil {
		t.Fatalf("Default: %v", err)
	}
	if !strings.HasSuffix(c.StorePath, filepath.Join(".local", "share", "todo", "tasks.json")) {
		t.Errorf("unexpected fallback path: %s", c.StorePath)
	}
}
