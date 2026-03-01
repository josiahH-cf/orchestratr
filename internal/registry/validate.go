package registry

import (
	"fmt"
	"strings"
)

// reservedChords lists chord keys that cannot be assigned to apps.
var reservedChords = map[string]bool{
	"?":     true,
	"space": true,
}

// wellKnownKeys lists multi-character key names that are valid chords.
var wellKnownKeys = map[string]bool{
	"f1": true, "f2": true, "f3": true, "f4": true,
	"f5": true, "f6": true, "f7": true, "f8": true,
	"f9": true, "f10": true, "f11": true, "f12": true,
	"space": true, "tab": true, "enter": true, "escape": true,
	"backspace": true, "delete": true, "insert": true,
	"home": true, "end": true, "pageup": true, "pagedown": true,
	"up": true, "down": true, "left": true, "right": true,
}

// ValidateConfig checks a Config for errors. It returns a slice of all
// validation errors found (empty if the config is valid).
func ValidateConfig(cfg *Config) []error {
	var errs []error

	for i, app := range cfg.Apps {
		errs = append(errs, validateAppEntry(i, &app)...)
	}

	errs = append(errs, checkDuplicateChords(cfg.Apps)...)

	return errs
}

// validateAppEntry checks a single app entry for required fields and
// valid values.
func validateAppEntry(index int, app *AppEntry) []error {
	var errs []error
	prefix := fmt.Sprintf("apps[%d]", index)

	if app.Name == "" {
		errs = append(errs, fmt.Errorf("%s: name is required", prefix))
	} else {
		prefix = fmt.Sprintf("apps[%d] (%s)", index, app.Name)
	}

	if app.Chord == "" {
		errs = append(errs, fmt.Errorf("%s: chord is required", prefix))
	} else if err := validateChord(app.Chord); err != nil {
		errs = append(errs, fmt.Errorf("%s: %w", prefix, err))
	}

	if app.Command == "" {
		errs = append(errs, fmt.Errorf("%s: command is required", prefix))
	}

	if err := validateEnvironment(app.Environment); err != nil {
		errs = append(errs, fmt.Errorf("%s: %w", prefix, err))
	}

	return errs
}

// validateChord checks that a chord key is a single character or a
// well-known key name, and is not reserved.
func validateChord(chord string) error {
	lower := strings.ToLower(chord)

	if reservedChords[lower] {
		return fmt.Errorf("chord %q is reserved", chord)
	}

	// Single character chords are always valid.
	if len([]rune(chord)) == 1 {
		return nil
	}

	// Multi-character must be a well-known key name.
	if !wellKnownKeys[lower] {
		return fmt.Errorf("chord %q is not a valid single key or known key name", chord)
	}

	return nil
}

// validateEnvironment checks that the environment field is one of the
// allowed values: "native", "wsl", or "wsl:<distro>".
func validateEnvironment(env string) error {
	switch {
	case env == "":
		return nil
	case env == "native":
		return nil
	case env == "wsl":
		return nil
	case strings.HasPrefix(env, "wsl:"):
		distro := strings.TrimPrefix(env, "wsl:")
		if distro == "" {
			return fmt.Errorf("environment %q has empty distro name", env)
		}
		return nil
	default:
		return fmt.Errorf("environment %q is not valid; use \"native\", \"wsl\", or \"wsl:<distro>\"", env)
	}
}

// checkDuplicateChords returns errors for any chord keys assigned to
// more than one app.
func checkDuplicateChords(apps []AppEntry) []error {
	seen := make(map[string]string) // chord -> first app name
	var errs []error

	for _, app := range apps {
		if app.Chord == "" {
			continue
		}
		lower := strings.ToLower(app.Chord)
		if first, ok := seen[lower]; ok {
			errs = append(errs, fmt.Errorf(
				"duplicate chord %q: assigned to both %q and %q",
				app.Chord, first, app.Name,
			))
		} else {
			seen[lower] = app.Name
		}
	}

	return errs
}
