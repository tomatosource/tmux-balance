package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type Layout struct {
	isRow      bool
	dimensions Dimensions
	children   []*Layout
	parent     *Layout
	pane       *Pane
}

type Pane struct {
	id         int
	dimensions Dimensions
}

type Dimensions struct {
	height int
	width  int
}

func main() {
	arg := os.Args[1]
	switch arg {
	case "x":
		kill()
	case "v":
		split(true)
	case "s":
		split(false)
	}
}

func split(horizontal bool) {
	if horizontal {
		mustExec("tmux", "split-window", "-h", `-c "$PWD"`)
	} else {
		mustExec("tmux", "split-window", "-v", `-c "$PWD"`)
	}
	activePane := getActivePane()
	rootLayout := getRootLayout()
	layout := getLayoutByPaneID(activePane, rootLayout)
	rebalanceLayout(layout, horizontal)
}

func kill() {
	activePane := getActivePane()
	rootLayout := getRootLayout()
	layout := getLayoutByPaneID(activePane, rootLayout)
	if layout.pane != nil {
		return
	}

	mustExec("tmux", "kill-pane")

	baseSize := layout.dimensions.height
	if layout.isRow {
		baseSize = layout.dimensions.width
	}
	newSizes := getNewSizes(len(layout.children)-1, baseSize)

	var offset int
	for i, child := range layout.children {
		if child.pane != nil && child.pane.id == activePane {
			offset = 1
			continue
		}
		if layout.isRow {
			setLayoutSize(child, newSizes[i-offset], child.dimensions.height)
		} else {
			setLayoutSize(child, child.dimensions.width, newSizes[i-offset])
		}
	}
}

func rebalanceLayout(layout *Layout, horizontal bool) {
	if layout.pane != nil {
		return
	}

	baseSize := layout.dimensions.height
	if horizontal {
		baseSize = layout.dimensions.width
	}
	newSizes := getNewSizes(len(layout.children), baseSize)

	for i, child := range layout.children {
		if horizontal {
			setLayoutSize(child, newSizes[i], child.dimensions.height)
		} else {
			setLayoutSize(child, child.dimensions.width, newSizes[i])
		}
	}
}

func setLayoutSize(layout *Layout, width, height int) {
	if layout.pane != nil {
		setPaneSize(layout.pane.id, width, height)
		return
	}

	baseSize := height
	if layout.isRow {
		baseSize = width
	}
	newSizes := getNewSizes(len(layout.children), baseSize)

	for i, child := range layout.children {
		if layout.isRow {
			setLayoutSize(child, newSizes[i], height)
		} else {
			setLayoutSize(child, width, newSizes[i])
		}
	}
}

func getLayoutByPaneID(paneID int, layout *Layout) *Layout {
	if layout.pane != nil && layout.pane.id == paneID {
		return layout.parent
	}
	for _, child := range layout.children {
		if l := getLayoutByPaneID(paneID, child); l != nil {
			return l
		}
	}
	return nil
}

func getRootLayout() *Layout {
	rawLayout := mustExec(
		"tmux", "list-windows", "-F", "'#{window_layout}'",
	)
	initComma := strings.Index(rawLayout, ",")
	rawLayout = rawLayout[initComma+1 : len(rawLayout)-2]
	layouts, _ := getLayout(rawLayout, nil, 0)
	return layouts[0]
}

func getLayout(layout string, parent *Layout, paneID int) ([]*Layout, int) {
	layout = strings.TrimSpace(layout)
	layouts := []*Layout{}

	for layout != "" {
		layout = strings.TrimPrefix(layout, ",")
		dimRegex, _ := regexp.Compile(`\d+x\d+`)
		rawDims := dimRegex.FindString(layout)
		dims := parseDimensions(rawDims)
		layout = string(layout[len(rawDims):])
		xyRegex, _ := regexp.Compile(`^,\d+,\d+`)
		xy := xyRegex.FindString(layout)
		layout = string(layout[len(xy):])

		if strings.HasPrefix(layout, ",") {
			layouts = append(layouts, &Layout{
				parent:     parent,
				dimensions: dims,
				pane: &Pane{
					dimensions: dims,
					id:         paneID,
				},
			})
			paneID++
			idRegex, _ := regexp.Compile(`^,\d+(,|$)`)
			id := idRegex.FindString(layout)
			layout = string(layout[len(id):])
		} else if strings.HasPrefix(layout, "{") {
			matchingIndex := getMatchIndex(layout, '{', '}')
			l := &Layout{
				isRow:      true,
				parent:     parent,
				dimensions: dims,
			}
			l.children, paneID = getLayout(layout[1:matchingIndex], l, paneID)
			layouts = append(layouts, l)
			layout = string(layout[matchingIndex+1:])
		} else if strings.HasPrefix(layout, "[") {
			matchingIndex := getMatchIndex(layout, '[', ']')
			l := &Layout{
				parent:     parent,
				dimensions: dims,
			}
			l.children, paneID = getLayout(layout[1:matchingIndex], l, paneID)
			layouts = append(layouts, l)
			layout = string(layout[matchingIndex+1:])
		}
	}
	return layouts, paneID
}

func parseDimensions(dims string) Dimensions {
	parts := strings.Split(dims, "x")
	w, _ := strconv.Atoi(parts[0])
	h, _ := strconv.Atoi(parts[1])

	return Dimensions{
		width:  w,
		height: h,
	}
}

func getMatchIndex(s string, opener, closer rune) int {
	x := 0
	for i, c := range s {
		if c == opener {
			x++
		} else if c == closer {
			x--
		}
		if x == 0 {
			return i
		}
	}
	return -1
}

func getNewSizes(count, total int) []int {
	newBaseSize := int(total / count)
	rem := total % count
	newSizes := []int{}
	for i := 0; i < count; i++ {
		newSize := newBaseSize
		if rem > 0 {
			newSize += 1
			rem -= 1
		}
		newSizes = append(newSizes, newSize)
	}
	return newSizes
}

func getActivePane() int {
	lines := strings.Split(mustExec("tmux", "list-panes"), "\n")
	for _, line := range lines {
		if strings.Contains(line, "active") {
			colonIndex := strings.Index(line, ":")
			i, _ := strconv.Atoi(line[:colonIndex])
			return i
		}
	}
	return 0
}

func mustExec(cmd string, args ...string) string {
	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		return "ERROR: " + string(out)
	}
	return string(out)
}

func setPaneSize(paneID, width, height int) {
	mustExec(
		"tmux",
		"resize-pane",
		fmt.Sprintf("-t %d", paneID),
		fmt.Sprintf("-x %d", width),
		fmt.Sprintf("-y %d", height),
	)
}
