Output "/tmp/ccw-demo.gif"

Set Shell "bash"
Set FontSize 24
Set Width 1440
Set Height 640
Set Padding 20
Set Theme "Catppuccin Mocha"
Set TypingSpeed 110ms
Set PlaybackSpeed 1.0

Hide
Type "cd /tmp/ccw-demo && export PATH=/tmp/ccw-demo-bin:/tmp/fake-gh:$PATH && clear"
Enter
Sleep 1000ms
Show

Type "ccw"
Enter
# initial picker view — PUSHED + [OPEN] highlighted
Sleep 7000ms

# walk through every worktree badge combination:
# PUSHED + [OPEN]  →  PUSHED + [MERGED]
Down
Sleep 3500ms
# PUSHED + [MERGED]  →  LOCAL + [DRAFT]
Down
Sleep 3500ms
# LOCAL + [DRAFT]  →  DIRTY + [CLOSED]
Down
Sleep 4000ms

# return to highlight PUSHED + [MERGED] for the submenu demo
Up
Sleep 2500ms
Up
Sleep 2500ms

# show [r] run / [d] delete / [b] back
Enter
Sleep 5000ms
Type "b"
Sleep 2500ms

# jump to [clean pushed] — from wt2 pass wt3/wt4 and [delete all]
Down 4
Sleep 2500ms
Enter
# bulk-confirm screen — targets preview
Sleep 5500ms

# cancel
Type "N"
Sleep 2500ms

# quit picker
Type "q"
Sleep 1500ms
