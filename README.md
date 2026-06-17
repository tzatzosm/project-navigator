# Project Navigator (`pn`)

A small CLI for managing and navigating your development projects: organize them
into nested groups, then jump into one (and open it in your editor) from an
interactive browser. Written in Go — a single static binary, no runtime to
install.

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

Downloads the right binary for your OS/arch from the latest GitHub Release,
verifies its checksum, and installs it. Override with `PN_VERSION` or
`PN_INSTALL_DIR`. No Go required.

### Homebrew (personal tap)

```bash
brew install tzatzosm/tap/project-navigator
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
that a small wrapper evaluates. Add this to your `~/.bashrc` / `~/.zshrc`
(or run `make shell-init`):

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

## Publishing a Homebrew release

The formula (`Formula/project-navigator.rb`) is a simple Go build — no vendored
dependencies. To cut a release:

1. Tag and push:
   ```bash
   git tag v0.2.0 && git push origin v0.2.0
   ```
2. Fill the tarball `sha256` into the formula:
   ```bash
   curl -sL https://github.com/tzatzosm/project-navigator/archive/refs/tags/v0.2.0.tar.gz | shasum -a 256
   ```
3. Copy the formula into the `tzatzosm/homebrew-tap` repo under
   `Formula/project-navigator.rb` and push.

## License

[MIT](LICENSE) © Marsel Tzatzos
