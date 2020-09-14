package image

import (
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/wagoodman/dive/dive/filetree"
)

const (
	// dockerLayerFormat = "%-15s %7s  %s"
	dockerLayerFormat = "%7s  %s"
)

type Layer interface {
	ID() string
	TarID() string
	ShortID() string
	Index() int
	Command() string
	Size() uint64
	Tree() *filetree.FileTree
	String() string
}

// dockerLayer represents a Docker image layer and metadata
type dockerLayer struct {
	tarPath string
	history imageHistory
	index   int
	tree    *filetree.FileTree
}

func (dockerLayer *dockerLayer) TarID() string {
	return strings.TrimSuffix(dockerLayer.tarPath, ".tar")
}

func (dockerLayer *dockerLayer) ID() string {
	return dockerLayer.history.ID
}

func (dockerLayer *dockerLayer) Index() int {
	return dockerLayer.index
}

// Size returns the number of bytes that this image is.
func (dockerLayer *dockerLayer) Size() uint64 {
	return dockerLayer.history.Size
}

// Tree returns the file tree representing the current dockerLayer.
func (dockerLayer *dockerLayer) Tree() *filetree.FileTree {
	return dockerLayer.tree
}

func (dockerLayer *dockerLayer) Command() string {
	return strings.TrimPrefix(dockerLayer.history.CreatedBy, "/bin/sh -c ")
}

// ShortId returns the truncated id of the current dockerLayer.
func (dockerLayer *dockerLayer) ShortID() string {
	rangeBound := 15
	id := dockerLayer.ID()
	if length := len(id); length < 15 {
		rangeBound = length
	}
	id = id[0:rangeBound]

	// show the tagged image as the last dockerLayer
	// if len(dockerLayer.History.Tags) > 0 {
	// 	id = "[" + strings.Join(dockerLayer.History.Tags, ",") + "]"
	// }

	return id
}

// String represents a dockerLayer in a columnar format.
func (dockerLayer *dockerLayer) String() string {

	if dockerLayer.index == 0 {
		return fmt.Sprintf(dockerLayerFormat,
			// dockerLayer.ShortId(),
			// fmt.Sprintf("%d",dockerLayer.Index()),
			humanize.Bytes(dockerLayer.Size()),
			"FROM "+dockerLayer.ShortID())
	}
	return fmt.Sprintf(dockerLayerFormat,
		// dockerLayer.ShortId(),
		// fmt.Sprintf("%d",dockerLayer.Index()),
		humanize.Bytes(dockerLayer.Size()),
		dockerLayer.Command())
}
