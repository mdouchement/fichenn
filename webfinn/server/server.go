package server

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"regexp"

	"github.com/gofrs/uuid"
	"github.com/mdouchement/logger"
)

type (
	Route struct {
		pattern *regexp.Regexp
		handler http.HandlerFunc
	}

	Application struct {
		ctx        context.Context
		log        logger.Logger
		filter     *Filter
		fs         fs.FS
		filesystem http.Handler
		routes     []Route
	}
)

func NewApplication(l logger.Logger, filter *Filter, fs fs.FS) *Application {
	app := &Application{
		ctx:        context.Background(),
		log:        l,
		filter:     filter,
		fs:         fs,
		filesystem: http.FileServer(http.FS(fs)),
	}
	app.Handle("/download/*", app.download)
	app.Handle("/proxy", app.proxy)

	return app
}

func (a *Application) Handle(pattern string, handler http.HandlerFunc) {
	a.routes = append(a.routes, Route{
		pattern: regexp.MustCompile(pattern),
		handler: handler,
	})
}

func (a *Application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, route := range a.routes {
		if matches := route.pattern.FindStringSubmatch(r.URL.Path); len(matches) > 0 {
			route.handler(w, r)
			return
		}
	}

	a.filesystem.ServeHTTP(w, r)
}

//

func (a *Application) download(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	f, err := a.fs.Open("index.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	_, err = io.Copy(w, f)
	if err != nil {
		a.log.WithPrefix("[download]").WithError(err).Error("Could not write response")
	}
}

func (a *Application) proxy(w http.ResponseWriter, r *http.Request) {
	requestID := uuid.Must(uuid.NewV4()).String()
	l := a.log.WithPrefix("[proxy]").WithField("request_id", requestID)

	// {
	// 	payload, err := httputil.DumpRequest(r, false)
	// 	if err != nil {
	// 		w.WriteHeader(http.StatusInternalServerError)
	// 		return
	// 	}
	// 	fmt.Println(string(payload))
	// }

	dst := r.Header.Get("X-Finn-Destination")
	if dst == "" {
		m := "Missing X-Finn-Destination"
		l.Info(m)

		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, m)
		return
	}

	if !a.filter.Match(r.Method, dst) {
		m := fmt.Sprintf("Unsupported X-Finn-Destination: %s %s", r.Method, dst)
		l.Info(m)

		w.WriteHeader(http.StatusUnprocessableEntity)
		io.WriteString(w, m)
		return
	}

	req, err := http.NewRequest(r.Method, dst, r.Body)
	if err != nil {
		l.WithError(err).Error("Could not build request")

		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, fmt.Sprintf("Internal Server Error (request_id: %s", requestID))
		return
	}
	defer req.Body.Close()
	for k, vs := range r.Header {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}

	req.Header.Del("X-Finn-Destination")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		l.WithError(err).Error("Could not perform request")

		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, fmt.Sprintf("Internal Server Error (request_id: %s", requestID))
		return
	}
	defer res.Body.Close()

	// {
	// 	payload, err := httputil.DumpResponse(res, false)
	// 	if err != nil {
	// 		w.WriteHeader(http.StatusInternalServerError)
	// 		log.Println(err.Error())
	// 		return
	// 	}
	// 	fmt.Println(string(payload))
	// }

	w.WriteHeader(res.StatusCode)
	for k, vs := range res.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}

	_, err = io.Copy(w, res.Body)
	if err != nil {
		l.WithError(err).Error("Could not write response")
	}
}
