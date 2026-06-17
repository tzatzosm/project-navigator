package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/charmbracelet/lipgloss/tree"
)

// Styles. The default renderer is pointed at stderr in main() so colours show
// even when stdout is captured by the shell wrapper.
var (
	styleBold = lipgloss.NewStyle().Bold(true)
	styleDim  = lipgloss.NewStyle().Faint(true)
	stylePath = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	styleGood = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	styleWarn = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	styleErr  = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

// All human-facing output goes to stderr; only the `cd` line goes to stdout.
func okf(format string, a ...any) {
	fmt.Fprintln(os.Stderr, styleGood.Render(fmt.Sprintf(format, a...)))
}
func warnf(format string, a ...any) {
	fmt.Fprintln(os.Stderr, styleWarn.Render(fmt.Sprintf(format, a...)))
}
func errf(format string, a ...any) {
	fmt.Fprintln(os.Stderr, styleErr.Render(fmt.Sprintf(format, a...)))
}
func dimf(format string, a ...any) {
	fmt.Fprintln(os.Stderr, styleDim.Render(fmt.Sprintf(format, a...)))
}

// runForm renders a single huh field to stderr and reads from stdin, so prompts
// stay off the stdout channel the `pn()` shell wrapper captures.
func runForm(field huh.Field) error {
	return huh.NewForm(huh.NewGroup(field)).
		WithOutput(os.Stderr).
		WithInput(os.Stdin).
		Run()
}

// chooseIndex shows a single-select list and returns the chosen index.
func chooseIndex(title string, labels []string) (int, error) {
	var idx int
	opts := make([]huh.Option[int], len(labels))
	for i, l := range labels {
		opts[i] = huh.NewOption(l, i)
	}
	err := runForm(huh.NewSelect[int]().Title(title).Options(opts...).Value(&idx))
	return idx, err
}

func promptText(title, def string) (string, error) {
	val := def
	if err := runForm(huh.NewInput().Title(title).Value(&val)); err != nil {
		return "", err
	}
	return strings.TrimSpace(val), nil
}

func promptConfirm(title string, def bool) (bool, error) {
	val := def
	err := runForm(huh.NewConfirm().Title(title).Affirmative("Yes").Negative("No").Value(&val))
	return val, err
}

// chooseGroup lets the user pick a destination container across the whole group
// hierarchy. With includeCreate, the returned bool signals "create a new group".
func chooseGroup(c *Config, title string, includeRoot, includeCreate bool) (target container, create bool, err error) {
	var labels []string
	var nodes []container
	if includeRoot {
		labels = append(labels, "(top level / no group)")
		nodes = append(nodes, c)
	}
	walkGroups(c, 0, func(depth int, g *Group) {
		labels = append(labels, strings.Repeat("  ", depth)+"📁 "+g.Name)
		nodes = append(nodes, g)
	})
	createIdx := -1
	if includeCreate {
		createIdx = len(labels)
		labels = append(labels, "➕ Create a new group")
	}
	idx, err := chooseIndex(title, labels)
	if err != nil {
		return nil, false, err
	}
	if idx == createIdx {
		return nil, true, nil
	}
	return nodes[idx], false, nil
}

// chooseExistingGroup selects a group node (no root, no create). Returns nil if
// there are no groups.
func chooseExistingGroup(c *Config, title string) (*Group, error) {
	var labels []string
	var groups []*Group
	walkGroups(c, 0, func(depth int, g *Group) {
		labels = append(labels, strings.Repeat("  ", depth)+"📁 "+g.Name)
		groups = append(groups, g)
	})
	if len(groups) == 0 {
		return nil, nil
	}
	idx, err := chooseIndex(title, labels)
	if err != nil {
		return nil, err
	}
	return groups[idx], nil
}

// chooseEditor returns the chosen editor command, or nil for "use default".
func chooseEditor(c *Config, title string) (*string, error) {
	labels := []string{"Use default"}
	for _, e := range c.Editors {
		labels = append(labels, fmt.Sprintf("%s  (%s)", e.Name, e.Command))
	}
	idx, err := chooseIndex(title, labels)
	if err != nil {
		return nil, err
	}
	if idx == 0 {
		return nil, nil
	}
	cmd := c.Editors[idx-1].Command
	return &cmd, nil
}

func renderEditorsTable(c *Config) {
	if len(c.Editors) == 0 {
		dimf("No editors configured.")
		return
	}
	t := table.New().
		Border(lipgloss.NormalBorder()).
		Headers("Name", "Command", "Default")
	for _, e := range c.Editors {
		mark := ""
		if c.DefaultEditor != nil && e.Command == *c.DefaultEditor {
			mark = "✓"
		}
		t.Row(e.Name, e.Command, mark)
	}
	fmt.Fprintln(os.Stderr, t.Render())
}

func projectLabel(p *Project) string {
	label := "📄 " + p.Name + "  " + stylePath.Render(p.Path)
	if !isDir(p.Path) {
		label += " " + styleWarn.Render("⚠️")
	}
	return label
}

func groupTree(g *Group) *tree.Tree {
	t := tree.Root("📁 " + styleBold.Render(g.Name))
	for _, sg := range g.Subgroups {
		t.Child(groupTree(sg))
	}
	for _, p := range g.Projects {
		t.Child(projectLabel(p))
	}
	return t
}
