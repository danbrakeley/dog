package dog

import (
	"net/http"
)

// bpakFileSystem is for use with http.FileServer
type bpakFileSystem struct{}

func (fs *bpakFileSystem) Open(name string) (http.File, error) {
	b, err := bpakGet(name)
	if err != nil {
		return nil, err
	}
	return &MemFile{
		Bytes: b,
		Name:  name,
	}, nil
}
