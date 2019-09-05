package image

// credit: github.com/dive

import (
	"fmt"
	"github.com/google/uuid"
	"os"
	"strings"
)

const (
	whiteoutPrefix       = ".wh."
	doubleWhiteoutPrefix = ".wh..wh.."
)

// FileTree represents a set of files, directories, and their relations.
type FileTree struct {
	ID       uuid.UUID
	Root     *FileNode
	Name     string
	Size     int
	FileSize uint64
}

// FileNode represents a single file, its relation to files beneath it, the tree it exists in, and the metadata of the given file.
type FileNode struct {
	Tree     *FileTree
	Parent   *FileNode
	Name     string
	Data     os.FileInfo
	Children map[string]*FileNode
	path     string
}

// NewFileTree creates an empty FileTree
func NewFileTree() (tree *FileTree) {
	tree = new(FileTree)
	tree.Size = 0
	tree.Root = new(FileNode)
	tree.Root.Tree = tree
	tree.Root.Children = make(map[string]*FileNode)
	tree.ID = uuid.New()
	return tree
}

// NewNode creates a new FileNode relative to the given parent node with a payload.
func NewNode(parent *FileNode, name string, data os.FileInfo) (node *FileNode) {
	node = new(FileNode)
	node.Name = name
	node.Data = data

	node.Children = make(map[string]*FileNode)
	node.Parent = parent
	if parent != nil {
		node.Tree = parent.Tree
	}

	return node
}

// AddChild creates a new node relative to the current FileNode.
func (node *FileNode) AddChild(name string, data os.FileInfo) (child *FileNode) {
	// never allow processing of purely whiteout flag files (for now)
	if strings.HasPrefix(name, doubleWhiteoutPrefix) {
		return nil
	}

	child = NewNode(node, name, data)
	if node.Children[name] != nil {
		// tree node already exists, replace the payload, keep the children
		node.Children[name].Data = data
	} else {
		node.Children[name] = child
		node.Tree.Size++
	}

	return child
}

// GetNode fetches a single node when given a slash-delimited string from root ('/') to the desired node (e.g. '/a/node/path')
func (tree *FileTree) GetNode(path string) (*FileNode, error) {
	nodeNames := strings.Split(strings.Trim(path, "/"), "/")
	node := tree.Root
	for _, name := range nodeNames {
		if name == "" {
			continue
		}
		if node.Children[name] == nil {
			return nil, fmt.Errorf("path does not exist: %s", path)
		}
		node = node.Children[name]
	}
	return node, nil
}

// Remove deletes the current FileNode from it's parent FileNode's relations.
func (node *FileNode) Remove() error {
	if node == node.Tree.Root {
		return fmt.Errorf("cannot remove the tree root")
	}
	for _, child := range node.Children {
		err := child.Remove()
		if err != nil {
			return err
		}
	}
	delete(node.Parent.Children, node.Name)
	node.Tree.Size--
	return nil
}

// AddPath adds a new node to the tree with the given payload
func (tree *FileTree) AddPath(path string, file File) (*FileNode, []*FileNode, error) {
	nodeNames := strings.Split(strings.Trim(path, "/"), "/")
	node := tree.Root
	addedNodes := make([]*FileNode, 0)
	for idx, name := range nodeNames {
		if name == "" {
			continue
		}
		// find or create node
		if node.Children[name] != nil {
			node = node.Children[name]
		} else {
			// don't add paths that should be deleted
			if strings.HasPrefix(name, doubleWhiteoutPrefix) {
				return nil, addedNodes, nil
			}

			// don't attach the payload. The payload is destined for the
			// Path's end node, not any intermediary node.
			node = node.AddChild(name, file.Data) // TODO: what should this be
			addedNodes = append(addedNodes, node)

			if node == nil {
				// the child could not be added
				return node, addedNodes, fmt.Errorf(fmt.Sprintf("could not add child node: '%s' (path:'%s')", name, path))
			}
		}

		// attach payload to the last specified node
		if idx == len(nodeNames)-1 {
			node.Data = file.Data
		}

	}
	return node, addedNodes, nil
}

// RemovePath removes a node from the tree given its path.
func (tree *FileTree) RemovePath(path string) error {
	node, err := tree.GetNode(path)
	if err != nil {
		return err
	}
	return node.Remove()
}
