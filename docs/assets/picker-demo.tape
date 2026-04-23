Output "/tmp/ccw-demo.gif"

Set Shell "bash"
Set FontSize 14
Set Width 1100
Set Height 700
Set Padding 20
Set Theme "Catppuccin Mocha"
Set TypingSpeed 40ms

Hide
Type "cd /tmp/ccw-demo && export PATH=/tmp/fake-gh:$PATH && clear"
Enter
Sleep 400ms
Show

Type "/tmp/ccw-demo-bin"
Enter
Sleep 2500ms

# show badges by moving selection across all 3 worktrees
Down
Sleep 900ms
Down
Sleep 900ms
Up
Sleep 700ms

# enter submenu on the currently-selected worktree
Enter
Sleep 1800ms

# return to picker list
Type "b"
Sleep 900ms

# navigate to [clean pushed] to trigger bulk-confirm
Down 3
Sleep 800ms
Enter
Sleep 2500ms

# cancel bulk delete
Type "N"
Sleep 1200ms

# quit picker
Type "q"
Sleep 700ms
