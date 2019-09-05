package main

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/apex/log"

	"github.com/blacktop/graboid/pkg/image"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type nodeValue string

func (nv nodeValue) String() string {
	return string(nv)
}

func parseFileSystem(files []image.File) (*image.FileTree, error) {

	tree := image.NewFileTree()
	tree.Name = "test"

	for _, file := range files {

		tree.FileSize += uint64(file.Data.Size())

		_, _, err := tree.AddPath(file.Path, file)
		if err != nil {
			return nil, err
		}
	}

	return tree, nil
}

func parseLayer(name string, r io.Reader) (*image.Layer, error) {

	var files []image.File

	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)

	for {
		hdr, err := tr.Next()

		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return nil, err
		}
		fmt.Println(hdr.Name)
		files = append(files, image.File{
			Name: hdr.FileInfo().Name(),
			Path: hdr.Name,
			Data: hdr.FileInfo(),
		})
	}

	tree, err := parseFileSystem(files)
	if err != nil {
		return nil, err
	}
	node, err := tree.GetNode("usr/local/share/bro/policy/protocols/ssl/")
	if err != nil {
		log.WithError(err).Error("Shit went sideways")
	} else {
		fmt.Println(node.Children)
	}

	layer := &image.Layer{
		Root: name,
		// Files: parsedFiles,
	}

	return layer, nil

}

func parseRepo(f *os.File) (*image.Repo, error) {

	var manifests []image.Manifest
	var img *image.Image
	var layers []*image.Layer

	gz, err := gzip.NewReader(bufio.NewReader(f))
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return nil, err
		}

		fmt.Println(hdr.Name)

		switch hdr.Typeflag {
		case tar.TypeDir:
			log.Error("should not have gotten a directory")
		case tar.TypeReg:
			if strings.EqualFold(filepath.Ext(hdr.Name), ".json") {
				if strings.EqualFold(hdr.Name, "manifest.json") {
					rawJSON, err := ioutil.ReadAll(tr)
					if err != nil {
						return nil, err
					}
					if err := json.Unmarshal(rawJSON, &manifests); err != nil {
						return nil, err
					}
					// } else if strings.EqualFold(hdr.Name, m.Config) {
				} else {
					rawJSON, err := ioutil.ReadAll(tr)
					if err != nil {
						return nil, err
					}
					img, err = image.NewFromJSON(rawJSON)
					if err != nil {
						return nil, err
					}
				}
			} else if strings.EqualFold(filepath.Ext(hdr.Name), ".tar") {
				layer, err := parseLayer(strings.TrimSuffix(hdr.Name, filepath.Ext(hdr.Name)), tr)
				if err != nil {
					return nil, err
				}
				layers = append(layers, layer)
			}
		}
	}

	repo := &image.Repo{
		Tag:           manifests[0].RepoTags[0],
		Created:       img.Created.String(),
		DockerVersion: img.DockerVersion,
	}

	nonEmptyLayerIdx := 0

	for _, history := range img.History {
		if !history.EmptyLayer {
			for _, layer := range layers {
				if strings.Contains(manifests[0].Layers[nonEmptyLayerIdx], layer.Root) {
					layer.Command = history.CreatedBy
					repo.Layers = append(repo.Layers, layer)
				}
			}
			nonEmptyLayerIdx++
		}
	}

	return repo, nil
}

func makeNodes(layers []*image.Layer) []*widgets.TreeNode {

	var nodes []*widgets.TreeNode

	return nodes
}

func main() {

	// c := make(chan os.Signal, 2)
	// signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	// go func() {
	// 	<-c
	// 	os.Exit(0)
	// }()

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

	repo, err := parseRepo(f)
	if err != nil {
		panic(err)
	}
	return

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	l := widgets.NewTree()
	l.TextStyle = ui.NewStyle(ui.ColorYellow)
	l.WrapText = false
	l.SetNodes(makeNodes(repo.Layers))
	l.Title = "Layers"
	l.TitleStyle.Fg = ui.ColorCyan
	l.PaddingTop = 1
	x, y := ui.TerminalDimensions()
	l.SetRect(0, 0, x, y)

	cmds := widgets.NewParagraph()
	// TODO
	// cmds.Text = commands
	cmds.Title = "Commands"
	cmds.TitleStyle.Fg = ui.ColorCyan
	cmds.PaddingTop = 1
	cmds.PaddingBottom = 1
	cmds.PaddingLeft = 2
	cmds.PaddingRight = 2
	cmds.WrapText = false

	t := widgets.NewParagraph()
	t.Text = "blacktop/ghidra"
	// TODO
	// t.Text = m.RepoTags[0]
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
			// idx := 0
			// found := false
			// for i, v := range m.Layers {
			// 	if strings.EqualFold(strings.TrimSuffix(v, ".tar"), l.SelectedNode().Value.String()) {
			// 		idx = i
			// 		found = true
			// 		break
			// 	}
			// }
			// if found {
			// 	ydx := 0
			// 	for _, h := range i.History {
			// 		if !h.EmptyLayer {
			// 			if ydx == idx {
			// 				// TODO: replace ; with new line also
			// 				cmds.Text = strings.ReplaceAll(strings.TrimPrefix(h.CreatedBy, "/bin/sh -c "), "&&", "\n")
			// 				// p.Text = space.ReplaceAllString(strings.TrimPrefix(h.CreatedBy, "/bin/sh -c "), " ")
			// 				ui.Render(grid)
			// 				break
			// 			}
			// 			ydx++
			// 		}
			// 	}
			// }
		case "k", "<Up>":
			l.ScrollUp()
			// idx := 0
			// found := false
			// for i, v := range m.Layers {
			// 	if strings.EqualFold(strings.TrimSuffix(v, ".tar"), l.SelectedNode().Value.String()) {
			// 		idx = i
			// 		found = true
			// 		break
			// 	}
			// }
			// if found {
			// 	ydx := 0
			// 	for _, h := range i.History {
			// 		if !h.EmptyLayer {
			// 			if ydx == idx {
			// 				cmds.Text = strings.ReplaceAll(strings.TrimPrefix(h.CreatedBy, "/bin/sh -c "), "&&", "\n")
			// 				// p.Text = space.ReplaceAllString(strings.TrimPrefix(h.CreatedBy, "/bin/sh -c "), " ")
			// 				ui.Render(grid)
			// 				break
			// 			}
			// 			ydx++
			// 		}
			// 	}
			// }
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
