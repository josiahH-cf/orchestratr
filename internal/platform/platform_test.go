package platform

import "testing"

func TestPlatformName(t *testing.T) {
	name := Name()
	if name == "" {
		t.Error("Name() returned empty string")
	}
}
