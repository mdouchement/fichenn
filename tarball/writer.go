package tarball

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// A Writer writes a directory from a tarball stream.
type Writer struct {
	once sync.Once
	root string
	pw   *io.PipeWriter
	pr   *io.PipeReader
	tr   *tar.Reader
}

// NewWriter returns a new Writer.
func NewWriter(dst string) (*Writer, error) {
	root, err := filepath.Abs(dst)
	if err != nil {
		return nil, err
	}

	if !strings.Contains(dst, string(filepath.Separator)) {
		root = "."
	}

	pR, pW := io.Pipe()

	return &Writer{
		root: root,
		pw:   pW,
		pr:   pR,
		tr:   tar.NewReader(pR),
	}, nil
}

// Write implements io.Writer.
func (w *Writer) Write(p []byte) (int, error) {
	w.once.Do(func() {
		go w.write()
	})
	return w.pw.Write(p)
}

// Close implements io.Closer.
func (w *Writer) Close() error {
	if err := w.pr.Close(); err != nil {
		w.pw.CloseWithError(err)
	}

	return w.pw.Close()
}

func (w *Writer) write() {
	for {
		hdr, err := w.tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			w.pr.CloseWithError(err)
			return
		}

		// Get path
		fname := hdr.Name
		if filepath.IsAbs(fname) {
			fname, err = filepath.Rel("/", fname)
			if err != nil {
				w.pr.CloseWithError(err)
				return
			}
		}
		fname = filepath.Join(w.root, fname)
		finfo := hdr.FileInfo()

		if finfo.Mode().IsDir() {
			// Create directory
			if err := os.MkdirAll(fname, 0755); err != nil {
				w.pr.CloseWithError(err)
				return
			}
			continue
		}

		// Create file
		if err := w.writeFile(fname, finfo); err != nil {
			w.pr.CloseWithError(err)
			return
		}
	}
}

func (w *Writer) writeFile(fname string, finfo os.FileInfo) error {
	f, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, finfo.Mode().Perm()) // Keep original permissions
	if err != nil {
		return err
	}
	defer f.Close()

	n, err := io.Copy(f, w.tr)
	if err != nil {
		return err
	}

	if n != finfo.Size() {
		return fmt.Errorf("file bad length: wrote %d, want %d", n, finfo.Size())
	}

	return f.Sync()
}
