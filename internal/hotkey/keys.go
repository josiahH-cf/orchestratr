// Package hotkey provides the cross-platform hotkey engine for capturing
// a global leader key and matching chord keystrokes to registered actions.
package hotkey

import "fmt"

// Modifier represents a keyboard modifier key.
type Modifier uint8

const (
	// ModCtrl represents the Ctrl modifier.
	ModCtrl Modifier = 1 << iota
	// ModAlt represents the Alt modifier.
	ModAlt
	// ModShift represents the Shift modifier.
	ModShift
	// ModSuper represents the Super/Win/Cmd modifier.
	ModSuper
)

// String returns a human-readable representation of the modifier.
func (m Modifier) String() string {
	parts := []string{}
	if m&ModCtrl != 0 {
		parts = append(parts, "ctrl")
	}
	if m&ModAlt != 0 {
		parts = append(parts, "alt")
	}
	if m&ModShift != 0 {
		parts = append(parts, "shift")
	}
	if m&ModSuper != 0 {
		parts = append(parts, "super")
	}
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += "+"
		}
		result += p
	}
	return result
}

// Key represents a key combination: zero or more modifiers plus a base key.
type Key struct {
	Modifiers Modifier
	Code      string // lowercase base key name, e.g. "space", "a", "f1"
}

// String returns the canonical string form of the key, e.g. "ctrl+space".
func (k Key) String() string {
	mod := k.Modifiers.String()
	if mod == "" {
		return k.Code
	}
	return mod + "+" + k.Code
}

// Equal reports whether two keys are identical.
func (k Key) Equal(other Key) bool {
	return k.Modifiers == other.Modifiers && k.Code == other.Code
}

// ParseKey parses a key string like "ctrl+shift+a" or "space" into a Key.
// Modifier names are case-insensitive. The last token is the base key.
// Returns an error if the string is empty or contains only modifiers.
func ParseKey(s string) (Key, error) {
	if s == "" {
		return Key{}, fmt.Errorf("empty key string")
	}

	tokens := splitPlus(s)
	if len(tokens) == 0 {
		return Key{}, fmt.Errorf("empty key string")
	}

	var mods Modifier
	baseIdx := -1

	for i, tok := range tokens {
		tok = toLower(tok)
		tokens[i] = tok // normalize

		switch tok {
		case "ctrl", "control":
			mods |= ModCtrl
		case "alt", "option", "opt":
			mods |= ModAlt
		case "shift":
			mods |= ModShift
		case "super", "win", "cmd", "command", "meta":
			mods |= ModSuper
		default:
			if baseIdx != -1 {
				return Key{}, fmt.Errorf("multiple base keys: %q and %q", tokens[baseIdx], tok)
			}
			baseIdx = i
		}
	}

	if baseIdx == -1 {
		return Key{}, fmt.Errorf("no base key in %q (only modifiers)", s)
	}

	return Key{Modifiers: mods, Code: tokens[baseIdx]}, nil
}

// knownConflicts maps leader key strings to the system they conflict with.
var knownConflicts = map[string]string{
	"ctrl+space":      "common IDE autocomplete / CJK input method toggle",
	"ctrl+escape":     "Windows Start menu",
	"super+space":     "macOS Spotlight / GNOME Activities",
	"alt+space":       "Windows window menu",
	"ctrl+alt+delete": "system interrupt on Windows/Linux",
}

// CheckConflicts returns a human-readable warning if the given key is known
// to conflict with common OS or application shortcuts. Returns "" if no
// conflict is known.
func CheckConflicts(k Key) string {
	canonical := k.String()
	if desc, ok := knownConflicts[canonical]; ok {
		return fmt.Sprintf("leader key %q may conflict with: %s", canonical, desc)
	}
	return ""
}

// splitPlus splits s on '+' characters, trimming whitespace from each token
// and discarding empty tokens.
func splitPlus(s string) []string {
	var tokens []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == '+' {
			tok := trimSpace(s[start:i])
			if tok != "" {
				tokens = append(tokens, tok)
			}
			start = i + 1
		}
	}
	return tokens
}

// toLower returns a lowercased copy of s (ASCII only, sufficient for key names).
func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

// trimSpace removes leading and trailing ASCII whitespace.
func trimSpace(s string) string {
	i, j := 0, len(s)
	for i < j && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r') {
		i++
	}
	for j > i && (s[j-1] == ' ' || s[j-1] == '\t' || s[j-1] == '\n' || s[j-1] == '\r') {
		j--
	}
	return s[i:j]
}
