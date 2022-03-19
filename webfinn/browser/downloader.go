package main

import (
	"io"
	"net/http"
	sjs "syscall/js"

	"github.com/mdouchement/fichenn/artifact"
	"github.com/mdouchement/fichenn/stream"
)

type Downloader struct {
	artifact artifact.Artifact
}

func NewDownloader(artifact artifact.Artifact) *Downloader {
	return &Downloader{
		artifact: artifact,
	}
}

func (d *Downloader) Download() error {
	filestream := sjs.Global().Get("streamSaver").Call("createWriteStream", d.artifact.Filename)

	go func() {
		resp, err := http.Get(d.artifact.URL)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		//

		r, err := stream.NewReader(d.artifact.Password, resp.Body)
		if err != nil {
			panic(err)
		}
		defer r.Close()

		//

		sjs.Global().Get("window").Set("writer", filestream.Call("getWriter"))
		writer := sjs.Global().Get("window").Get("writer")
		defer writer.Call("close")

		buf := make([]byte, 4096)
		for {
			n, err := r.Read(buf)
			if err != nil && err != io.EOF {
				panic(err)
			}

			blob := sjs.Global().Get("Uint8Array").New(n)
			sjs.CopyBytesToJS(blob, buf[:n])
			writer.Call("write", blob)
			if err == io.EOF {
				return
			}
		}
	}()

	return nil
}
