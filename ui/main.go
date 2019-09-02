// +build ignore

package main

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	humanize "github.com/dustin/go-humanize"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type nodeValue string

func (nv nodeValue) String() string {
	return string(nv)
}

func makeNodes(r io.Reader) []*widgets.TreeNode {
	nodes := []*widgets.TreeNode{}

	zr, err := gzip.NewReader(r)
	if err != nil {
		// log.Fatal(err)
		return nodes
	}
	defer zr.Close()

	tr := tar.NewReader(zr)
	node := &widgets.TreeNode{}

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(hdr.Name)
		if hdr.FileInfo().IsDir() {
			node.Value = nodeValue(hdr.Name)
		} else {
			if strings.EqualFold(filepath.Ext(hdr.Name), ".tar") {
				for _, n := range makeNodes(tr) {
					nodes = append(nodes, n)
				}
			}
			node.Nodes = append(node.Nodes, &widgets.TreeNode{
				Value: nodeValue(fmt.Sprintf("%s (%s)", hdr.Name, humanize.Bytes(uint64(hdr.Size)))),
				Nodes: nil,
			})
		}
		// if _, err := io.Copy(os.Stdout, tr); err != nil {
		// 	log.Fatal(err)
		// }
	}

	nodes = append(nodes, node)

	return nodes
}

func main() {
	if len(os.Args) == 0 {
		log.Fatal("you must suppy a docker image tar to extract from")
	}
	if _, err := os.Stat(os.Args[1]); os.IsNotExist(err) {
		log.Fatalf("file does not exist: %s", os.Args[1])
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer f.Close()

	nodes := makeNodes(bufio.NewReader(f))
	// return

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	l := widgets.NewTree()
	l.TextStyle = ui.NewStyle(ui.ColorYellow)
	l.WrapText = false
	l.SetNodes(nodes)

	x, y := ui.TerminalDimensions()

	l.SetRect(0, 0, x, y)

	ui.Render(l)

	previousKey := ""
	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			return
		case "j", "<Down>":
			l.ScrollDown()
		case "k", "<Up>":
			l.ScrollUp()
		case "<C-d>":
			l.ScrollHalfPageDown()
		case "<C-u>":
			l.ScrollHalfPageUp()
		case "<C-f>":
			l.ScrollPageDown()
		case "<C-b>":
			l.ScrollPageUp()
		case "g":
			if previousKey == "g" {
				l.ScrollTop()
			}
		case "<Home>":
			l.ScrollTop()
		case "<Space>":
			dateCmd := exec.Command("say", "-v", "Fiona", l.SelectedNode().Value.String())
			go dateCmd.Output()
			if l.SelectedNode().Nodes != nil {
				for _, node := range l.SelectedNode().Nodes {
					dateCmd := exec.Command("say", "-v", "Fiona", node.Value.String())
					dateCmd.Output()
				}
			}
			// return
		case "<Enter>":
			l.ToggleExpand()
		case "G", "<End>":
			l.ScrollBottom()
		case "E":
			l.ExpandAll()
		case "C":
			l.CollapseAll()
		case "<Resize>":
			x, y := ui.TerminalDimensions()
			l.SetRect(0, 0, x, y)
		}

		if previousKey == "g" {
			previousKey = ""
		} else {
			previousKey = e.ID
		}

		ui.Render(l)
	}
}
