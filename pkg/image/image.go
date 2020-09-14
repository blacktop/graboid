package image

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/gizak/termui/v3/widgets"
	"github.com/wagoodman/dive/dive/filetree"
)

// Parse parses an image tar.gz file
func Parse(r io.Reader) (*Tar, error) {

	i := &Tar{}

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

		switch hdr.Typeflag {
		case tar.TypeSymlink:
		case tar.TypeReg:
			switch filepath.Ext(hdr.Name) {
			case ".json":
				if strings.EqualFold(hdr.Name, "manifest.json") {
					rawJSON, err := ioutil.ReadAll(tr)
					if err != nil {
						return nil, err
					}
					var manifests []Manifest
					if err := json.Unmarshal(rawJSON, &manifests); err != nil {
						return nil, err
					}
					i.Manifest = manifests[0]
				} else {
					rawJSON, err := ioutil.ReadAll(tr)
					if err != nil {
						return nil, err
					}
					i.Config, err = NewFromJSON(rawJSON)
					if err != nil {
						return nil, err
					}
				}
			case ".tar":
				// if err = i.processLayerTar(hdr.Name, currentLayer, tar.NewReader(tr)); err != nil {
				if err = i.processLayerTar(hdr.Name, 0, tr); err != nil {
					return nil, err
				}
			}
		}
	}

	i.Tag = i.Manifest.RepoTags[0]
	i.Layers = make([]Layer, len(i.RefTrees))

	nonEmptyLayerIdx := 0 // TODO

	for _, history := range i.Config.History {
		if !history.EmptyLayer {
			for _, tree := range i.RefTrees {
				if strings.Contains(i.Manifest.Layers[nonEmptyLayerIdx], tree.Name) {
					history.Size = tree.FileSize
					i.Layers[nonEmptyLayerIdx] = &dockerLayer{
						history: history,
						index:   nonEmptyLayerIdx,
						tree:    tree,
						tarPath: i.Manifest.Layers[nonEmptyLayerIdx],
					}
				}
			}
			nonEmptyLayerIdx++
		}
	}

	return i, nil
}

func (i *Tar) processLayerTar(name string, layerIdx uint, reader io.Reader) error {
	tree := filetree.NewFileTree()
	tree.Name = name

	fileInfos, err := i.getFileList(reader)
	if err != nil {
		return err
	}

	for _, element := range fileInfos {
		tree.FileSize += uint64(element.Size)

		_, _, err := tree.AddPath(element.Path, element)
		if err != nil {
			return err
		}
	}

	i.RefTrees = append(i.RefTrees, tree) // TODO
	return nil
}

func (i *Tar) getFileList(r io.Reader) ([]filetree.FileInfo, error) {

	var files []filetree.FileInfo

	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		name := header.Name

		switch header.Typeflag {
		case tar.TypeXGlobalHeader:
			return nil, fmt.Errorf("unexptected tar file: (XGlobalHeader): type=%v name=%s", header.Typeflag, name)
		case tar.TypeXHeader:
			return nil, fmt.Errorf("unexptected tar file (XHeader): type=%v name=%s", header.Typeflag, name)
		default:
			files = append(files, filetree.NewFileInfoFromTarHeader(tr, header, name))
		}
	}

	return files, nil
}

// Extract extracts a path from a tar.gz and can handle a set depth of nested tar.gz(s)
func (i *Tar) Extract(r io.Reader, path string, depth int) error {

	depth--

	gz, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)

	for {
		hdr, err := tr.Next()

		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}

		switch hdr.Typeflag {
		case tar.TypeSymlink:
		case tar.TypeReg:
			switch filepath.Ext(hdr.Name) {
			case ".tar":
				if depth > 0 {
					if err = i.Extract(tr, path, depth); err != nil {
						return err
					}
				}
			default:
				// fmt.Println(hdr.Name)
				var name string
				if hdr.Typeflag == tar.TypeSymlink {
					name = hdr.Linkname
				} else {
					name = hdr.Name
				}
				if strings.Contains(path, name) {
					// fmt.Println(name)
					// fmt.Println(filepath.Dir(name))
					// os.MkdirAll(filepath.Dir(name), os.ModePerm)
					f, err := os.OpenFile(filepath.Base(name), os.O_CREATE|os.O_RDWR, 0644)
					// f, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR, os.FileMode(hdr.Mode)) // TODO
					if err != nil {
						return err
					}
					if _, err := io.Copy(f, tr); err != nil {
						return err
					}
					f.Close()
					return nil
				}
			}
		}
	}

	return nil
	// return fmt.Errorf("did not find path: %s", path)
}

type nodeValue string

func (nv nodeValue) String() string {
	return string(nv)
}

// Nodes returns the filetree as widget treenodes
func (i *Tar) Nodes() []*widgets.TreeNode {

	var nodes []*widgets.TreeNode
	// currentTree := 0

	for _, layer := range i.Layers {
		var treeNodes []*widgets.TreeNode
		// tree := layer.Tree()
		// visitor := func(node *filetree.FileNode) error {

		// 	if !node.IsWhiteout() {
		// 		sizer := func(curNode *filetree.FileNode) error {
		// 			sizeBytes += curNode.Name
		// 			return nil
		// 		}
		// 		previousTreeNode, err := tree.GetNode(node.Path())
		// 		if err != nil {
		// 			log.Debug(fmt.Sprintf("CurrentTree: %d : %s", currentTree, err))
		// 		} else if previousTreeNode.Data.FileInfo.IsDir {
		// 			err = previousTreeNode.VisitDepthChildFirst(sizer, nil)
		// 			if err != nil {
		// 				log.Errorf("unable to propagate dir: %+v", err)
		// 			}
		// 		}
		// 	}

		// 	data.Nodes = append(data.Nodes, node)

		// 	return nil
		// }
		// visitEvaluator := func(node *filetree.FileNode) bool {
		// 	return node.IsLeaf()
		// }
		// if err := tree.VisitDepthChildFirst(visitor, visitEvaluator); err != nil {
		// 	return nil
		// }

		layer.Tree().VisitDepthChildFirst(func(node *filetree.FileNode) error {
			for _, child := range node.Children {
				if !child.IsWhiteout() {
					display := child.Path()
					if child.Data.FileInfo.TypeFlag == tar.TypeSymlink || child.Data.FileInfo.TypeFlag == tar.TypeLink {
						display += " → " + child.Data.FileInfo.Linkname
					}
					display += fmt.Sprintf(" (%s)", humanize.Bytes(uint64(child.Data.FileInfo.Size)))
					treeNodes = append(treeNodes, &widgets.TreeNode{
						Value: nodeValue(display),
						// Nodes: child.Children,
					})
				}
			}
			return nil
		}, nil)

		nodes = append(nodes, &widgets.TreeNode{
			Value: nodeValue(layer.TarID()),
			Nodes: treeNodes,
		})
	}

	return nodes
}
