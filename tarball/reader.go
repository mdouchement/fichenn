package tarball

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// A Reader reads a directory and build a tarball stream.
type Reader struct {
	once sync.Once
	root string
	pr   *io.PipeReader
	pw   *io.PipeWriter
	tw   *tar.Writer
}

// NewReader returns a new Reader.
func NewReader(path string) (*Reader, error) {
	root, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	pR, pW := io.Pipe()

	r := &Reader{
		root: root,
		pr:   pR,
		pw:   pW,
		tw:   tar.NewWriter(pW),
	}

	return r, nil
}

// Name returns the name of the tarball.
func (r *Reader) Name() string {
	return r.root + ".tar"
}

// Read implements io.Reader.
func (r *Reader) Read(p []byte) (int, error) {
	r.once.Do(func() {
		go r.walk()
	})
	return r.pr.Read(p)
}

// Close implements io.Closer.
func (r *Reader) Close() error {
	if err := r.tw.Flush(); err != nil {
		r.pr.CloseWithError(err)
	}
	if err := r.tw.Close(); err != nil {
		r.pr.CloseWithError(err)
	}

	if err := r.pw.Close(); err != nil {
		r.pr.CloseWithError(err)
	}

	return r.pr.Close()
}

func (r *Reader) walk() {
	base := filepath.Base(r.root)

	err := filepath.Walk(r.root, func(fpath string, finfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create tar Header whether it's a file or directory
		hdr, err := tar.FileInfoHeader(finfo, finfo.Name())
		if err != nil {
			return err
		}

		relative := fpath
		if filepath.IsAbs(fpath) {
			relative, err = filepath.Rel(r.root, fpath)
			if err != nil {
				return err
			}
			relative = filepath.Join(base, relative)
		}
		hdr.Name = relative // Ensure header has relative file path

		if err := r.tw.WriteHeader(hdr); err != nil {
			return err
		}

		// Stop here if we got a directory
		if finfo.Mode().IsDir() {
			return nil
		}

		// Add file
		f, err := os.Open(fpath)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(r.tw, f)
		return err
	})

	if err != nil {
		r.pr.CloseWithError(err)
	}

	if err := r.tw.Flush(); err != nil {
		r.pr.CloseWithError(err)
	}
	if err := r.tw.Close(); err != nil {
		r.pr.CloseWithError(err)
	}

	if err := r.pw.Close(); err != nil {
		r.pr.CloseWithError(err)
	}
}
