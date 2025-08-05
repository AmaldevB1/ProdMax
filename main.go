package main

import (
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var jsonData interface{}
var rightContainer *fyne.Container
var statusLabel *widget.Label

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("JSON Editor")
	myWindow.Resize(fyne.NewSize(800, 600))

	// raw JSON entry
	entry := widget.NewMultiLineEntry()
	entry.SetPlaceHolder("Enter JSON here...")
	leftScroll := container.NewVScroll(entry)

	// status bar for errors
	statusLabel = widget.NewLabel("")

	// right pane placeholder
	rightContainer = container.NewMax(widget.NewLabel("JSON tree will appear here"))

	// split pane: left (editor) and right (tree)
	split := container.NewHSplit(leftScroll, rightContainer)
	split.Offset = 0.5

	// beautify & render button
	beautifyBtn := widget.NewButton("Beautify & Render", func() {
		raw := entry.Text
		var tmp interface{}
		err := json.Unmarshal([]byte(raw), &tmp)
		if err != nil {
			statusLabel.SetText(fmt.Sprintf("JSON Error: %v", err))
			return
		}
		statusLabel.SetText("")

		// prettify editor text
		b, _ := json.MarshalIndent(tmp, "", "  ")
		entry.SetText(string(b))
		jsonData = tmp
		refreshTree()
	})

	// layout: toolbar top, split center, status bottom
	content := container.NewBorder(
		container.NewHBox(beautifyBtn),
		statusLabel,
		nil, nil,
		split,
	)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}

// refreshTree generates and displays the JSON Tree
func refreshTree() {
	// maps for branch relations and leaf values
	dataMap := make(map[string][]string)
	valueMap := make(map[string]string)

	// recursive traversal to populate maps
	var traverse func(prefix string, v interface{})
	traverse = func(prefix string, v interface{}) {
		switch val := v.(type) {
		case map[string]interface{}:
			for k, child := range val {
				id := prefix + "/" + k
				dataMap[prefix] = append(dataMap[prefix], id)
				traverse(id, child)
			}
		case []interface{}:
			for i, child := range val {
				id := fmt.Sprintf("%s/[%d]", prefix, i)
				dataMap[prefix] = append(dataMap[prefix], id)
				traverse(id, child)
			}
		default:
			valueMap[prefix] = fmt.Sprintf("%v", val)
		}
	}

	// initialize root node
	root := "root"
	dataMap[""] = []string{root}
	traverse(root, jsonData)

	// build Tree widget
	tree := widget.NewTree(
		// child node IDs
		func(uid widget.TreeNodeID) []widget.TreeNodeID {
			return dataMap[string(uid)]
		},
		// isBranch
		func(uid widget.TreeNodeID) bool {
			_, ok := dataMap[string(uid)]
			return ok
		},
		// create widget for each node
		func(branch bool) fyne.CanvasObject {
			lbl := widget.NewLabel("")
			lbl.Wrapping = fyne.TextWrapOff
			return lbl
		},
		// update widget content
		func(uid widget.TreeNodeID, branch bool, obj fyne.CanvasObject) {
			lbl := obj.(*widget.Label)
			key := string(uid)
			if idx := lastIndex(key, '/'); idx >= 0 {
				key = key[idx+1:]
			}
			if branch {
				lbl.SetText(key)
			} else {
				lbl.SetText(fmt.Sprintf("%s: %s", key, valueMap[string(uid)]))
			}
		},
	)

	// wrap tree in scroll
	sc := container.NewVScroll(tree)
	sc.SetMinSize(fyne.NewSize(200, 0))

	// replace right pane content
	rightContainer.Objects = []fyne.CanvasObject{sc}
	rightContainer.Refresh()
}

// helper: find last index of sep in s
func lastIndex(s string, sep byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == sep {
			return i
		}
	}
	return -1
}
