package artifact

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"

	"github.com/fxamacker/cbor/v2"
)

// An Artifact represents an uploaded file.
type Artifact struct {
	URL         string     `cbor:"u"`
	Password    string     `cbor:"p"`
	Filename    string     `cbor:"f"`
	Extractable bool       `cbor:"-"`
	Header      url.Values `cbor:"-"`
}

// ParseFromCLI parses artifact from cli.
func ParseFromCLI(cli string) (Artifact, error) {
	var artifact Artifact

	p, err := parsecli(cli)
	if err != nil {
		return artifact, err
	}

	artifact.URL = p.params["url"]
	artifact.Password = p.params["password"]
	artifact.Filename = p.params["output"]
	_, artifact.Extractable = p.params["extractable"]
	return artifact, nil
}

// ParseFromLink parses artifact from link's param.
func ParseFromLink(param string) (Artifact, error) {
	var artifact Artifact

	payload, err := base64.RawURLEncoding.DecodeString(param)
	if err != nil {
		return artifact, err
	}

	return artifact, cbor.Unmarshal(payload, &artifact)
}

// CLI returns the download artifact command.
func (a Artifact) CLI() string {
	command := fmt.Sprintf(
		"finn --pass %s %s -o %s",
		strconv.Quote(a.Password),
		strconv.Quote(a.URL),
		strconv.Quote(a.Filename),
	)

	if a.Extractable {
		command += " --extract"
	}
	return command
}

// Link returns the download artifact URL used by webfinn.
// host is the URL of the webfinn instance.
func (a Artifact) Link(host url.URL) string {
	payload, err := cbor.Marshal(a)
	if err != nil {
		panic(err) // Should never occur
	}

	u := url.URL{
		Scheme: host.Scheme,
		Host:   host.Host,
		Path:   "/download/" + base64.RawURLEncoding.EncodeToString(payload),
	}
	return u.String()
}
