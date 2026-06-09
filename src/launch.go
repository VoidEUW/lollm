package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// launch resolves the active provider and replaces the current process with
// the configured command (claude, ollama launch, ...). The lollm process is
// gone after this returns successfully, so the child inherits the terminal
// directly — important for an interactive TUI like Claude Code.
//
// extra are arguments passed straight through to the underlying command, so
// `lollm -c "fix the bug"` becomes `claude --model <model> -c "fix the bug"`.
func launch(cfg *Config, extra []string) error {
	name := cfg.Active
	if v := os.Getenv("LOLLM_PROVIDER"); v != "" {
		name = v // one-off override without touching the config file
	}
	if name == "" {
		return fmt.Errorf("no active provider set (run: lollm config use <provider>)")
	}

	p, ok := cfg.Providers[name]
	if !ok {
		return fmt.Errorf("active provider %q is not defined in the config", name)
	}

	command := p.Command
	if command == "" {
		command = "claude"
	}
	path, err := exec.LookPath(command)
	if err != nil {
		return fmt.Errorf("command %q not found in PATH", command)
	}

	argv := []string{command}
	argv = append(argv, p.Args...)
	if p.Model != "" {
		argv = append(argv, "--model", p.Model)
	}
	argv = append(argv, extra...)

	// Build the child environment from a map so provider settings override
	// inherited values instead of producing duplicate keys.
	env := map[string]string{}
	for _, kv := range os.Environ() {
		if i := strings.IndexByte(kv, '='); i >= 0 {
			env[kv[:i]] = kv[i+1:]
		}
	}
	if p.BaseURL != "" {
		env["ANTHROPIC_BASE_URL"] = p.BaseURL
	}
	if p.AuthToken != "" {
		env["ANTHROPIC_AUTH_TOKEN"] = p.AuthToken
	}
	for k, v := range p.Env {
		env[k] = v
	}

	envv := make([]string, 0, len(env))
	for k, v := range env {
		envv = append(envv, k+"="+v)
	}

	fmt.Fprintf(os.Stderr, "lollm: launching %q -> %s\n", name, strings.Join(argv, " "))
	return syscall.Exec(path, argv, envv)
}
