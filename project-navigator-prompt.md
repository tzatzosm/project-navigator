# Project Navigator CLI — Python Script Specification

Create a Python CLI tool for managing and navigating development projects. The tool should be a single Python script with no web framework dependencies.

## Core Requirements

The script must use the `inquirer` library for all interactive prompts and the `rich` library for styled terminal output. Configuration must be stored in a JSON file at `~/.project-navigator/config.json`. The JSON structure should support projects, nested groups, and a predefined editors list like this:

    {
      "default_editor": "code",
      "editors": [
        { "name": "VS Code", "command": "code" },
        { "name": "Neovim", "command": "nvim" },
        { "name": "WebStorm", "command": "webstorm" }
      ],
      "groups": [
        {
          "name": "Work",
          "subgroups": [
            {
              "name": "Backend",
              "subgroups": [],
              "projects": [
                { "name": "API Service", "path": "/home/user/work/api", "editor": null }
              ]
            }
          ],
          "projects": []
        }
      ],
      "projects": []
    }

## Commands

Implement the following CLI commands using `argparse`:

- `pn add [path]` — add the current directory (or given path) as a project. Interactively prompt for a name and optionally assign it to a group (or create a new one). Prompt for a project-specific editor override using an `inquirer` list built from the `editors` list in config — include a "Use default" option at the top. If no editors are configured yet, skip the editor selection and use the global default.
- `pn open` — open the interactive project browser (described below).
- `pn groups` — add, rename, or delete groups interactively.
- `pn config` — set the global default editor by selecting from the configured editors list, or fall through to `pn editors` if the list is empty.
- `pn editors` — manage the editors list: add a new editor (prompt for display name and CLI command), remove an existing one, or set one as the default. When adding, validate that the command exists on the system using `shutil.which`.
- `pn list` — print all saved projects as a tree, grouped by their group hierarchy, showing the full path under each project name.

## Interactive Project Browser (`pn open`)

Using `inquirer`, display a navigable list of all top-level groups and ungrouped projects. Groups show as `📁 Group Name` and projects show as `📄 Project Name`. Selecting a group drills into it, showing subgroups and projects within. Always include a `← Back` option to go up a level.

When a project is selected, display a final prompt asking how to open it with two options:

- `e` — open in editor and cd into the directory
- `Enter` — cd only

Since a Python script cannot change the parent shell's directory directly, implement the `cd` behavior by printing a `cd /path/to/project` command and instructing the user to add the following shell function to their `.bashrc` / `.zshrc`:

    pn() {
      result=$(command pn "$@")
      if echo "$result" | grep -q "^cd "; then
        eval "$result"
      else
        echo "$result"
      fi
    }

## Opening Behavior

- **`e` keystroke**: run `subprocess.Popen([editor, path])` using the project's editor override or the global default, then print `cd /path/to/project` for the shell function to evaluate.
- **`Enter` keystroke**: print `cd /path/to/project` only.

## `pn list` Tree Output

Use `rich.tree` to render something like this:

    📁 Work
    └── 📁 Backend
        └── 📄 API Service  /home/user/work/api
    📁 Personal
    └── 📄 Blog  /home/user/personal/blog
    📄 Dotfiles  /home/user/dotfiles

## `pn editors` Table Output

Display the editors list as a table using `rich.table` with columns for Name, Command, and a `✓` marker on the current default:

    Name        Command      Default
    VS Code     code         ✓
    Neovim      nvim
    WebStorm    webstorm

## Installation & Naming

The script file should be named `pn.py`. Include instructions at the top of the script (in a comment block) for how to make it available as the `pn` command globally:

    # Option 1: pip install (recommended)
    # Add a pyproject.toml that registers the entry point:
    #   [project.scripts]
    #   pn = "pn:main"
    # Then run: pip install -e .

    # Option 2: Manual
    # chmod +x pn.py
    # sudo cp pn.py /usr/local/bin/pn

Also generate a minimal `pyproject.toml` alongside the script that defines the package name as `project-navigator`, version `0.1.0`, and registers the `pn` entry point pointing to the `main()` function in `pn.py`. This way the user can run `pip install -e .` from the directory and have `pn` available everywhere as a proper CLI command.

## Error Handling

- Warn (but don't crash) if a saved project path no longer exists — mark it with `⚠️` in the list.
- Gracefully handle a missing config file by creating it with defaults on first run.
- If the configured default editor command is not found on the system at runtime, print a warning and prompt the user to run `pn editors` to fix it.

## Dependencies

The script must only require `inquirer`, `rich`, and Python standard library modules. Include a `requirements.txt` with exact packages and installation instructions in a comment block at the top of the script.
