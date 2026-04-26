Set Shell "bash"
Set FontSize 18
Set Width 1200
Set Height 720
Set Padding 16
Set Theme "Catppuccin Mocha"
Set TypingSpeed 110ms
Set PlaybackSpeed 1.0

Output "/tmp/ccw-demo.gif"

Env CCW_LANG en

Hide
Type "cd /tmp/ccw-demo && export HOME=/tmp/ccw-demo-home PATH=/tmp/ccw-demo-bin:/tmp/fake-gh:$PATH && clear"
Enter
Sleep 1000ms
Show

Type "ccw"
Enter
# initial picker view — rows render as 3 lines (🌲 name + status / branch / pr)
# top row hovered: 🌲 chore-cleanup · [DIRTY] + NEW + [CLOSED] #45
Sleep 7000ms

# walk through every worktree row to show the 4 status × 4 PR-state combos:
# 🌲 chore-cleanup [DIRTY] [CLOSED]  →  🌲 feat-dashboard [PUSHED] [MERGED]
Down
Sleep 3000ms
# 🌲 feat-dashboard [PUSHED] [MERGED]  →  🌲 feat-login [PUSHED] [OPEN]
Down
Sleep 3000ms
# 🌲 feat-login [PUSHED] [OPEN]  →  🌲 feat-picker [LOCAL] [DRAFT]
Down
Sleep 3000ms

# return to feat-dashboard ([PUSHED] + [MERGED] + RESUME) for the submenu demo
Up
Sleep 2500ms
Up
Sleep 2500ms

# submenu shows the full worktree path (replaces the path row that used to be in the list)
# [r] run / [d] delete / [b] back
Enter
Sleep 5000ms
Type "b"
Sleep 2500ms

# jump to [clean pushed] — past feat-login / feat-picker / [delete all]
Down 4
Sleep 2500ms
Enter
# bulk-confirm screen — preview of pushed targets
Sleep 5500ms

# cancel
Type "N"
Sleep 2500ms

# quit picker
Type "q"
Sleep 1500ms
