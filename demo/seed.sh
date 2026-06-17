# Demo setup for the VHS recording (demo/demo.tape).
# Sourced from the tape's hidden section — sets up a throwaway config with a
# few sample projects and defines the `pn` cd-wrapper.

ROOT=/tmp/pn-demo
rm -rf "$ROOT"
mkdir -p "$ROOT/work/api" "$ROOT/work/web" "$ROOT/personal/blog" "$ROOT/dotfiles"

export PN_CONFIG_DIR="$ROOT/.config"
mkdir -p "$PN_CONFIG_DIR"
cat >"$PN_CONFIG_DIR/config.json" <<EOF
{
  "default_editor": null,
  "editors": [],
  "groups": [
    {"name": "Work", "subgroups": [], "projects": [
      {"name": "API Service", "path": "$ROOT/work/api", "editor": null},
      {"name": "Web App", "path": "$ROOT/work/web", "editor": null}
    ]},
    {"name": "Personal", "subgroups": [], "projects": [
      {"name": "Blog", "path": "$ROOT/personal/blog", "editor": null}
    ]}
  ],
  "projects": [
    {"name": "Dotfiles", "path": "$ROOT/dotfiles", "editor": null}
  ]
}
EOF

# The cd-wrapper, so `pn` actually changes directory during the demo.
pn() {
  r=$(command pn "$@")
  if echo "$r" | grep -q "^cd "; then eval "$r"; else echo "$r"; fi
}
