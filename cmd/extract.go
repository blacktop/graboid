/*
Copyright © 2019 blacktop

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/blacktop/graboid/pkg/image"

	// "github.com/dustin/go-humanize"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/spf13/cobra"
)

const instructions = "<q> Quit | <enter> Expand/Collapse | <space> Extract"

// extractCmd represents the extract command
var extractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extract files from image",
	Long:  `Launches an interactive UI to explore and extract files from downloaded image TARs.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		if Verbose {
			log.SetLevel(log.DebugLevel)
		}

		tarPath := filepath.Clean(args[0])

		if _, err := os.Stat(tarPath); os.IsNotExist(err) {
			log.Fatalf("file does not exist: %s", tarPath)
		}

		fmt.Println()
		log.Infof(getFmtStr(), "[ANALYZING] Please wait...")

		f, err := os.Open(tarPath)
		if err != nil {
			return err
		}
		defer f.Close()

		i, err := image.Parse(bufio.NewReader(f))
		if err != nil {
			return err
		}
		// for _, layer := range i.Layers {
		// 	fmt.Printf("LAYER: %s (%s)\n", layer.Tree().Name, humanize.Bytes(layer.Size()))
		// 	fmt.Println(layer.Tree().String(false))
		// 	// node, err := layer.Tree().GetNode("/usr/local/share/bro/site")
		// 	// if err == nil {
		// 	// 	log.Info(layer.String())
		// 	// 	for _, v := range node.Children {
		// 	// 		log.Infof(v.Path())
		// 	// 	}
		// 	// }
		// }

		// nodes := i.Nodes()
		// fmt.Println(nodes)
		// return nil

		if err := ui.Init(); err != nil {
			log.Fatalf("failed to initialize termui: %v", err)
		}
		defer ui.Close()

		l := widgets.NewTree()
		l.TextStyle = ui.NewStyle(ui.ColorYellow)
		l.WrapText = false
		l.SetNodes(i.Nodes())
		l.Title = "Layers"
		l.TitleStyle.Fg = ui.ColorCyan
		l.PaddingTop = 1
		// x, y := ui.TerminalDimensions()
		// l.SetRect(0, 0, x, y)

		cmds := widgets.NewParagraph()
		if len(i.Layers) > 0 {
			cmds.Text = i.Layers[0].Command()
		}
		cmds.Title = "Commands"
		cmds.TitleStyle.Fg = ui.ColorCyan
		cmds.PaddingTop = 1
		cmds.PaddingBottom = 1
		cmds.PaddingLeft = 2
		cmds.PaddingRight = 2
		// cmds.WrapText = false

		t := widgets.NewParagraph()
		t.Text = strings.TrimPrefix(i.Tag, "library/")
		t.Title = "Image"
		t.PaddingTop = 1
		t.PaddingLeft = 2
		t.TitleStyle.Fg = ui.ColorCyan
		t.TextStyle.Modifier = ui.ModifierBold
		// t.TextStyle.Fg = ui.ColorMagenta
		t.TextStyle.Fg = ui.ColorBlue
		// t.Border = false

		instrns := widgets.NewParagraph()
		instrns.Text = instructions
		// instructions.Title = "Keymap"
		// instructions.PaddingTop = 1
		instrns.PaddingLeft = 2
		instrns.TitleStyle.Fg = ui.ColorCyan
		instrns.TextStyle.Modifier = ui.ModifierBold
		instrns.TextStyle.Fg = ui.ColorYellow
		// instructions.Border = false

		grid := ui.NewGrid()
		termWidth, termHeight := ui.TerminalDimensions()
		grid.SetRect(0, 0, termWidth, termHeight)
		grid.Set(
			ui.NewRow(1.0/1,
				ui.NewCol(1.0/2,
					ui.NewRow(1.85/2, l),
					ui.NewRow(0.15/2, instrns),
				),
				ui.NewCol(1.0/2,
					ui.NewRow(0.2/2, t),
					ui.NewRow(1.8/2, cmds),
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
				return nil
			case "j", "<Down>":
				l.ScrollDown()
				for _, layer := range i.Layers {
					if strings.EqualFold(l.SelectedNode().Value.String(), layer.TarID()) {
						cmds.Text = layer.Command()
					}
				}
			case "k", "<Up>":
				l.ScrollUp()
				for _, layer := range i.Layers {
					if strings.EqualFold(l.SelectedNode().Value.String(), layer.TarID()) {
						cmds.Text = layer.Command()
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
				instrns.Text = fmt.Sprintf("Extracting - %s", l.SelectedNode().Value.String())
				ui.Render(grid)
				f.Seek(0, 0)
				if err := i.Extract(bufio.NewReader(f), l.SelectedNode().Value.String(), 2); err != nil {
					return err
				}
				instrns.Text = "DONE!"
				ui.Render(grid)
				time.Sleep(1 * time.Second)
				instrns.Text = instructions
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
				// l.SetRect(0, 0, x, y)
				grid.SetRect(0, 0, x, y)
			}

			if previousKey == "g" {
				previousKey = ""
			} else {
				previousKey = e.ID
			}

			ui.Render(grid)
		}
	},
}

func init() {
	rootCmd.AddCommand(extractCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// extractCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// extractCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
