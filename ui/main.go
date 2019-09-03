package main

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"github.com/blacktop/graboid/pkg/image"
	humanize "github.com/dustin/go-humanize"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type nodeValue string

func (nv nodeValue) String() string {
	return string(nv)
}

func parseManifest(r io.Reader) (*image.Manifest, error) {

	manifests := []image.Manifest{}

	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	tr := tar.NewReader(zr)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return nil, err
		}

		if strings.EqualFold(hdr.Name, "manifest.json") {
			rawJSON, err := ioutil.ReadAll(tr)
			if err != nil {
				return nil, err
			}
			if err := json.Unmarshal(rawJSON, &manifests); err != nil {
				return nil, err
			}
			return &manifests[0], nil
		}
	}
	return nil, nil
}

func parseImage(r io.Reader, m *image.Manifest) (*image.Image, error) {

	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	tr := tar.NewReader(zr)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return nil, err
		}

		if strings.EqualFold(hdr.Name, m.Config) {
			rawJSON, err := ioutil.ReadAll(tr)
			if err != nil {
				return nil, err
			}
			img, err := image.NewFromJSON(rawJSON)
			if err != nil {
				return nil, err
			}
			return img, nil
		}
	}
	return nil, nil
}

func parseLayers(r io.Reader, m *image.Manifest, i *image.Image) []*widgets.TreeNode {

	nodes := []*widgets.TreeNode{}

	for _, layer := range m.Layers {
		nodes = append(nodes, &widgets.TreeNode{
			Value: nodeValue(strings.TrimSuffix(layer, ".tar")),
			Nodes: nil,
		})
	}

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
			if node.Value == nil {
				node.Value = nodeValue(hdr.Name)
			} else {
				nodes = append(nodes, node)
				node = &widgets.TreeNode{}
			}

		} else {
			// if strings.EqualFold(filepath.Ext(hdr.Name), ".tar") {
			// 	for _, n := range makeNodes(tr) {
			// 		nodes = append(nodes, n)
			// 	}
			// }
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

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Exit(0)
	}()

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

	m, err := parseManifest(bufio.NewReader(f))
	if err != nil {
		panic(err)
	}

	f.Seek(0, 0)

	i, err := parseImage(bufio.NewReader(f), m)
	if err != nil {
		panic(err)
	}
	commands := ""
	space := regexp.MustCompile(`\s+`)
	for _, h := range i.History {
		if !h.EmptyLayer {
			commands = fmt.Sprintln(space.ReplaceAllString(h.CreatedBy, " "))
			break
		}
	}

	f.Seek(0, 0)

	nodes := parseLayers(bufio.NewReader(f), m, i)

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	l := widgets.NewTree()
	l.TextStyle = ui.NewStyle(ui.ColorYellow)
	l.WrapText = false
	l.SetNodes(nodes)
	l.Title = "Layers"
	l.TitleStyle.Fg = ui.ColorCyan
	l.PaddingTop = 1

	x, y := ui.TerminalDimensions()

	l.SetRect(0, 0, x, y)

	cmds := widgets.NewParagraph()
	cmds.Text = commands
	cmds.Title = "Commands"
	cmds.TitleStyle.Fg = ui.ColorCyan
	cmds.PaddingTop = 1
	cmds.PaddingBottom = 1
	cmds.PaddingLeft = 2
	cmds.PaddingRight = 2
	cmds.WrapText = false

	t := widgets.NewParagraph()
	t.Text = m.RepoTags[0]
	t.Title = "Image"
	t.PaddingTop = 1
	t.PaddingLeft = 2
	t.TitleStyle.Fg = ui.ColorCyan
	t.TextStyle.Modifier = ui.ModifierBold
	// t.TextStyle.Fg = ui.ColorMagenta
	t.TextStyle.Fg = ui.ColorBlue
	// t.Border = false

	grid := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)
	grid.Set(
		ui.NewRow(1.0/1,
			ui.NewCol(1.0/2, l),
			ui.NewCol(1.0/2,
				ui.NewRow(0.15/2, t),
				ui.NewRow(1.85/2, cmds),
			),
		),
	)

	ui.Render(grid)

	previousKey := ""
	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			return
		case "j", "<Down>":
			l.ScrollDown()
			idx := 0
			found := false
			for i, v := range m.Layers {
				if strings.EqualFold(strings.TrimSuffix(v, ".tar"), l.SelectedNode().Value.String()) {
					idx = i
					found = true
					break
				}
			}
			if found {
				ydx := 0
				for _, h := range i.History {
					if !h.EmptyLayer {
						if ydx == idx {
							// TODO: replace ; with new line also
							cmds.Text = strings.ReplaceAll(strings.TrimPrefix(h.CreatedBy, "/bin/sh -c "), "&&", "\n")
							// p.Text = space.ReplaceAllString(strings.TrimPrefix(h.CreatedBy, "/bin/sh -c "), " ")
							ui.Render(grid)
							break
						}
						ydx++
					}
				}
			}
		case "k", "<Up>":
			l.ScrollUp()
			idx := 0
			found := false
			for i, v := range m.Layers {
				if strings.EqualFold(strings.TrimSuffix(v, ".tar"), l.SelectedNode().Value.String()) {
					idx = i
					found = true
					break
				}
			}
			if found {
				ydx := 0
				for _, h := range i.History {
					if !h.EmptyLayer {
						if ydx == idx {
							cmds.Text = strings.ReplaceAll(strings.TrimPrefix(h.CreatedBy, "/bin/sh -c "), "&&", "\n")
							// p.Text = space.ReplaceAllString(strings.TrimPrefix(h.CreatedBy, "/bin/sh -c "), " ")
							ui.Render(grid)
							break
						}
						ydx++
					}
				}
			}
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
			// dateCmd := exec.Command("say", "-v", "Fiona", l.SelectedNode().Value.String())
			// dateCmd.Output()
			// if l.SelectedNode().Nodes != nil {
			// 	for _, node := range l.SelectedNode().Nodes {
			// 		dateCmd := exec.Command("say", "-v", "Fiona", node.Value.String())
			// 		dateCmd.Output()
			// 		time.Sleep(2 * time.Second)
			// 	}
			// }
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
