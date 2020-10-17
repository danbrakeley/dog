package main

import (
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Revision History
//    0.3.0 - changed how files are encoded from []byte of decimal numbers to string of \x and ascii

var version = "0.3.0"

type Vars struct {
	RootPath   string
	TargetFile string
	Package    string
	Version    string
	Files      []FileEntry
}

type FileEntry struct {
	RelativePath string
	CleanName    string
	Size         int64
	Start        int
	End          int
}

func main() {
	vars := Vars{
		Version: version,
	}
	flag.StringVar(&vars.RootPath, "root", "", "the root path to look for files/folders")
	flag.StringVar(&vars.TargetFile, "file", "", "the go file to generate")
	flag.StringVar(&vars.Package, "package", "", "the packge in which the generated file will live")
	flag.Parse()

	if len(vars.RootPath) == 0 || len(vars.TargetFile) == 0 || len(vars.Package) == 0 {
		fmt.Printf("bpak generator v%s\n", version)
		fmt.Printf("Usage: %s\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
		return
	}

	if err := bpakGen(&vars); err != nil {
		panic(err)
	}
}

func bpakGen(vars *Vars) error {
	var total int64

	// Find all files
	vars.Files = make([]FileEntry, 0, 1024)
	err := filepath.Walk(vars.RootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		total += info.Size()
		vars.Files = append(vars.Files, FileEntry{
			RelativePath: path,
			CleanName:    filepath.ToSlash(strings.TrimPrefix(path, vars.RootPath)),
			Size:         info.Size(),
		})
		return nil
	})
	if err != nil {
		return fmt.Errorf("error while walking %s: %w", vars.RootPath, err)
	}

	// max exists because bpak does everything in ram and is meant for relatively small html/css/js files
	const max = 1024 * 1024 * 100
	if total > max {
		return fmt.Errorf("found %d bytes, which is currently outside the (somewhat arbitrary) max of %d", total, max)
	}

	if len(vars.Files) == 0 {
		return fmt.Errorf("no files found in %s", vars.RootPath)
	}

	// bpak found files
	var offset int
	allbytes := make([]byte, 0, int(total))
	fmt.Printf("__filename_______________________________________________original________gzip__\n")
	for i, e := range vars.Files {
		b, err := ioutil.ReadFile(e.RelativePath)
		if err != nil {
			return fmt.Errorf("error opening %s: %w", e.RelativePath, err)
		}

		// compress bytes
		var buf bytes.Buffer
		buf.Grow(int(e.Size))
		zw := gzip.NewWriter(&buf)
		if _, err := zw.Write(b); err != nil {
			return fmt.Errorf("error compressing %s: %w", e.RelativePath, err)
		}
		if err := zw.Close(); err != nil {
			return fmt.Errorf("error compressing %s: %w", e.RelativePath, err)
		}

		// copy bytes
		zb := buf.Bytes()
		allbytes = append(allbytes, zb...)

		fmt.Printf(" - %-50s %11d %11d\n", e.CleanName, e.Size, len(zb))

		// update file entry
		vars.Files[i].Start = offset
		vars.Files[i].End = offset + len(zb)

		offset = vars.Files[i].End
	}

	tmpl, err := template.New("").Parse(headerTemplate)
	if err != nil {
		return fmt.Errorf("error parsing header template: %w", err)
	}

	// build string to hold output
	var sb strings.Builder
	sb.Grow(len(headerTemplate) + int(total*4) + len(footer))

	// write header
	err = tmpl.Execute(&sb, vars)
	if err != nil {
		return fmt.Errorf("error executing header template: %w", err)
	}

	// encode raw bytes into a valid Go string
	for _, b := range allbytes {
		switch {
		case (b >= 0x20 && b < 0x22) || (b > 0x22 && b < 0x5C) || (b > 0x5C && b <= 0x7D):
			sb.WriteByte(b)
		case b == 0x22:
			sb.WriteByte(0x5C)
			sb.WriteByte(0x22)
		case b == 0x5C:
			sb.WriteByte(0x5C)
			sb.WriteByte(0x5C)
		default:
			sb.WriteString(fmt.Sprintf("\\x%02X", b))
		}
	}

	// write footer
	sb.WriteString(footer)

	// write string to target file
	f, err := os.OpenFile(vars.TargetFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	_, err = f.WriteString(sb.String())
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}

const (
	headerTemplate = `package {{.Package}}

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

//  _                 _
// | |               | |
// | |__  ____  _____| |  _
// |  _ \|  _ \(____ | |_/ )
// | |_) ) |_| / ___ |  _ (
// |____/|  __/\_____|_| \_) v{{.Version}}
//       |_|
//
// --- AUTOGENERATED FILE ---
// -- (don't edit by hand) --

func bpakGet(filename string) ([]byte, error) {
	filename = filepath.ToSlash(filename)
	if filename[0] != '/' {
		filename = fmt.Sprintf("/%s", filename)
	}
	var s string
	var size int
	switch filename {
{{range .Files}}	case "{{.CleanName}}":
		s = bpak_bytes[{{.Start}}:{{.End}}]
		size = {{.Size}}
{{end}}	default:
		return []byte{}, fmt.Errorf("file not found: %s", filename)
	}
	zr, err := gzip.NewReader(strings.NewReader(s))
	if err != nil {
		return []byte{}, fmt.Errorf("unable to begin to decompress %s: %w", filename, err)
	}
	var out bytes.Buffer
	out.Grow(size)
	n, err := io.Copy(&out, zr)
	if err != nil {
		return []byte{}, fmt.Errorf("unable to decompress %s: %w", filename, err)
	}
	if n != int64(size) {
		return []byte{}, fmt.Errorf("decompressed size of %s is %d, but expected %d", filename, n, size)
	}
	return out.Bytes(), nil
}

var bpak_bytes = "`
	footer = `"`
)
