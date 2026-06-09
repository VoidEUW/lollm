# Installation

This guide explains how to build `lollm` and install it into a custom
directory on your `$PATH` (e.g. `~/.local/bin`).

## Requirements

- [Go](https://go.dev/dl/) 1.26 or newer (`go version`)
- `claude` (Claude Code) and your provider — LM Studio and/or Ollama —
  available on your `$PATH`

## 1. Build

From the repository root:

```sh
make build        # produces build/lollm
```

or without `make`:

```sh
go build -o build/lollm ./src
```

## 2. Install into a custom path

### Option A — `make install` (recommended)

`make install` copies the binary to `$PREFIX/bin`. `PREFIX` defaults to
`~/.local`, so the binary lands in `~/.local/bin/lollm`:

```sh
make install                       # -> ~/.local/bin/lollm
```

Pick a different location with `PREFIX`:

```sh
make install PREFIX=~/.local       # -> ~/.local/bin/lollm   (your case)
make install PREFIX=/usr/local     # -> /usr/local/bin/lollm (may need sudo)
make install PREFIX=~/bin          # -> ~/bin/lollm
```

### Option B — copy it yourself

```sh
mkdir -p ~/.local/bin
cp build/lollm ~/.local/bin/
chmod +x ~/.local/bin/lollm
```

## 3. Make sure the directory is on your `$PATH`

Check whether your target directory is already on the `PATH`:

```sh
echo "$PATH" | tr ':' '\n' | grep -F "$HOME/.local/bin" && echo "already on PATH"
```

If nothing prints, add it. You are on **zsh**, so edit `~/.zshrc`:

```sh
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

<details>
<summary>Other shells</summary>

```sh
# bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc && source ~/.bashrc

# fish
fish_add_path ~/.local/bin
```
</details>

## 4. Verify

```sh
which lollm        # -> /Users/<you>/.local/bin/lollm
lollm version      # -> lollm 0.1.0
lollm config path  # -> ~/.lollm/config.toml
```

The first time you run `lollm` (or `lollm config`), it creates the default
config at `~/.lollm/config.toml`. See [README.md](README.md) for usage and
configuration.

## Updating

Rebuild and reinstall after pulling changes:

```sh
make install PREFIX=~/.local
```

## Uninstalling

```sh
rm ~/.local/bin/lollm     # the binary
rm -rf ~/.lollm           # optional: your config
```
