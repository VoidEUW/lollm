package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Provider describes one LLM backend that Claude Code can be pointed at.
//
// The launcher turns a Provider into a command line plus an environment:
//
//	<command> <args...> [--model <model>] <passthrough args...>
//
// with ANTHROPIC_BASE_URL / ANTHROPIC_AUTH_TOKEN derived from BaseURL /
// AuthToken and any extra variables from Env layered on top of the parent
// environment.
type Provider struct {
	Command   string            `toml:"command,omitempty"`
	Args      []string          `toml:"args,omitempty"`
	BaseURL   string            `toml:"base_url,omitempty"`
	AuthToken string            `toml:"auth_token,omitempty"`
	Model     string            `toml:"model,omitempty"`
	Env       map[string]string `toml:"env,omitempty"`
}

// Config is the whole ~/.lollm/config.toml document.
type Config struct {
	Active    string              `toml:"active"`
	Providers map[string]Provider `toml:"providers"`
}

// configPath returns the path to the config file, honouring $LOLLM_HOME and
// then $HOME. The returned directory may not exist yet.
func configPath() (string, error) {
	if home := os.Getenv("LOLLM_HOME"); home != "" {
		return filepath.Join(home, "config.toml"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".lollm", "config.toml"), nil
}

// defaultConfig is what gets written on first run.
func defaultConfig() *Config {
	return &Config{
		Active: "lmstudio",
		Providers: map[string]Provider{
			"lmstudio": {
				Command:   "claude",
				BaseURL:   "http://localhost:1234",
				AuthToken: "lmstudio",
				Model:     "qwen3.5-9b-claude-4.6-opus-reasoning-distilled-v2",
				Env: map[string]string{
					"CLAUDE_CODE_ATTRIBUTION_HEADER": "0",
				},
			},
			"ollama": {
				Command: "ollama",
				Args:    []string{"launch", "claude"},
			},
		},
	}
}

// loadConfig reads the config file, creating it with defaults if it does not
// exist yet.
func loadConfig() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		cfg := defaultConfig()
		if err := saveConfig(cfg); err != nil {
			return nil, err
		}
		fmt.Fprintf(os.Stderr, "lollm: created default config at %s\n", path)
		return cfg, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	if cfg.Providers == nil {
		cfg.Providers = map[string]Provider{}
	}
	return &cfg, nil
}

// saveConfig writes the config back to disk, creating the directory if needed.
func saveConfig(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(cfg); err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}

// set assigns a dotted key to a value. Recognised keys:
//
//	active
//	<provider>.command
//	<provider>.base_url
//	<provider>.auth_token
//	<provider>.model
//	<provider>.args          (value is split on whitespace)
//	<provider>.env.<NAME>
//
// Unknown providers are created on demand.
func (c *Config) set(key, value string) error {
	if key == "active" {
		if _, ok := c.Providers[value]; !ok {
			return fmt.Errorf("no such provider %q", value)
		}
		c.Active = value
		return nil
	}

	parts := strings.SplitN(key, ".", 3)
	if len(parts) < 2 {
		return fmt.Errorf("invalid key %q (expected active or <provider>.<field>)", key)
	}
	name := parts[0]
	field := parts[1]

	p := c.Providers[name] // zero value if absent
	switch field {
	case "command":
		p.Command = value
	case "base_url":
		p.BaseURL = value
	case "auth_token":
		p.AuthToken = value
	case "model":
		p.Model = value
	case "args":
		if strings.TrimSpace(value) == "" {
			p.Args = nil
		} else {
			p.Args = strings.Fields(value)
		}
	case "env":
		if len(parts) != 3 || parts[2] == "" {
			return fmt.Errorf("env keys look like %s.env.NAME", name)
		}
		if p.Env == nil {
			p.Env = map[string]string{}
		}
		if value == "" {
			delete(p.Env, parts[2])
		} else {
			p.Env[parts[2]] = value
		}
	default:
		return fmt.Errorf("unknown field %q (command, base_url, auth_token, model, args, env.NAME)", field)
	}
	if c.Providers == nil {
		c.Providers = map[string]Provider{}
	}
	c.Providers[name] = p
	return nil
}

// get returns the string value for a dotted key, or an error if unknown.
func (c *Config) get(key string) (string, error) {
	if key == "active" {
		return c.Active, nil
	}
	parts := strings.SplitN(key, ".", 3)
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid key %q", key)
	}
	p, ok := c.Providers[parts[0]]
	if !ok {
		return "", fmt.Errorf("no such provider %q", parts[0])
	}
	switch parts[1] {
	case "command":
		return p.Command, nil
	case "base_url":
		return p.BaseURL, nil
	case "auth_token":
		return p.AuthToken, nil
	case "model":
		return p.Model, nil
	case "args":
		return strings.Join(p.Args, " "), nil
	case "env":
		if len(parts) != 3 {
			return "", fmt.Errorf("env keys look like %s.env.NAME", parts[0])
		}
		return p.Env[parts[2]], nil
	default:
		return "", fmt.Errorf("unknown field %q", parts[1])
	}
}
