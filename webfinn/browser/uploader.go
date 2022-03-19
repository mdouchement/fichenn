package main

import (
	"io"
	"sync"
	sjs "syscall/js"

	"github.com/knadh/koanf"
	"github.com/mdouchement/fichenn/artifact"
	"github.com/mdouchement/fichenn/crypto"
	"github.com/mdouchement/fichenn/storage"
	"github.com/mdouchement/fichenn/stream"
	"github.com/pkg/errors"
)

const chunksize = 1024

type Uploader struct {
	konf   *koanf.Koanf
	blob   sjs.Value
	slice  string
	size   int
	offset int
}

func NewUploader(konf *koanf.Koanf, blob sjs.Value) (*Uploader, error) {
	u := &Uploader{
		konf: konf,
		blob: blob,
		size: blob.Get("size").Int(),
	}

	u.slice = "slice"
	slice := blob.Get(u.slice)
	if slice.IsNull() {
		u.slice = "mozSlice"
		slice = blob.Get(u.slice)
	}
	if slice.IsNull() {
		u.slice = "webkitSlice"
		slice = blob.Get(u.slice)
	}
	if slice.IsNull() {
		return nil, errors.New("could not get slice")
	}

	return u, nil
}

func (u *Uploader) Upload(artifact *artifact.Artifact) error {
	//
	// Process upload
	storage, err := storage.NewFrom(u.konf)
	if err != nil {
		return errors.Wrap(err, "loading storage")
	}

	artifact.Password = crypto.NewPassword(u.konf.Int("passphrase_length"))

	pR, pW := io.Pipe()
	defer pR.Close()

	var w io.WriteCloser

	//

	var wg sync.WaitGroup

	onload := sjs.FuncOf(func(this sjs.Value, args []sjs.Value) interface{} {
		defer wg.Done()

		if err := args[0].Get("target").Get("error"); !err.IsNull() {
			pW.CloseWithError(errors.New(err.String()))
			return nil
		}

		blob := args[0].Get("target").Get("result")
		blob = sjs.Global().Get("Uint8Array").New(blob)

		chunk := make([]byte, chunksize)
		n := sjs.CopyBytesToGo(chunk, blob)

		_, err = w.Write(chunk[:n])
		if err != nil {
			pW.CloseWithError(err)
		}
		return nil
	})

	go func() {
		defer pW.Close()

		w, err = stream.NewWriter(artifact.Password, pW)
		if err != nil {
			pW.CloseWithError(err)
			return
		}
		defer w.Close()

		for u.offset < u.size {
			wg.Add(1)

			jfr := sjs.Global().Get("FileReader").New()
			jfr.Set("onload", onload)

			jfr.Call("readAsArrayBuffer", u.blob.Call(u.slice, u.offset, u.offset+chunksize))
			u.offset += chunksize
		}

		wg.Wait() // Wait for all 'onload' callbacks execution
	}()

	err = storage.Upload(artifact, pR)
	if err != nil {
		return errors.Wrap(err, "could not upload")
	}

	return nil
}
