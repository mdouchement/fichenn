package main

import (
	"embed"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"runtime"

	"github.com/mdouchement/fichenn/webfinn/server"
	"github.com/mdouchement/logger"
	"github.com/pkg/errors"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	version  = "dev"
	revision = "none"
	date     = "unknown"
)

type configuration struct {
	Address string         `yaml:"addr"`
	Logger  string         `yaml:"logger"`
	Filters []*server.Rule `yaml:"filters"`
}

//go:embed assets/*
var assets embed.FS

func main() {
	var cfg string

	l := logrus.New()
	l.SetFormatter(&logger.LogrusTextFormatter{
		DisableColors:   false,
		ForceColors:     true,
		ForceFormatting: true,
		PrefixRE:        regexp.MustCompile(`^(\[.*?\])\s`),
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	log := logger.WrapLogrus(l)

	c := &cobra.Command{
		Use:     "webfinn",
		Short:   "Fichenn secured uploads from web browser",
		Version: fmt.Sprintf("%s - build %.7s @ %s - %s", version, revision, date, runtime.Version()),
		Args:    cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			log = log.WithField("version", c.Version)

			var config configuration
			var f *server.Filter
			{

				log.Infof("Reading configuration from %s", cfg)
				payload, err := ioutil.ReadFile(cfg)
				if err != nil {
					if err != nil {
						return errors.Wrapf(err, "could not read configuration file %s", cfg)
					}
				}

				err = yaml.Unmarshal(payload, &config)
				if err != nil {
					if err != nil {
						return errors.Wrapf(err, "could not parse configuration file %s", cfg)
					}
				}

				if config.Logger != "" {
					level, err := logrus.ParseLevel(config.Logger)
					if err != nil {
						return errors.Wrapf(err, "could not parse logger level %s", cfg)
					}
					l.SetLevel(level)
				}

				f, err = server.NewFilter(config.Filters)
				if err != nil {
					return errors.Wrapf(err, "could not build name resolver %s", cfg)
				}
			}

			//

			fs, err := fs.Sub(assets, "assets")
			if err != nil {
				return errors.Wrap(err, "could not get asset sub-filesystem")
			}

			app := server.NewApplication(log, f, fs)

			log.Println("Listening on", config.Address)
			corsfs := cors.Default().Handler(app)
			return http.ListenAndServe(config.Address, corsfs)
		},
	}
	c.Flags().StringVarP(&cfg, "config", "c", os.Getenv("WEBFINN_CONFIG"), "Server's configuration")

	if err := c.Execute(); err != nil {
		log.Fatal(err)
	}
}
