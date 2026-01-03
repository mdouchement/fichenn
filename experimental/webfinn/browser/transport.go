package main

import (
	"net/http"
	"net/url"
)

type Transport struct {
	proxy url.URL
	super *http.Transport
}

func NewTransport(proxy url.URL) (*Transport, error) {
	return &Transport{
		proxy: url.URL{
			Scheme: proxy.Scheme,
			Host:   proxy.Host,
			Path:   "/proxy",
		},
		super: &http.Transport{},
	}, nil
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("X-Finn-Destination", req.URL.String())
	req.URL = &t.proxy

	// payload, _ := httputil.DumpRequest(req, false)
	// fmt.Println(string(payload))
	return t.super.RoundTrip(req)
}
