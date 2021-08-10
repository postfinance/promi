package web

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/postfinance/promi/internal/prometheus"
	"github.com/postfinance/promi/internal/web/ui"
	"go.uber.org/zap"
)

// API represents the server API.
type API struct {
	router        chi.Router
	reactApp      http.FileSystem
	urlPathPrefix string
	listenAddr    string
	deduplicate   bool
	l             *zap.SugaredLogger
	cli           *prometheus.Client
	timeout       time.Duration
}

// New initializes the API.
func New(l *zap.SugaredLogger, timeout time.Duration, deduplicate bool, urls ...string) (*API, error) {
	client, err := prometheus.New(urls...)
	if err != nil {
		return nil, err
	}

	react, err := ui.ReactApp()
	if err != nil {
		return nil, err
	}

	a := API{
		l:             l,
		timeout:       timeout,
		deduplicate:   deduplicate,
		reactApp:      react,
		listenAddr:    ":8080",
		urlPathPrefix: "/",
		cli:           client,
	}

	if !strings.HasPrefix(a.urlPathPrefix, "/") {
		return nil, errors.New("url prefix must start with '/'")
	}

	r := chi.NewRouter()
	a.router = r

	return &a, nil
}

// Start starts the server.
func (a *API) Start() error {
	if err := a.routes(); err != nil {
		return err
	}

	httpSrv := &http.Server{
		Addr:    ":8080",
		Handler: a.router,
	}

	return httpSrv.ListenAndServe()
}

type status string

const (
	statusSuccess status = "success"
	statusError   status = "error"
)

type errorType string

const (
	errorNone        errorType = ""
	errorTimeout     errorType = "timeout"
	errorCanceled    errorType = "canceled"
	errorExec        errorType = "execution"
	errorBadData     errorType = "bad_data"
	errorInternal    errorType = "internal"
	errorUnavailable errorType = "unavailable"
	errorNotFound    errorType = "not_found"
)

type response struct {
	Status    status      `json:"status"`
	Data      interface{} `json:"data,omitempty"`
	ErrorType errorType   `json:"errorType,omitempty"`
	Error     string      `json:"error,omitempty"`
	Warnings  []string    `json:"warnings,omitempty"`
}
