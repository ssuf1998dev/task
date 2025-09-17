package execext

import (
	"io"
	"os"

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
	if f, err := f.fs.OpenFile(filename, os.O_RDWR, 0o666); err == nil {
		return f
	}
	if f, err := f.fs.Create(filename); err == nil {
		return f
	}
	return devNull{}
}
