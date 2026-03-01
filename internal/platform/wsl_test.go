package platform

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsWSL2_WithMicrosoftInProcVersion(t *testing.T) {
	tmpDir := t.TempDir()
	procVersion := filepath.Join(tmpDir, "version")
	content := "Linux version 5.15.90.1-microsoft-standard-WSL2 (oe-user@oe-host) (x86_64-msft-linux-gcc (GCC) 12.2.0, GNU ld (GNU Binutils) 2.38)"
	if err := os.WriteFile(procVersion, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if !isWSL2WithPath(procVersion) {
		t.Error("isWSL2WithPath() = false, want true for Microsoft WSL2 kernel")
	}
}

func TestIsWSL2_WithWSL2InProcVersion(t *testing.T) {
	tmpDir := t.TempDir()
	procVersion := filepath.Join(tmpDir, "version")
	content := "Linux version 5.15.90-WSL2 something"
	if err := os.WriteFile(procVersion, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if !isWSL2WithPath(procVersion) {
		t.Error("isWSL2WithPath() = false, want true for WSL2 kernel")
	}
}

func TestIsWSL2_WithRegularLinux(t *testing.T) {
	tmpDir := t.TempDir()
	procVersion := filepath.Join(tmpDir, "version")
	content := "Linux version 6.1.0-18-amd64 (debian-kernel@lists.debian.org) (gcc-12 (Debian 12.2.0-14) 12.2.0, GNU ld (GNU Binutils for Debian) 2.40)"
	if err := os.WriteFile(procVersion, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if isWSL2WithPath(procVersion) {
		t.Error("isWSL2WithPath() = true, want false for regular Linux kernel")
	}
}

func TestIsWSL2_MissingProcVersion(t *testing.T) {
	if isWSL2WithPath("/nonexistent/proc/version") {
		t.Error("isWSL2WithPath() = true for nonexistent file")
	}
}

func TestIsWSL2_CaseInsensitive(t *testing.T) {
	tmpDir := t.TempDir()
	procVersion := filepath.Join(tmpDir, "version")
	content := "Linux version 5.15.90-MICROSOFT-STANDARD-WSL2"
	if err := os.WriteFile(procVersion, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if !isWSL2WithPath(procVersion) {
		t.Error("isWSL2WithPath() = false, want true (case-insensitive)")
	}
}

func TestWSL2Warning(t *testing.T) {
	msg := WSL2Warning()
	if msg == "" {
		t.Error("WSL2Warning() returned empty string")
	}
	if !strings.Contains(msg, "WSL2") {
		t.Errorf("WSL2Warning() = %q, expected to mention WSL2", msg)
	}
	if !strings.Contains(msg, "Windows") {
		t.Errorf("WSL2Warning() = %q, expected to mention Windows", msg)
	}
}
