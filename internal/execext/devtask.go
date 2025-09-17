package execext

import (
	"io"
	"os"
	"sync"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
)

type devTask struct {
	fs billy.Filesystem
	mu sync.Mutex
}

var DevTask = devTask{fs: memfs.New()}

func (f *devTask) File(filename string) io.ReadWriteCloser {
	if len(filename) <= 0 || filename == "/" {
		return DevNull{}
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f, err := f.fs.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0o666); err == nil {
		return f
	}
	return DevNull{}
}
