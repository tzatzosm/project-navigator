package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func cmdAdd(c *Config, pathArg string) error {
	raw := pathArg
	if raw == "" {
		raw, _ = os.Getwd()
	}
	path, err := filepath.Abs(expandUser(raw))
	if err != nil {
		return err
	}
	if !isDir(path) {
		warnf("⚠️  Not an existing directory: %s", path)
		ok, err := promptConfirm("Add it anyway?", false)
		if err != nil {
			return err
		}
		if !ok {
			dimf("Cancelled.")
			return nil
		}
	}

	defName := filepath.Base(path)
	name, err := promptText("Project name", defName)
	if err != nil {
		return err
	}
	if name == "" {
		name = defName
	}

	target, create, err := chooseGroup(c, "Assign to group", true, true)
	if err != nil {
		return err
	}
	if create {
		gname, err := promptText("New group name", "")
		if err != nil {
			return err
		}
		if gname == "" {
			dimf("Cancelled.")
			return nil
		}
		g := newGroup(gname)
		*c.groups() = append(*c.groups(), g) // new groups land at the top level
		target = g
	}

	var editor *string
	if len(c.Editors) > 0 {
		editor, err = chooseEditor(c, "Editor for this project")
		if err != nil {
			return err
		}
	}

	*target.projects() = append(*target.projects(), &Project{Name: name, Path: path, Editor: editor})
	if err := saveConfig(c); err != nil {
		return err
	}

	where := "top level"
	if g, ok := target.(*Group); ok {
		where = "📁 " + g.Name
	}
	okf("✓ Added %s → %s", name, where)
	return nil
}

// choice is one row in the interactive browser.
type choice struct {
	kind  string // group | project | back | quit
	group *Group
	proj  *Project
}

func cmdOpen(c *Config) error {
	var node container = c
	var nodeStack []container
	var nameStack []string

	for {
		groups := *node.groups()
		projects := *node.projects()

		var labels []string
		var rows []choice
		for _, g := range groups {
			labels = append(labels, "📁 "+g.Name)
			rows = append(rows, choice{kind: "group", group: g})
		}
		for _, p := range projects {
			label := "📄 " + p.Name
			if !isDir(p.Path) {
				label += "  ⚠️"
			}
			labels = append(labels, label)
			rows = append(rows, choice{kind: "project", proj: p})
		}
		if len(nodeStack) > 0 {
			labels = append(labels, "← Back")
			rows = append(rows, choice{kind: "back"})
		} else {
			labels = append(labels, "✖ Quit")
			rows = append(rows, choice{kind: "quit"})
		}

		if len(groups) == 0 && len(projects) == 0 {
			dimf("(this level is empty)")
		}

		breadcrumb := strings.Join(append([]string{"~"}, nameStack...), " / ")
		idx, err := chooseIndex(breadcrumb, labels)
		if err != nil {
			return err
		}

		switch row := rows[idx]; row.kind {
		case "group":
			nodeStack = append(nodeStack, node)
			nameStack = append(nameStack, row.group.Name)
			node = row.group
		case "back":
			node = nodeStack[len(nodeStack)-1]
			nodeStack = nodeStack[:len(nodeStack)-1]
			nameStack = nameStack[:len(nameStack)-1]
		case "project":
			return openProject(c, row.proj)
		case "quit":
			return nil
		}
	}
}

func openProject(c *Config, p *Project) error {
	if !isDir(p.Path) {
		warnf("⚠️  Path no longer exists: %s", p.Path)
		ok, err := promptConfirm("cd there anyway?", false)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
	}
	idx, err := chooseIndex("Open "+p.Name, []string{"Open in editor and cd", "cd only"})
	if err != nil {
		return err
	}
	if idx == 0 {
		launchEditor(c, p)
	}
	emitCD(p.Path)
	return nil
}

func cmdGroups(c *Config) error {
	idx, err := chooseIndex("Manage groups", []string{"Add a group", "Rename a group", "Delete a group"})
	if err != nil {
		return err
	}

	switch idx {
	case 0:
		name, err := promptText("New group name", "")
		if err != nil {
			return err
		}
		if name == "" {
			dimf("Cancelled.")
			return nil
		}
		parent, _, err := chooseGroup(c, "Create it under", true, false)
		if err != nil {
			return err
		}
		*parent.groups() = append(*parent.groups(), newGroup(name))
		okf("✓ Created group %s", name)

	case 1:
		g, err := chooseExistingGroup(c, "Rename which group")
		if err != nil {
			return err
		}
		if g == nil {
			warnf("No groups yet.")
			return nil
		}
		name, err := promptText("New name", g.Name)
		if err != nil {
			return err
		}
		if name != "" {
			g.Name = name
			okf("✓ Renamed")
		}

	case 2:
		g, err := chooseExistingGroup(c, "Delete which group")
		if err != nil {
			return err
		}
		if g == nil {
			warnf("No groups yet.")
			return nil
		}
		ok, err := promptConfirm(fmt.Sprintf("Delete '%s' and everything inside it?", g.Name), false)
		if err != nil {
			return err
		}
		if ok {
			removeGroup(c, g)
			okf("✓ Deleted")
		}
	}

	return saveConfig(c)
}

func cmdConfig(c *Config) error {
	if len(c.Editors) == 0 {
		warnf("No editors configured yet — let's add one.")
		return cmdEditors(c)
	}
	var labels []string
	for _, e := range c.Editors {
		marker := ""
		if c.DefaultEditor != nil && e.Command == *c.DefaultEditor {
			marker = "  ✓"
		}
		labels = append(labels, fmt.Sprintf("%s  (%s)%s", e.Name, e.Command, marker))
	}
	idx, err := chooseIndex("Default editor", labels)
	if err != nil {
		return err
	}
	cmd := c.Editors[idx].Command
	c.DefaultEditor = &cmd
	if err := saveConfig(c); err != nil {
		return err
	}
	okf("✓ Default editor set to %s", cmd)
	return nil
}

func cmdEditors(c *Config) error {
	renderEditorsTable(c)
	idx, err := chooseIndex("Manage editors", []string{"Add an editor", "Remove an editor", "Set the default"})
	if err != nil {
		return err
	}

	switch idx {
	case 0:
		name, err := promptText("Display name", "")
		if err != nil {
			return err
		}
		if name == "" {
			dimf("Cancelled.")
			return nil
		}
		command, err := promptText("CLI command", "")
		if err != nil {
			return err
		}
		if command == "" {
			dimf("Cancelled.")
			return nil
		}
		if !commandExists(command) {
			warnf("⚠️  '%s' was not found on your PATH.", command)
			ok, err := promptConfirm("Add it anyway?", false)
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
		}
		c.Editors = append(c.Editors, Editor{Name: name, Command: command})
		if c.DefaultEditor == nil {
			cmd := command // first editor becomes the default
			c.DefaultEditor = &cmd
		}
		okf("✓ Added editor %s", name)

	case 1:
		if len(c.Editors) == 0 {
			warnf("No editors to remove.")
			return nil
		}
		var labels []string
		for _, e := range c.Editors {
			labels = append(labels, e.Name)
		}
		i, err := chooseIndex("Remove which editor", labels)
		if err != nil {
			return err
		}
		removed := c.Editors[i]
		c.Editors = append(c.Editors[:i], c.Editors[i+1:]...)
		if c.DefaultEditor != nil && *c.DefaultEditor == removed.Command {
			if len(c.Editors) > 0 {
				cmd := c.Editors[0].Command
				c.DefaultEditor = &cmd
			} else {
				c.DefaultEditor = nil
			}
		}
		okf("✓ Removed %s", removed.Name)

	case 2:
		if len(c.Editors) == 0 {
			warnf("No editors yet.")
			return nil
		}
		var labels []string
		for _, e := range c.Editors {
			labels = append(labels, e.Name)
		}
		i, err := chooseIndex("Set the default editor", labels)
		if err != nil {
			return err
		}
		cmd := c.Editors[i].Command
		c.DefaultEditor = &cmd
		okf("✓ Default editor set")
	}

	return saveConfig(c)
}

func cmdList(c *Config) error {
	if len(c.Groups) == 0 && len(c.Projects) == 0 {
		dimf("No projects yet. Add one with `pn add`.")
		return nil
	}
	var lines []string
	for _, g := range c.Groups {
		lines = append(lines, groupTree(g).String())
	}
	for _, p := range c.Projects {
		lines = append(lines, projectLabel(p))
	}
	fmt.Fprintln(os.Stderr, strings.Join(lines, "\n"))
	return nil
}
