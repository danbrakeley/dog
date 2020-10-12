package dog

import (
	"fmt"
	"io"
	"os"
	"time"
)

// MemFile implements http.File interface
type MemFile struct {
	Bytes  []byte
	Name   string
	Offset int64
}

func (f *MemFile) Close() error {
	return nil
}

func (f *MemFile) Read(p []byte) (n int, err error) {
	if f.Offset >= int64(len(f.Bytes)) {
		return 0, io.EOF
	}
	n = copy(p, f.Bytes[f.Offset:])
	f.Offset += int64(n)
	return n, nil
}

func (f *MemFile) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case 0:
		f.Offset = offset
	case 1:
		f.Offset += offset
	case 2:
		f.Offset = int64(len(f.Bytes)) - offset
	default:
		return 0, fmt.Errorf("unrecognized whence of %d", whence)
	}
	return f.Offset, nil
}

func (f *MemFile) Readdir(count int) ([]os.FileInfo, error) {
	return []os.FileInfo{}, fmt.Errorf("unimplemented")
}

func (f *MemFile) Stat() (os.FileInfo, error) {
	return &MemFileInfo{
		FileName: f.Name,
		FileSize: int64(len(f.Bytes)),
	}, nil
}

type MemFileInfo struct {
	FileName string
	FileSize int64
}

func (i *MemFileInfo) Name() string {
	return i.FileName
}

func (i *MemFileInfo) Size() int64 {
	return i.FileSize
}

func (i *MemFileInfo) Mode() os.FileMode {
	return 0444
}

func (i *MemFileInfo) ModTime() time.Time {
	return time.Time{}
}

func (i *MemFileInfo) IsDir() bool {
	return false
}

func (i *MemFileInfo) Sys() interface{} {
	return nil
}
