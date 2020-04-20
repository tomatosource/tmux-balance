# tmux-balance

Aims to change split behaviour to match vim or iterm.

## Installing

`go get github.com/tomatosource/tmux-balance`

## Usage

Add the following to your `.tmux.conf`

```
bind v run-shell "tmux-balance v &> /dev/null" # split horizontal
bind s run-shell "tmux-balance s &> /dev/null" # split vertical 
bind x run-shell "tmux-balance x &> /dev/null" # kill pane
```

## TODO

Closing the middle split on the following doesn't balance properly yet:

```
| | |_|
| | | |
```
