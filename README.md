# tmux-balance

Aims to change split behaviour to match vim/screen/iterm2.

## Installing

`go get github.com/tomatosource/tmux-balance`

## Usage

Add the following to your `.tmux.conf`

```
bind v run-shell "tmux-balance v &> /dev/null" # split horizontal
bind s run-shell "tmux-balance s &> /dev/null" # split vertical
bind x run-shell "tmux-balance x &> /dev/null" # kill pane/window
```

Readers exercise to rebind ctrl+t/(shift)ctrl tab for new window, next/prev window.

## asciinema

[![demo](https://asciinema.org/a/D0rixCa3pKa6R9ZkXjAphB1cP.svg)](https://asciinema.org/a/D0rixCa3pKa6R9ZkXjAphB1cP?autoplay=1)
