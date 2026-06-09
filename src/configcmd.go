package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
)

const configUsage = `lollm config - manage ~/.lollm/config.toml

Usage:
  lollm config                 show the current config
  lollm config show            show the current config
  lollm config path            print the config file path
  lollm config list            list providers (* marks the active one)
  lollm config use <provider>  set the active provider
  lollm config get <key>       print a single value
  lollm config set <key> <val> set a value (creates providers on demand)
  lollm config edit            open the config in $EDITOR
  lollm config init [--force]  (re)create the default config

Keys:
  active
  <provider>.command
  <provider>.base_url
  <provider>.auth_token
  <provider>.model
  <provider>.args            value is split on whitespace
  <provider>.env.<NAME>      set empty to delete the variable

Examples:
  lollm config use ollama
  lollm config set ollama.model qwen3-coder:30b
  lollm config set lmstudio.base_url http://localhost:1234
  lollm config set lmstudio.env.CLAUDE_CODE_ATTRIBUTION_HEADER 0
`

// runConfig handles `lollm config ...`.
func runConfig(args []string) error {
	if len(args) == 0 {
		return showConfig()
	}

	switch args[0] {
	case "show":
		return showConfig()

	case "help", "-h", "--help":
		fmt.Print(configUsage)
		return nil

	case "path":
		path, err := configPath()
		if err != nil {
			return err
		}
		fmt.Println(path)
		return nil

	case "list", "providers":
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		names := make([]string, 0, len(cfg.Providers))
		for name := range cfg.Providers {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			marker := "  "
			if name == cfg.Active {
				marker = "* "
			}
			fmt.Printf("%s%s\n", marker, name)
		}
		return nil

	case "use":
		if len(args) != 2 {
			return fmt.Errorf("usage: lollm config use <provider>")
		}
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		if err := cfg.set("active", args[1]); err != nil {
			return err
		}
		if err := saveConfig(cfg); err != nil {
			return err
		}
		fmt.Printf("active provider is now %q\n", args[1])
		return nil

	case "get":
		if len(args) != 2 {
			return fmt.Errorf("usage: lollm config get <key>")
		}
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		val, err := cfg.get(args[1])
		if err != nil {
			return err
		}
		fmt.Println(val)
		return nil

	case "set":
		if len(args) < 3 {
			return fmt.Errorf("usage: lollm config set <key> <value>")
		}
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		key := args[1]
		value := strings.Join(args[2:], " ")
		if err := cfg.set(key, value); err != nil {
			return err
		}
		if err := saveConfig(cfg); err != nil {
			return err
		}
		fmt.Printf("%s = %s\n", key, value)
		return nil

	case "edit":
		return editConfig()

	case "init":
		force := len(args) > 1 && (args[1] == "--force" || args[1] == "-f")
		return initConfig(force)

	default:
		return fmt.Errorf("unknown config subcommand %q (try: lollm config help)", args[0])
	}
}

// showConfig prints the config path followed by its rendered contents.
func showConfig() error {
	path, err := configPath()
	if err != nil {
		return err
	}
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(cfg); err != nil {
		return err
	}
	fmt.Printf("# %s\n\n%s", path, buf.String())
	return nil
}

// editConfig opens the config file in $EDITOR (falling back to vi).
func editConfig() error {
	path, err := configPath()
	if err != nil {
		return err
	}
	// Make sure the file exists before editing.
	if _, err := loadConfig(); err != nil {
		return err
	}
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// initConfig writes the default config, refusing to clobber an existing file
// unless force is set.
func initConfig(force bool) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if !force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("%s already exists (use: lollm config init --force)", path)
		}
	}
	if err := saveConfig(defaultConfig()); err != nil {
		return err
	}
	fmt.Printf("wrote default config to %s\n", path)
	return nil
}
