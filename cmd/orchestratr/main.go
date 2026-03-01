package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/josiahH-cf/orchestratr/internal/registry"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		fmt.Fprintln(stdout, "orchestratr — system-wide app launcher")
		fmt.Fprintln(stdout, "Usage: orchestratr [start|stop|status|reload|list|version]")
		return nil
	}

	switch args[0] {
	case "version":
		fmt.Fprintln(stdout, "orchestratr v0.0.0-dev")
		return nil

	case "list":
		return runList(stdout, stderr)

	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

// runList loads the config and prints the app registry as a table.
func runList(stdout, stderr io.Writer) error {
	path := registry.DefaultConfigPath()

	// Allow override via environment variable for testing.
	if envPath := os.Getenv("ORCHESTRATR_CONFIG"); envPath != "" {
		path = envPath
	}

	cfg, err := registry.LoadAndValidate(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("config not found at %s\nRun orchestratr once to generate a default config, or set ORCHESTRATR_CONFIG", path)
		}
		return fmt.Errorf("loading config: %w", err)
	}

	if len(cfg.Apps) == 0 {
		fmt.Fprintln(stdout, "No apps registered. Edit your config file to add apps.")
		fmt.Fprintf(stdout, "Config: %s\n", path)
		return nil
	}

	w := tabwriter.NewWriter(stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tCHORD\tCOMMAND\tENV\tDESCRIPTION")
	for _, app := range cfg.Apps {
		env := app.Environment
		if env == "" {
			env = "native"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			app.Name, app.Chord, app.Command, env, app.Description)
	}
	return w.Flush()
}


