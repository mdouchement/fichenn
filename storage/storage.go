package storage

import (
	"io"

	"github.com/knadh/koanf"
	"github.com/mdouchement/fichenn/artifact"
	"github.com/pkg/errors"
)

// UserAgent used in requests.
const UserAgent = "fichenn/1.0"

// A Storage is an interface used for uploading secured data.
type Storage interface {
	Upload(a *artifact.Artifact, r io.Reader) error
}

// NewFrom returns a new storage based on the given configuration.
func NewFrom(konf *koanf.Koanf) (Storage, error) {
	switch konf.String("storage") {
	case "plik":
		return &plik{
			konf: konf,
		}, nil
	default:
		return nil, errors.Errorf("unknown storage: %s", konf.String("storage"))
	}
}
