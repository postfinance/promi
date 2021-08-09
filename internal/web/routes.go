package web

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"path"

	"github.com/postfinance/promi/internal/web/fileserver"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

func (a *API) routes() error {
	a.router.Get("/", redirectToTargets)
	a.router.Get("/classic/graph", redirectToTargets)
	a.router.Get("/graph", redirectToTargets)

	a.router.Get(path.Join(a.urlPathPrefix, "/api/v1/targets"), a.targets)
	a.router.Get(path.Join(a.urlPathPrefix, "/-/ready"), a.ready)

	fileserver.FileServer(a.router, "/targets", a.reactApp)

	return nil
}

func redirectToTargets(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/targets", http.StatusFound)
}

func (a *API) targets(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
	defer cancel()

	targets, err := a.cli.Targets(ctx, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if a.deduplicate {
		targets = targets.Deduplicate()
	}
	targets.Sort()

	enc := json.NewEncoder(w)
	resp := struct {
		Status    string      `json:"status"`
		Data      interface{} `json:"data,omitempty"`
		ErrorType string      `json:"errorType,omitempty"`
		Error     string      `json:"error,omitempty"`
		Warnings  []string    `json:"warnings,omitempty"`
	}{
		Status: statusSuccess,
		Data: struct {
			ActiveTargets  []*v1.ActiveTarget  `json:"activeTargets"`
			DroppedTargets []*v1.DroppedTarget `json:"droppedTargets"`
		}{
			ActiveTargets:  targets.Active(),
			DroppedTargets: []*v1.DroppedTarget{},
		},
	}
	if err := enc.Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func (a *API) ready(w http.ResponseWriter, r *http.Request) {
	_, _ = io.WriteString(w, "Prometheus is Ready.")
}

const (
	statusSuccess = "success"
)

/*
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
*/
