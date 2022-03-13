package storage

import (
	"io"

	"github.com/knadh/koanf"
	"github.com/mdouchement/fichenn/artifact"
	"github.com/mdouchement/fichenn/ftime"
	plikpkg "github.com/mdouchement/fichenn/storage/plik"
)

type plik struct {
	konf *koanf.Koanf
}

func (s *plik) Upload(a *artifact.Artifact, r io.Reader) error {
	client, err := plikpkg.NewDefault(s.konf.String("plik.url"))
	if err != nil {
		return err
	}

	a.URL, err = client.Upload(a.Filename, r,
		plikpkg.UserAgent(UserAgent),
		plikpkg.Header(a.Header),
		plikpkg.TTL(ftime.MustParseDuration(s.konf.String("plik.ttl"))),
		plikpkg.OneShotFrom(s.konf.Bool("plik.one_shot")),
	)
	return err
}
