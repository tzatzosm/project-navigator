# Project Navigator (`pn`)

A small CLI for managing and navigating your development projects: organize them
into nested groups, then jump into one (and open it in your editor) from an
interactive browser. Written in Go — a single static binary, no runtime to
install.

![pn demo](demo/demo.gif)

## Features

- **Interactive browser** — arrow-key through nested groups and projects, drill
  in and out with `← Back`.
- **`cd` into projects** — selecting a project changes your shell's directory
  (via a tiny shell wrapper, see below).
- **Editor integration** — open a project in your editor with one keystroke;
  per-project overrides fall back to a global default.
- **Nested groups** — arbitrarily deep group hierarchy, rendered as a tree.
- **OS-native config** — JSON in the standard config dir for your platform.
- **Resilient** — missing config is created on first run; projects whose path
  no longer exists are flagged with `⚠️` instead of crashing.

## Install

### Quick install (prebuilt binary)

```bash
curl -fsSL https://raw.githubusercontent.com/tzatzosm/project-navigator/main/install.sh | bash
```

Downloads the right binary for your OS/arch from GitHub Releases, verifies its
checksum, and installs it. No Go required.

**Where it installs:** `PN_INSTALL_DIR` if set, otherwise `/usr/local/bin` when
that directory is writable (the installer uses `sudo` if needed), falling back
to `~/.local/bin`. If the chosen directory isn't on your `PATH`, the installer
prints the `export PATH=…` line to add.

**Which version:** `PN_VERSION` if set (e.g. `v0.2.0`), otherwise the latest
release (`latest`).

```bash
# Examples — note the variables go on the `bash` side of the pipe, not `curl`
curl -fsSL https://raw.githubusercontent.com/tzatzosm/project-navigator/main/install.sh | PN_VERSION=v0.2.0 bash
curl -fsSL https://raw.githubusercontent.com/tzatzosm/project-navigator/main/install.sh | PN_INSTALL_DIR="$HOME/bin" bash
```

### go install

```bash
go install github.com/tzatzosm/project-navigator/cmd/pn@latest
```

Ensure your Go bin directory is on `PATH` (`export PATH="$(go env GOPATH)/bin:$PATH"`).

### From source

```bash
make build        # produces ./pn
make install      # go install into your Go bin dir
```

### Shell integration (required for `cd`)

A process can't change your shell's directory, so `pn` prints a `cd <path>` line
that a small wrapper evaluates. **The `curl | bash` installer adds this wrapper
to your `~/.zshrc` / `~/.bashrc` automatically** (idempotently; set
`PN_NO_WRAPPER=1` to skip). For other install methods, add it yourself (or run
`make shell-init`):

```bash
pn() {
  result=$(command pn "$@")
  if echo "$result" | grep -q "^cd "; then
    eval "$result"
  else
    echo "$result"
  fi
}
```

Reload your shell (`exec $SHELL`) and you're set.

## Usage

```
pn add [path]   # add the current dir (or path) as a project
pn open         # interactive browser — navigate groups and open a project
pn groups       # add, rename, or delete groups
pn editors      # manage your editors list (add/remove/set default)
pn config       # set the global default editor
pn list         # print all projects as a tree
```

Running `pn` with no arguments opens the browser.

### Example

```bash
cd ~/work/api
pn add                 # name it, drop it in a group, pick an editor
pn list                # see everything as a tree
pn                     # open the browser, pick "API Service", hit enter → you're cd'd in
```

## Config

Stored as `config.json` in your OS config directory:

- **macOS** — `~/Library/Application Support/project-navigator/`
- **Linux** — `~/.config/project-navigator/` (or `$XDG_CONFIG_HOME`)
- **Windows** — `%APPDATA%\project-navigator\`

Override with `PN_CONFIG_DIR`. A config from a pre-1.0 `~/.project-navigator/`
location is migrated automatically on first run.

## Develop

```bash
make test         # go vet + compile check
make build        # ./pn
```

The interactive UI uses [huh](https://github.com/charmbracelet/huh); styled
output (tree, table, colors) uses [lipgloss](https://github.com/charmbracelet/lipgloss).

## Releasing

Releases are fully automated. Push a `vX.Y.Z` tag and the
[`release`](.github/workflows/release.yml) workflow runs
[GoReleaser](https://goreleaser.com) to cross-compile every platform and publish
the binaries + checksums to a GitHub Release:

```bash
git tag v0.3.0 && git push origin v0.3.0
```

The `curl | bash` installer above always points at the latest release.

## License

[MIT](LICENSE) © Marsel Tzatzos
