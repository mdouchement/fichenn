package plik

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"

	"github.com/pkg/errors"
)

type (
	// A Client defines all interactions that can be performed on a Plik server.
	Client interface {
		Upload(name string, r io.Reader, opts ...Option) (string, error)
	}

	client struct {
		http     *http.Client
		endpoint string
	}
)

// NewDefault returns a new Client with default HTTP client.
func NewDefault(endpoint string) (Client, error) {
	return New(http.DefaultClient, endpoint)
}

// New returns a new Client.
func New(c *http.Client, endpoint string) (Client, error) {
	_, err := url.Parse(endpoint)
	return &client{endpoint: endpoint, http: c}, errors.Wrap(err, "could not parse endpoint")
}

func (c *client) Upload(name string, r io.Reader, opts ...Option) (string, error) {
	o := &Options{
		TTL: 86400,
	}
	for _, setter := range opts {
		setter(o)
	}

	o, err := c.bucket(o)
	if err != nil {
		return "", errors.Wrap(err, "bucket")
	}

	id, name, err := c.multipart(name, r, o)
	if err != nil {
		return "", errors.Wrap(err, "upload")
	}

	u, err := url.Parse(c.endpoint)
	if err != nil {
		return "", errors.Wrap(err, "could not parse endpoint")
	}
	u.Path = path.Join("/file", o.ID, id, name)
	return u.String(), err
}

// returns an upload bucket where to put files.
func (c *client) bucket(params *Options) (*Options, error) {
	u, err := url.Parse(c.endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse endpoint")
	}
	u.Path = "/upload"

	//
	// Build request
	body, err := json.Marshal(params)
	if err != nil {
		return nil, errors.Wrap(err, "could not serialize params")
	}

	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewReader(body))
	if err != nil {
		return nil, errors.Wrap(err, "could not build request")
	}
	req.Close = true
	for k, vs := range params.header {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	if params.useragent != "" {
		req.Header.Add("User-Agent", params.useragent)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	//
	// Perform request
	res, err := c.http.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "could not perform request")
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		body, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, errors.Wrapf(err, "%d", res.StatusCode)
		}
		return nil, errors.Errorf("[%d] %s", res.StatusCode, body)
	}

	params.authorization = res.Header.Get("Authorization") // Get BasicAuth authorization for the upload.

	dec := json.NewDecoder(res.Body)
	return params, dec.Decode(&params)
}

// upload io.Reader with given name.
func (c *client) multipart(name string, r io.Reader, params *Options) (string, string, error) {
	u, err := url.Parse(c.endpoint)
	if err != nil {
		return "", "", errors.Wrap(err, "could not parse endpoint")
	}
	u.Path = "/file/" + params.ID

	//
	// Build request
	contentType, stream := c.streamer(name, r)

	req, err := http.NewRequest(http.MethodPost, u.String(), stream)
	if err != nil {
		return "", "", errors.Wrap(err, "could not build request")
	}
	req.Close = true
	for k, vs := range params.header {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	if params.useragent != "" {
		req.Header.Add("User-Agent", params.useragent)
	}
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-UploadToken", params.UploadToken)
	if params.authorization != "" {
		req.Header.Add("Authorization", params.authorization)
	}

	//
	// Perform request
	res, err := c.http.Do(req)
	if err != nil {
		return "", "", errors.Wrap(err, "could not perform request")
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return "", "", errors.Wrapf(err, "%d", res.StatusCode)
		}
		return "", "", errors.Errorf("[%d] %s", res.StatusCode, body)
	}

	var file struct {
		ID   string `json:"id"`
		Name string `json:"fileName"`
	}

	dec := json.NewDecoder(res.Body)
	return file.ID, file.Name, dec.Decode(&file)
}

func (c *client) streamer(name string, r io.Reader) (string, io.ReadCloser) {
	pR, pW := io.Pipe()
	multipartW := multipart.NewWriter(pW)

	go func() {
		defer pW.Close()
		defer multipartW.Close()

		partW, err := multipartW.CreateFormFile("file", name)
		if err != nil {
			pW.CloseWithError(err)
		}

		_, err = io.Copy(partW, r)
		if err != nil {
			pW.CloseWithError(err)
		}
	}()

	return multipartW.FormDataContentType(), pR
}
