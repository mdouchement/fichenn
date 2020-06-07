package inode

import (
	"os"

	"github.com/mdouchement/fichenn/tarball"
)

// A Reader can read a folder as a tarball or a file.
type Reader struct {
	name  string
	isDir bool
	read  func(p []byte) (int, error)
	close func() error
}

// NewReader returns a new Reader.
func NewReader(inode string) (*Reader, error) {
	s, err := os.Stat(inode)
	if err != nil {
		return nil, err
	}

	r := &Reader{
		isDir: s.IsDir(),
	}

	if !r.isDir {
		f, err := os.Open(inode)
		if err != nil {
			return nil, err
		}

		r.name = f.Name()
		r.read = f.Read
		r.close = f.Close

		return r, nil
	}

	tr, err := tarball.NewReader(inode)
	if err != nil {
		return nil, err
	}

	r.name = tr.Name()
	r.read = tr.Read
	r.close = tr.Close
	return r, nil
}

// Name returns the name of the reader.
func (r *Reader) Name() string {
	return r.name
}

// IsTarball returns true if the reader reads a folder as tarball.
func (r *Reader) IsTarball() bool {
	return r.isDir
}

// Read implements io.Reader.
func (r *Reader) Read(p []byte) (int, error) {
	return r.read(p)
}

// Close implements io.Closer.
func (r *Reader) Close() error {
	return r.close()
}
