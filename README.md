# lollm

A tiny launcher that routes [Claude Code](https://claude.com/claude-code) to a
local LLM provider (LM Studio, Ollama, ...) so you don't have to export the
environment variables by hand every time.

```sh
lollm           # launch Claude Code against the active provider
```

instead of

```sh
export ANTHROPIC_BASE_URL=http://localhost:1234
export ANTHROPIC_AUTH_TOKEN=lmstudio
export CLAUDE_CODE_ATTRIBUTION_HEADER=0
claude --model qwen3.5-9b-claude-4.6-opus-reasoning-distilled-v2
```

## Layout

```
src/      Go source (package main)
build/    compiled binary (git-ignored)
Makefile  build / install helpers
go.mod    module root
```

## Install

```sh
make build               # -> build/lollm
make install             # -> ~/.local/bin/lollm (override with PREFIX=...)
```

Or directly with the Go toolchain:

```sh
go build -o build/lollm ./src
```

## Usage

```sh
lollm                 # run the active provider
lollm -c "keep going" # any extra args are passed straight to the command
lollm config ...      # manage configuration (see: lollm config help)
lollm version
lollm help
```

`LOLLM_PROVIDER=ollama lollm` overrides the active provider for a single run
without changing the config.

## Configuration

Config lives in `~/.lollm/config.toml` (override the directory with
`$LOLLM_HOME`). It is created with sensible defaults on first run.

```toml
active = "lmstudio"

[providers.lmstudio]
  command = "claude"
  base_url = "http://localhost:1234"      # -> ANTHROPIC_BASE_URL
  auth_token = "lmstudio"                  # -> ANTHROPIC_AUTH_TOKEN
  model = "qwen3.5-9b-claude-4.6-opus-reasoning-distilled-v2"
  [providers.lmstudio.env]
    CLAUDE_CODE_ATTRIBUTION_HEADER = "0"

[providers.ollama]
  command = "ollama"
  args = ["launch", "claude"]
  # model = "qwen3-coder:30b"
```

Each provider becomes:

```
<command> <args...> [--model <model>] <passthrough args...>
```

with `base_url`/`auth_token` mapped to the matching `ANTHROPIC_*` variables and
everything under `env` layered on top of the inherited environment.

### Editing from the terminal

```sh
lollm config                 # show the current config
lollm config list            # list providers (* = active)
lollm config use ollama      # switch active provider
lollm config get ollama.model
lollm config set ollama.model qwen3-coder:30b
lollm config set lmstudio.base_url http://localhost:1234
lollm config set lmstudio.env.CLAUDE_CODE_ATTRIBUTION_HEADER 0
lollm config edit            # open in $EDITOR
lollm config init --force    # reset to defaults
```

Setting a key on an unknown provider creates it. Set an `env.NAME` to an empty
value to delete that variable.
