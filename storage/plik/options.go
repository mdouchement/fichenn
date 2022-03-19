package plik

import (
	"net/url"
	"time"
)

type (
	// Option is a function used in the Functional Options pattern.
	Option func(*Options)

	// Options contains upload characteristics.
	Options struct {
		ID       string `json:"id"`
		Creation int64  `json:"uploadDate"`
		TTL      int    `json:"ttl"`

		// Files map[string]*File `json:"files"`

		UploadToken string `json:"uploadToken,omitempty"`
		User        string `json:"user,omitempty"`
		Token       string `json:"token,omitempty"`
		IsAdmin     bool   `json:"admin"`

		Stream    bool `json:"stream"`
		OneShot   bool `json:"oneShot"`
		Removable bool `json:"removable"`

		ProtectedByPassword bool   `json:"protectedByPassword"`
		Login               string `json:"login,omitempty"`
		Password            string `json:"password,omitempty"`

		// ProtectedByYubikey bool   `json:"protectedByYubikey"`
		// Yubikey            string `json:"yubikey,omitempty"`

		authorization string
		useragent     string
		header        url.Values
	}
)

// TTL is the time before removing remote file.
func TTL(ttl time.Duration) Option {
	return func(opts *Options) {
		opts.TTL = int(ttl.Seconds())
	}
}

// OneShot removes the remote file after the first download.
func OneShot() Option {
	return func(opts *Options) {
		opts.OneShot = true
	}
}

// OneShotFrom removes the remote file after the first download.
func OneShotFrom(b bool) Option {
	return func(opts *Options) {
		opts.OneShot = b
	}
}

// BasicAuth protects upload with HTTP basic auth.
func BasicAuth(login, password string) Option {
	return func(opts *Options) {
		opts.Login = login
		opts.Password = password
	}
}

// UserAgent sets the User-Agent for each requests.
func UserAgent(name string) Option {
	return func(opts *Options) {
		opts.useragent = name
	}
}

// Header sets the headers for each requests.
func Header(header url.Values) Option {
	return func(opts *Options) {
		opts.header = header
	}
}
