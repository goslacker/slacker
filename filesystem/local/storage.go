package local

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func NewLocal(root string, urlPrefix string) *Local {
	r, err := filepath.Abs(root)
	if err != nil {
		panic(err)
	}
	return &Local{
		root:      r,
		urlPrefix: urlPrefix,
		perm:      os.ModePerm,
	}
}

type Local struct {
	root      string
	urlPrefix string
	perm      fs.FileMode
}

func (l *Local) Put(path string, content []byte) (err error) {
	path = l.preparePath(path)
	err = os.MkdirAll(filepath.Dir(path), os.ModeDir)
	if err != nil {
		return
	}

	err = os.WriteFile(path, content, l.perm)
	if err != nil {
		err = fmt.Errorf("write file failed: %w", err)
		return
	}
	return
}

func (l *Local) Del(path string) (err error) {
	err = os.Remove(l.preparePath(path))
	if err != nil {
		err = fmt.Errorf("remove file failed: %w", err)
		return
	}
	return
}

func (l *Local) Get(path string) (content []byte, err error) {
	content, err = os.ReadFile(l.preparePath(path))
	if err != nil {
		err = fmt.Errorf("read file failed: %w", err)
		return
	}
	return
}

func (l *Local) Url(path string) (url string, err error) {
	return strings.Trim(l.urlPrefix, "/") + "/" + strings.Trim(strings.Replace(path, "\\", "/", -1), "/"), nil
}

func (l *Local) preparePath(path string) string {
	return filepath.Join(l.root, path)
}
