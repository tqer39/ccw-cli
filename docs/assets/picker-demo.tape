Output "/tmp/ccw-demo.gif"

Set Shell "bash"
Set FontSize 28
Set Width 1280
Set Height 780
Set Padding 28
Set Theme "Catppuccin Mocha"
Set TypingSpeed 110ms
Set PlaybackSpeed 0.5

Hide
Type "cd /tmp/ccw-demo && export PATH=/tmp/fake-gh:$PATH && clear"
Enter
Sleep 800ms
Show

Type "/tmp/ccw-demo-bin"
Enter
Sleep 4500ms

# show badges by moving selection across all 3 worktrees
Down
Sleep 1600ms
Down
Sleep 1600ms
Up
Sleep 1200ms

# enter submenu on the currently-selected worktree
Enter
Sleep 3200ms

# return to picker list
Type "b"
Sleep 1600ms

# navigate to [clean pushed] to trigger bulk-confirm
Down 3
Sleep 1400ms
Enter
Sleep 4500ms

# cancel bulk delete
Type "N"
Sleep 2000ms

# quit picker
Type "q"
Sleep 1200ms
