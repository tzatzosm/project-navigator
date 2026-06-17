# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`pn` is a Go CLI for navigating/opening dev projects. It was originally a Python single-file script (see `project-navigator-prompt.md` for the original spec); it has since been **rewritten in Go** so it ships as a single static binary with no runtime dependencies. The Python sources are gone — Go is the source of truth.

## Layout

- `cmd/pn/` — `package main`, split across:
  - `main.go` — arg dispatch, `usage()`, `emitCD()`, `launchEditor()`, `isDir`/`commandExists`.
  - `config.go` — the on-disk model (`Config`/`Group`/`Project`/`Editor`), OS-aware config paths, load/save, legacy migration, and the `container` abstraction.
  - `commands.go` — one function per subcommand (`cmdAdd`/`cmdOpen`/`cmdGroups`/`cmdConfig`/`cmdEditors`/`cmdList`).
  - `ui.go` — `huh` prompt wrappers and `lipgloss` styled output (table + tree).
- `Formula/project-navigator.rb` — Homebrew formula (Go build).
- `Makefile` — `build` / `install` (`go install`) / `shell-init` / `test`.

## Build / test / run

```bash
make build              # -> ./pn
make test               # go vet ./... + compile check
go run ./cmd/pn list    # run a subcommand directly
```

There is no unit-test suite yet; `make test` is vet + build. The interactive prompts require a TTY, so they can't be exercised by piping stdin — drive them with a pty if you need to (and answer the terminal's OSC-11/cursor queries, or `huh` blocks before rendering).

## Architecture notes that span multiple concerns

- **The `cd` trick is central.** A process can't change its parent shell's cwd. Contract: `pn` **prints `cd <path>` to stdout**, and a user-installed shell wrapper `eval`s any output starting with `cd `. Therefore:
  - **stdout carries only the `cd` line** (`emitCD`). **Everything else — every prompt and all styled output — goes to stderr.** `huh` forms are run via `runForm()` with `WithOutput(os.Stderr)`/`WithInput(os.Stdin)`; `main()` calls `lipgloss.SetDefaultRenderer(lipgloss.NewRenderer(os.Stderr))` so color detection keys off stderr (stdout is often a captured pipe). Breaking this — printing anything else to stdout — corrupts the `grep "^cd "` wrapper.
  - Launched editors get nil `Cmd.Stdout`/`Stderr` (→ the null device) so a chatty editor can't pollute the `cd` channel.

- **One recursive model backs everything.** Root config and every group satisfy the `container` interface (`groups()`/`projects()` returning slice pointers), so `walkGroups`, `removeGroup`, the `pn open` browser, `pn list`'s tree, and `pn add`/`pn groups` mutation all share one traversal instead of duplicating it. Editor resolution is layered: per-project `Editor` (a `*string`, nil ⇒ use default) → global `DefaultEditor`.

- **Config location** is resolved per-OS by `configDir()` (`PN_CONFIG_DIR` > `XDG_CONFIG_HOME` > macOS `~/Library/Application Support/project-navigator`, Windows `%APPDATA%`, Linux `~/.config`). `migrateLegacyConfig()` moves a pre-1.0 `~/.project-navigator/config.json` over once. `normalize()` replaces nil slices with `[]` so JSON round-trips cleanly.

- **Resilience:** missing config → create with defaults; a saved path that no longer exists → marked `⚠️` in `list`/`open`, never a crash; an unresolved editor → warn and still `cd`.

## Dependencies

`github.com/charmbracelet/huh` (all interactive prompts) and `github.com/charmbracelet/lipgloss` + its `table`/`tree` subpackages (all styled output). Keep new deps minimal; the value proposition is a small self-contained binary.

## Homebrew formula

`Formula/project-navigator.rb` is a `depends_on "go" => :build` formula that runs `go build … ./cmd/pn` — far simpler than the old Python virtualenv form (no vendored resources). The `url` is a GitHub release tarball; its `sha256` must be filled in after tagging. The formula must also be copied into the `tzatzosm/homebrew-tap` repo for `brew install tzatzosm/tap/project-navigator` to work. Validate with `brew style Formula/project-navigator.rb`.
