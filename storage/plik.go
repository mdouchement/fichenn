package storage

import (
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/knadh/koanf"
	plikpkg "github.com/mdouchement/fichenn/storage/plik"
)

type plik struct {
	konf *koanf.Koanf
	url  string
}

func (s *plik) Upload(name string, r io.Reader) error {
	client, err := plikpkg.NewDefault(s.konf.String("plik.url"))
	if err != nil {
		return err
	}

	s.url, err = client.Upload(name, r,
		plikpkg.UserAgent(UserAgent),
		plikpkg.TTL(time.Duration(s.konf.Duration("plik.ttl"))),
		plikpkg.OneShotFrom(s.konf.Bool("plik.one_shot")),
	)
	return err
}

func (s *plik) CommandArtefact() string {
	return fmt.Sprintf(`"%s" -o "%s"`, s.url, filepath.Base(s.url))
}
