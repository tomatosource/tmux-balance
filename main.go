package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	fmt.Println(mustExec("tmux", "list-windows", "-F", "'#{window_active} #{window_layout}'"))
	fmt.Println(os.Getenv("TMUX_PANE"))
}

func mustExec(cmd string, args ...string) string {
	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		return "ERROR: " + string(out)
	}
	return string(out)
}
