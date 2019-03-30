package samplefs

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/apisite/tpl2x/lookupfs"
)

func TestFS(t *testing.T) {
	fs := FS()
	_, ok := fs.(lookupfs.FileSystem)
	assert.True(t, ok)
}
