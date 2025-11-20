package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"time"

	"github.com/atotto/clipboard"
	"github.com/k0kubun/go-ansi"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
	"github.com/mdouchement/fichenn/artifact"
	"github.com/mdouchement/fichenn/crypto"
	"github.com/mdouchement/fichenn/inode"
	"github.com/mdouchement/fichenn/storage"
	"github.com/mdouchement/fichenn/stream"
	"github.com/pkg/errors"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

const (
	runcom    = ".fichennrc"
	runcomEnv = "FICHENN_RC"
)

var (
	version  = "dev"
	revision = "none"
	date     = "unknown"
)

func main() {
	ctrl := new(controller)

	ctrl.Command = &cobra.Command{
		Use:     "finn",
		Short:   "Fichenn secured uploads",
		Version: fmt.Sprintf("%s - build %.7s @ %s - %s", version, revision, date, runtime.Version()),
		Args:    cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) (err error) {
			if ctrl.passphrase != "" {
				return ctrl.download(args[0])
			}

			return ctrl.upload(args[0])
		},
	}
	ctrl.Command.Flags().StringVarP(&ctrl.passphrase, "pass", "p", "", "passphrase used to decrypt")
	ctrl.Command.Flags().StringVarP(&ctrl.output, "output", "o", "", "write output to given destination")
	ctrl.Command.Flags().BoolVarP(&ctrl.chmodX, "chmod+x", "c", false, "perform `chmod +x' on downloaded file")
	ctrl.Command.Flags().BoolVarP(&ctrl.extract, "extract", "x", false, "Tarball extract")

	if err := ctrl.Command.Execute(); err != nil {
		log.Fatal(err)
	}
}

type controller struct {
	*cobra.Command
	passphrase string
	output     string
	chmodX     bool
	extract    bool
}

func (c *controller) download(url string) error {
	res, err := http.Get(url)
	if err != nil {
		return errors.Wrap(err, "could not get download")
	}
	defer res.Body.Close()

	bar := progress(res.ContentLength, "downloading")
	body := io.TeeReader(res.Body, bar)

	r, err := stream.NewReader(c.passphrase, ioutil.NopCloser(body))
	if err != nil {
		return err
	}

	if c.output == "" {
		c.output = filepath.Base(url)
	}

	//

	f, err := inode.NewWriter(c.output, c.extract)
	if err != nil {
		return errors.Wrap(err, "could not create destination file")
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	if err != nil {
		return errors.Wrap(err, "could not write file")
	}

	if err = f.Sync(); err != nil {
		return errors.Wrap(err, "could not finalize download")
	}

	//

	if c.chmodX {
		stat, err := f.Stat()
		if err != nil {
			if err == inode.ErrUnsupported {
				fmt.Println("could not chmod, file has only been extracted keeping original permissions")
				return nil
			}
			return errors.Wrap(err, "could not get output file info")
		}

		os.Chmod(c.output, stat.Mode()|os.FileMode(0o111)) // chmod +x <output>
	}

	return nil
}

func (c *controller) upload(src string) error {
	// Configuration
	konf, err := c.config()
	if err != nil {
		return errors.Wrap(err, "config")
	}

	//
	// Process upload

	storage, err := storage.NewFrom(konf)
	if err != nil {
		return errors.Wrap(err, "loading storage")
	}

	var artifact artifact.Artifact
	artifact.Password = crypto.NewPassword(konf.Int("passphrase_length"))
	fmt.Printf("Passphrase: %s\n\n", artifact.Password)

	f, err := inode.NewReader(src)
	if err != nil {
		return errors.Wrap(err, "open argument:")
	}
	defer f.Close()

	bar := progress(-1, "uploading")
	r := io.TeeReader(f, bar)

	pR, pW := io.Pipe()
	defer pR.Close()

	go func() {
		defer pW.Close()

		w, err := stream.NewWriter(artifact.Password, pW)
		if err != nil {
			pW.CloseWithError(err)
			return
		}
		defer w.Close()

		_, err = io.Copy(w, r)
		if err != nil {
			pW.CloseWithError(err)
		}
	}()

	artifact.Extractable = f.IsTarball()
	artifact.Filename = filepath.Base(f.Name())
	err = storage.Upload(&artifact, pR)
	if err != nil {
		return errors.Wrap(err, "could not upload")
	}

	command := artifact.CLI()
	fmt.Println("\nCommand:\n", command)

	if konf.Bool("clipboard") {
		clipboard.WriteAll(command)
		fmt.Println("Copied to the clipboard")
	}
	return nil
}

func (c *controller) config() (*koanf.Koanf, error) {
	// Configuration
	konf := koanf.New(".")

	defaults := map[string]any{
		"passphrase_length": 24,
		"storage":           "plik",
		"clipboard":         true,
		"plik": map[string]any{
			"url":      "https://plik.root.gg",
			"ttl":      "24h",
			"one_shot": false,
		},
	}
	if err := konf.Load(confmap.Provider(defaults, ""), nil); err != nil {
		panic(err) // We need to parse defaults
	}

	// Load user configuration
	cfg, err := lookup()
	if err != nil {
		return nil, errors.Wrap(err, "lookup")
	}
	fmt.Println("Use:", cfg)

	if err := konf.Load(file.Provider(cfg), toml.Parser()); err != nil {
		return nil, errors.Wrap(err, "parsing")
	}

	return konf, nil
}

func lookup() (string, error) {
	if path := os.Getenv(runcomEnv); path != "" {
		return path, nil
	}

	//

	workdir, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "current directory:")
	}

	usr, err := user.Current()
	if err != nil {
		return "", errors.Wrap(err, "could not get shell user")
	}

	for _, workdir := range []string{workdir, usr.HomeDir} {
		var previous string

		for workdir != previous {
			filename := filepath.Join(workdir, runcom)

			_, err := os.Stat(filename)
			if err == nil {
				return filename, nil
			}
			if os.IsNotExist(err) {
				previous = workdir
				workdir = filepath.Dir(workdir)
				continue
			}

			return "", err
		}
	}

	return "", errors.Errorf("no %s found", runcom)
}

func progress(size int64, caption string) *progressbar.ProgressBar {
	bar := progressbar.NewOptions64(size,
		progressbar.OptionSetDescription(caption),
		progressbar.OptionSetWidth(10),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)
	bar.RenderBlank()
	return bar
}
