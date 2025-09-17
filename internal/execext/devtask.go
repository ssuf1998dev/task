package execext

import (
	"io"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
)

type devTaskFs struct {
	fs billy.Filesystem
}

var devTask = devTaskFs{fs: memfs.New()}

func (f *devTaskFs) File(filename string) io.ReadWriteCloser {
	if len(filename) <= 0 {
		return devNull{}
	}
	if f, err := f.fs.Open(filename); err == nil {
		return f
	}
	if f, err := f.fs.Create(filename); err == nil {
		return f
	}
	return devNull{}
}
