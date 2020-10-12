package dog

import (
	"io"
	"testing"
)

func Test_MemFile_Read(t *testing.T) {
	cases := []struct {
		Name            string
		Bytes           []byte
		ReadBufSizeDiff int
	}{
		{"simple", []byte("abcdefghijklmnop"), 0},
		{"smaller buffer", []byte("abcdefghijklmnop"), -1},
		{"larger buffer", []byte("abcdefghijklmnop"), 1},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			file := MemFile{
				Name:  tc.Name,
				Bytes: tc.Bytes,
			}

			total := 0
			for {
				buf := make([]byte, len(tc.Bytes)+tc.ReadBufSizeDiff)
				n, err := file.Read(buf)

				if n == 0 {
					if err != io.EOF {
						t.Fatalf("read 0 bytes, expected EOF, got %v", err)
					}
					if total != len(tc.Bytes) {
						t.Fatalf("read 0 bytes, but expecting %d more bytes", total-len(tc.Bytes))
					}
					break
				}

				if err != nil {
					t.Fatalf("unexpected read error: %v", err)
				}

				total += n

				if total > len(tc.Bytes) {
					t.Fatalf("read %d bytes, but there only should have been %d bytes to be read", total, len(tc.Bytes))
				}
			}
		})
	}
}

func Test_MemFile_Seek(t *testing.T) {
	cases := []struct {
		Name     string
		Bytes    []byte
		PreSeek  int64
		Whence   int
		SeekAmt  int64
		Expected []byte
	}{
		{"from start", []byte("abcdefghijklmnop"), 0, 0, 5, []byte("fghijklmnop")},
		{"from start after seek", []byte("abcdefghijklmnop"), 8, 0, 5, []byte("fghijklmnop")},
		{"from end", []byte("abcdefghijklmnop"), 0, 2, 10, []byte("abcdef")},
		{"from end after seek", []byte("abcdefghijklmnop"), 16, 2, 10, []byte("abcdef")},
		{"from middle", []byte("abcdefghijklmnop"), 5, 1, 4, []byte("jklmnop")},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			file := MemFile{
				Name:  tc.Name,
				Bytes: tc.Bytes,
			}

			ret1, err := file.Seek(tc.PreSeek, 0)
			if err != nil {
				t.Fatalf("unexpected error during pre-seek: %v", err)
			}
			if ret1 != tc.PreSeek {
				t.Fatalf("expected pre-seek to offset to %d, but got %d", tc.PreSeek, ret1)
			}

			ret2, err := file.Seek(tc.SeekAmt, tc.Whence)
			if err != nil {
				t.Fatalf("unexpected error during seek: %v", err)
			}
			switch tc.Whence {
			case 0:
				if ret2 != tc.SeekAmt {
					t.Fatalf("expected seek to offset to %d, but got %d", tc.SeekAmt, ret2)
				}
			case 1:
				expected := ret1 + tc.SeekAmt
				if ret2 != expected {
					t.Fatalf("expected seek to offset to %d, but got %d", expected, ret2)
				}
			case 2:
				expected := int64(len(tc.Bytes)) - tc.SeekAmt
				if ret2 != expected {
					t.Fatalf("expected seek to offset to %d, but got %d", expected, ret2)
				}
			default:
				t.Fatalf("whence of %d not supported", tc.Whence)
			}

			buf := make([]byte, len(tc.Expected))
			n, err := file.Read(buf)
			if err != nil {
				t.Fatalf("unexpected error during read after seek: %v", err)
			}
			if n != len(buf) {
				t.Fatalf("read %d byte(s), expected %d", n, len(buf))
			}
		})
	}
}
