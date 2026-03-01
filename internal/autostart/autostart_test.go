package autostart

import (
	"testing"
)

func TestNewManagerReturnsNonNil(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager() returned nil")
	}
}

func TestNewManagerDescriptionNotEmpty(t *testing.T) {
	m := NewManager()
	if m.Description() == "" {
		t.Error("NewManager().Description() returned empty string")
	}
}
