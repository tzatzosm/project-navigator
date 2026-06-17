// Command pn is a CLI for managing and navigating development projects.
//
// A process cannot change its parent shell's directory, so `pn` prints a
// `cd <path>` line on stdout that a small shell wrapper evaluates:
//
//	pn() {
//	  result=$(command pn "$@")
//	  if echo "$result" | grep -q "^cd "; then
//	    eval "$result"
//	  else
//	    echo "$result"
//	  fi
//	}
//
// Every prompt and all styled output is written to stderr; the ONLY thing sent
// to stdout is that `cd <path>` command.
package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

func main() {
	// Point lipgloss colour detection at stderr (where our output goes), so
	// colours survive even when stdout is captured by the shell wrapper.
	lipgloss.SetDefaultRenderer(lipgloss.NewRenderer(os.Stderr))

	args := os.Args[1:]
	command := "open" // bare `pn` opens the browser
	var rest []string
	if len(args) > 0 {
		command, rest = args[0], args[1:]
	}

	switch command {
	case "-h", "--help", "help":
		usage()
		return
	}

	c, err := loadConfig()
	if err != nil {
		errf("%v", err)
		os.Exit(1)
	}

	switch command {
	case "add":
		path := ""
		if len(rest) > 0 {
			path = rest[0]
		}
		err = cmdAdd(c, path)
	case "open":
		err = cmdOpen(c)
	case "groups":
		err = cmdGroups(c)
	case "config":
		err = cmdConfig(c)
	case "editors":
		err = cmdEditors(c)
	case "list":
		err = cmdList(c)
	default:
		errf("unknown command: %s", command)
		usage()
		os.Exit(2)
	}

	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			dimf("Cancelled.")
			os.Exit(130)
		}
		errf("%v", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Print(`usage: pn [command] [args]

Navigate and open your dev projects.

Commands:
  add [path]   Add the current dir (or path) as a project
  open         Open the interactive project browser (default)
  groups       Add, rename, or delete groups
  config       Set the global default editor
  editors      Manage the editors list
  list         Print all saved projects as a tree
`)
}

// emitCD writes the only line that ever goes to stdout — captured and eval'd by
// the shell wrapper.
func emitCD(path string) {
	fmt.Fprintf(os.Stdout, "cd %s\n", path)
}

func isDir(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func launchEditor(c *Config, p *Project) {
	editor := resolveEditor(c, p)
	if editor == "" {
		warnf("No editor configured. Run `pn editors` to add one.")
		return
	}
	if !commandExists(editor) {
		warnf("⚠️  Editor command not found: %s", editor)
		dimf("Run `pn editors` to fix your editor configuration.")
		return
	}
	cmd := exec.Command(editor, p.Path)
	// stdout/stderr stay nil → connected to the null device, so a chatty editor
	// can never pollute the `cd` channel.
	if err := cmd.Start(); err != nil {
		errf("Failed to launch %s: %v", editor, err)
		return
	}
	// Release the child so it isn't tied to this short-lived process.
	_ = cmd.Process.Release()
	okf("Opening in %s…", editor)
}
