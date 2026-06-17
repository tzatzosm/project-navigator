# Project Navigator (`pn`)

A small CLI for managing and navigating your development projects: organize them
into nested groups, then jump into one (and open it in your editor) from an
interactive browser. Written in Go — a single static binary, no runtime to
install.

## Install

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
