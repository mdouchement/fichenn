package artifact

import (
	"bytes"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type cliparser struct {
	pass     *regexp.Regexp
	output   *regexp.Regexp
	previous string
	params   map[string]string
}

func parsecli(cli string) (cliparser, error) {
	p := cliparser{
		pass:   regexp.MustCompile(`^(?:-p|--pass)=?(.*)$`),
		output: regexp.MustCompile(`^(?:-o|--output)=?(.*)$`),
		params: make(map[string]string),
	}

	buf := bytes.NewBufferString(cli)
	for {
		token, err := buf.ReadString(' ')
		exit := err == io.EOF
		if err != nil && !exit {
			return p, err
		}

		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}

		err = p.parse(token)
		if err != nil {
			return p, err
		}

		if exit {
			return p, nil
		}
	}
}

func (p *cliparser) parse(token string) (err error) {
	switch token {
	case "finn":
		return
	case "-c", "--chmod+x":
		return
	case "-x", "--extract":
		p.params["extractable"] = ""
		return
	}

	m := p.pass.FindStringSubmatch(token)
	if len(m) == 2 {
		if m[1] != "" {
			p.params["password"], err = strconv.Unquote(m[1])
			return
		}
		p.previous = "password"
		return
	}

	m = p.output.FindStringSubmatch(token)
	if len(m) == 2 {
		if m[1] != "" {
			p.params["output"], err = strconv.Unquote(m[1])
			return
		}
		p.previous = "output"
		return
	}

	//
	//

	token, err = strconv.Unquote(token)
	if err != nil {
		return err
	}

	if p.previous == "" && strings.HasPrefix(token, "http") {
		p.params["url"] = token
		return
	}

	p.params[p.previous] = token
	p.previous = ""
	return nil
}
