package platform

import "testing"

func TestCheckAccessibility(t *testing.T) {
	// On non-macOS platforms, CheckAccessibility always returns true, nil.
	// On macOS, it returns the actual accessibility status.
	granted, err := CheckAccessibility()
	if err != nil {
		t.Fatalf("CheckAccessibility() error: %v", err)
	}
	// On Linux/Windows test environments, we expect true.
	// On macOS CI without accessibility, it may return false.
	// This test just ensures the function runs without error.
	_ = granted
}

func TestAccessibilityPrompt(t *testing.T) {
	msg := AccessibilityPrompt()
	if msg == "" {
		t.Error("AccessibilityPrompt() returned empty string")
	}
}
