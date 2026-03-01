package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("orchestratr — system-wide app launcher")
		fmt.Println("Usage: orchestratr [start|stop|status|reload|version]")
		os.Exit(0)
	}

	switch os.Args[1] {
	case "version":
		fmt.Println("orchestratr v0.0.0-dev")
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
