package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTar(t *testing.T) {
	src, err := ioutil.TempDir(os.TempDir(), "tar-src-")
	assert.NoError(t, err)

	writeFileInDir(t, src, "foo.txt", "bar")

	dest, err := ioutil.TempFile(os.TempDir(), "tar-dest-")
	assert.NoError(t, err)

	err = Tar(src, dest)
	assert.NoError(t, err)

	info, err := dest.Stat()
	assert.NoError(t, err)

	assert.True(t, info.Size() > 0)
}

func writeFileInDir(t *testing.T, dir string, name string, input string) {
	path := filepath.Join(dir, name)
	err := ioutil.WriteFile(path, []byte(input), 0644)
	assert.NoError(t, err)
}
