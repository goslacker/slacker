package client

import (
	"bytes"
	"github.com/goslacker/slacker/core/errx"
	"io"
	"os"
)

type File struct {
	Name    string
	Content []byte
	Path    string
}

func (f *File) Open() (io.ReadCloser, error) {
	if len(f.Content) > 0 {
		return io.NopCloser(bytes.NewReader(f.Content)), nil
	} else if f.Path != "" {
		return os.Open(f.Path)
	}
	return nil, errx.New("file not found")
}
