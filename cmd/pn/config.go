package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const appName = "project-navigator"

// Editor is a named CLI editor command.
type Editor struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}

// Project is a saved directory. Editor is nil to use the global default.
type Project struct {
	Name   string  `json:"name"`
	Path   string  `json:"path"`
	Editor *string `json:"editor"`
}

// Group is a recursively nestable collection of subgroups and projects.
type Group struct {
	Name      string     `json:"name"`
	Subgroups []*Group   `json:"subgroups"`
	Projects  []*Project `json:"projects"`
}

// Config is the whole on-disk model. DefaultEditor is nil when unset.
type Config struct {
	DefaultEditor *string    `json:"default_editor"`
	Editors       []Editor   `json:"editors"`
	Groups        []*Group   `json:"groups"`
	Projects      []*Project `json:"projects"`
}

// container unifies the root config and any group: both expose child groups
// and projects, so traversal/mutation logic is written once.
type container interface {
	groups() *[]*Group
	projects() *[]*Project
}

func (c *Config) groups() *[]*Group     { return &c.Groups }
func (c *Config) projects() *[]*Project { return &c.Projects }
func (g *Group) groups() *[]*Group      { return &g.Subgroups }
func (g *Group) projects() *[]*Project  { return &g.Projects }

func newGroup(name string) *Group {
	return &Group{Name: name, Subgroups: []*Group{}, Projects: []*Project{}}
}

func defaultConfig() *Config {
	return &Config{Editors: []Editor{}, Groups: []*Group{}, Projects: []*Project{}}
}

// walkGroups visits every group beneath c, depth-first, with its nesting depth.
func walkGroups(c container, depth int, fn func(depth int, g *Group)) {
	for _, g := range *c.groups() {
		fn(depth, g)
		walkGroups(g, depth+1, fn)
	}
}

// removeGroup deletes target wherever it lives in the tree.
func removeGroup(c container, target *Group) bool {
	gs := c.groups()
	for i, g := range *gs {
		if g == target {
			*gs = append((*gs)[:i], (*gs)[i+1:]...)
			return true
		}
		if removeGroup(g, target) {
			return true
		}
	}
	return false
}

func resolveEditor(c *Config, p *Project) string {
	if p.Editor != nil && *p.Editor != "" {
		return *p.Editor
	}
	if c.DefaultEditor != nil {
		return *c.DefaultEditor
	}
	return ""
}

// ---------------------------------------------------------------------------
// Paths
// ---------------------------------------------------------------------------

func expandUser(p string) string {
	if strings.HasPrefix(p, "~") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, p[1:])
		}
	}
	return p
}

// configDir resolves the OS-appropriate config directory.
//
//  1. $PN_CONFIG_DIR       — explicit override
//  2. $XDG_CONFIG_HOME     — honoured on every platform if set
//  3. per-OS default       — macOS Application Support / Windows APPDATA / Linux ~/.config
func configDir() string {
	if v := os.Getenv("PN_CONFIG_DIR"); v != "" {
		return expandUser(v)
	}
	if v := os.Getenv("XDG_CONFIG_HOME"); v != "" {
		return filepath.Join(expandUser(v), appName)
	}
	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "windows":
		base := os.Getenv("APPDATA")
		if base == "" {
			base = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(base, appName)
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", appName)
	default:
		return filepath.Join(home, ".config", appName)
	}
}

func configFile() string { return filepath.Join(configDir(), "config.json") }

func legacyConfigFile() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".project-navigator", "config.json")
}

// ---------------------------------------------------------------------------
// Load / save
// ---------------------------------------------------------------------------

func loadConfig() (*Config, error) {
	migrateLegacyConfig()
	path := configFile()
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		c := defaultConfig()
		if err := saveConfig(c); err != nil {
			return nil, err
		}
		dimf("Created a new config at %s", path)
		return c, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read config at %s: %w", path, err)
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("could not parse config at %s: %w", path, err)
	}
	normalize(&c)
	return &c, nil
}

func saveConfig(c *Config) error {
	dir := configDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(filepath.Join(dir, "config.json"), data, 0o644)
}

// migrateLegacyConfig moves a pre-1.0 ~/.project-navigator/config.json over once.
func migrateLegacyConfig() {
	dst := configFile()
	src := legacyConfigFile()
	if _, err := os.Stat(dst); err == nil {
		return // new config already exists
	}
	if _, err := os.Stat(src); err != nil {
		return // nothing to migrate
	}
	if src == dst {
		return
	}
	if err := os.MkdirAll(configDir(), 0o755); err != nil {
		warnf("Could not migrate old config: %v", err)
		return
	}
	if err := os.Rename(src, dst); err != nil {
		warnf("Could not migrate old config: %v", err)
		return
	}
	os.Remove(filepath.Dir(src)) // remove now-empty legacy dir (ignored if non-empty)
	dimf("Moved config to %s", dst)
}

// normalize replaces nil slices with empty ones so the JSON round-trips as [].
func normalize(c *Config) {
	if c.Editors == nil {
		c.Editors = []Editor{}
	}
	if c.Groups == nil {
		c.Groups = []*Group{}
	}
	if c.Projects == nil {
		c.Projects = []*Project{}
	}
	var fix func(g *Group)
	fix = func(g *Group) {
		if g.Subgroups == nil {
			g.Subgroups = []*Group{}
		}
		if g.Projects == nil {
			g.Projects = []*Project{}
		}
		for _, sg := range g.Subgroups {
			fix(sg)
		}
	}
	for _, g := range c.Groups {
		fix(g)
	}
}
