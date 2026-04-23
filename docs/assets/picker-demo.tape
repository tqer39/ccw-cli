Output "/tmp/ccw-demo.gif"

Set Shell "bash"
Set FontSize 20
Set Width 1400
Set Height 900
Set Padding 24
Set Theme "Catppuccin Mocha"
Set TypingSpeed 90ms

Hide
Type "cd /tmp/ccw-demo && export PATH=/tmp/fake-gh:$PATH && clear"
Enter
Sleep 800ms
Show

Type "/tmp/ccw-demo-bin"
Enter
Sleep 5000ms

# show badges by moving selection across all 3 worktrees
Down
Sleep 1800ms
Down
Sleep 1800ms
Up
Sleep 1400ms

# enter submenu on the currently-selected worktree
Enter
Sleep 3600ms

# return to picker list
Type "b"
Sleep 1800ms

# navigate to [clean pushed] to trigger bulk-confirm
Down 3
Sleep 1600ms
Enter
Sleep 5000ms

# cancel bulk delete
Type "N"
Sleep 2400ms

# quit picker
Type "q"
Sleep 1400ms
