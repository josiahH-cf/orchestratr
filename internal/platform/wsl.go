package platform

import (
	"os"
	"strings"
)

const procVersionPath = "/proc/version"

// IsWSL2 reports whether the current environment is running inside
// WSL2 by checking /proc/version for Microsoft/WSL2 markers.
func IsWSL2() bool {
	return isWSL2WithPath(procVersionPath)
}

// isWSL2WithPath is the testable core of IsWSL2. It reads the given
// file and checks for WSL2/Microsoft markers.
func isWSL2WithPath(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	lower := strings.ToLower(string(data))
	return strings.Contains(lower, "microsoft") || strings.Contains(lower, "wsl2")
}

// WSL2Warning returns the warning message to display when orchestratr
// is running inside WSL2.
func WSL2Warning() string {
	return `⚠ orchestratr is running inside WSL2.
  System-wide hotkeys require the daemon to run on the Windows side.

  Recommended: install orchestratr in Windows PowerShell instead:
    .\install.ps1

  Continuing will install, but hotkeys will not work from WSL2.
  Apps launched by orchestratr from Windows can still target WSL commands.`
}
