# pn — project-navigator demo tape.
#
# Regenerate the GIF from the repo root (after `go build -o pn ./cmd/pn`):
#   vhs demo/demo.tape

Output demo/demo.gif

Set Shell "bash"
Set FontSize 18
Set Width 1200
Set Height 700
Set Padding 20
Set Theme "Catppuccin Mocha"
Set TypingSpeed 70ms

# --- hidden setup: put ./pn on PATH and seed sample projects ---
Hide
Type "export PATH=$PWD:$PATH" Enter
Type "source demo/seed.sh" Enter
Type "clear" Enter
Show

# List saved projects as a tree
Type "pn list"
Sleep 700ms
Enter
Sleep 2.5s

# Open the interactive browser
Type "pn"
Sleep 700ms
Enter
Sleep 1.5s

# Browse: glance down the list, then drill into "Work"
Down
Sleep 700ms
Down
Sleep 700ms
Up
Sleep 500ms
Up
Sleep 800ms
Enter
Sleep 1.5s

# Select "API Service"
Sleep 600ms
Enter
Sleep 1.5s

# Choose "cd only"
Down
Sleep 800ms
Enter
Sleep 1.2s

# We've changed into the project directory
Type "pwd"
Sleep 500ms
Enter
Sleep 2.5s
