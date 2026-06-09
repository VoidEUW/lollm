// Command lollm routes Claude Code to a local LLM provider (LM Studio,
// Ollama, ...) using settings stored in ~/.lollm/config.toml.
//
//	lollm                run Claude Code against the active provider
//	lollm <args...>      same, passing <args...> through to the command
//	lollm config ...     inspect and edit the configuration
package main

import (
	"fmt"
	"os"
)

const version = "0.1.0"

const usage = `lollm - route Claude Code to a local LLM provider

Usage:
  lollm                  launch the active provider (passes extra args through)
  lollm <args...>        e.g. lollm -c "keep going"
  lollm config ...       manage the configuration (lollm config help)
  lollm version          print the lollm version
  lollm help             show this help

Config lives in ~/.lollm/config.toml (override the dir with $LOLLM_HOME).
Set $LOLLM_PROVIDER to override the active provider for a single run.
`

func main() {
	args := os.Args[1:]

	if len(args) > 0 {
		switch args[0] {
		case "config":
			fail(runConfig(args[1:]))
			return
		case "help", "--help", "-h":
			fmt.Print(usage)
			return
		case "version", "--version":
			fmt.Printf("lollm %s\n", version)
			return
		}
	}

	cfg, err := loadConfig()
	fail(err)
	// launch only returns on error; on success it replaces this process.
	fail(launch(cfg, args))
}

func fail(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "lollm: %s\n", err)
		os.Exit(1)
	}
}
