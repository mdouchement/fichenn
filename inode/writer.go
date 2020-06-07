package inode

import (
	"errors"
	"os"

	"github.com/mdouchement/fichenn/tarball"
)

// ErrUnsupported is used when an action is not supported.
var ErrUnsupported = errors.New("unsupported action")

// A Writer can write a folder from a tarball or a file.
type Writer struct {
	write func(p []byte) (int, error)
	sync  func() error
	close func() error
	stat  func() (os.FileInfo, error)
}

// NewWriter returns a new Writer.
func NewWriter(inode string, extract bool) (*Writer, error) {
	if extract {
		tw, err := tarball.NewWriter(inode)
		if err != nil {
			return nil, err
		}

		return &Writer{
			write: tw.Write,
			sync: func() error {
				return nil
			},
			close: tw.Close,
			stat: func() (os.FileInfo, error) {
				return nil, ErrUnsupported
			},
		}, nil
	}

	f, err := os.Create(inode)
	if err != nil {
		return nil, err
	}

	return &Writer{
		write: f.Write,
		sync:  f.Sync,
		close: f.Close,
		stat:  f.Stat,
	}, nil
}

// Write implements io.Writer.
func (w *Writer) Write(p []byte) (int, error) {
	return w.write(p)
}

// Close implements io.Closer.
func (w *Writer) Close() error {
	return w.close()
}

// Sync commits the current contents of the file to stable storage.
func (w *Writer) Sync() error {
	return w.sync()
}

// Stat returns the FileInfo structure describing file.
func (w *Writer) Stat() (os.FileInfo, error) {
	return w.stat()
}
