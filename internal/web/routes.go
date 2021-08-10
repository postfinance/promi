package web

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/postfinance/promi/internal/web/fileserver"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

func (a *API) routes() error {
	a.router.Get("/", redirectToTargets)
	a.router.Get("/classic/graph", redirectToTargets)
	a.router.Get("/graph", redirectToTargets)

	a.router.Get(path.Join(a.urlPathPrefix, "/api/v1/targets"), a.targets)
	a.router.Get(path.Join(a.urlPathPrefix, "/api/v1/rules"), a.rules)
	a.router.Get(path.Join(a.urlPathPrefix, "/-/ready"), a.ready)

	fileserver.FileServer(a.router, "/targets", a.reactApp)
	fileserver.FileServer(a.router, "/alerts", a.reactApp)

	return nil
}

func redirectToTargets(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/targets", http.StatusFound)
}

func (a *API) targets(w http.ResponseWriter, r *http.Request) {
	resp := response{}

	ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
	defer cancel()

	targets, err := a.cli.Targets(ctx, true)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp.Status = statusError
		resp.Error = err.Error()
		resp.ErrorType = errorInternal
	} else {
		if a.deduplicate {
			targets = targets.Deduplicate()
		}

		targets.Sort()

		resp.Status = statusSuccess
		resp.Data = struct {
			ActiveTargets  []*v1.ActiveTarget  `json:"activeTargets"`
			DroppedTargets []*v1.DroppedTarget `json:"droppedTargets"`
		}{
			ActiveTargets:  targets.Active(),
			DroppedTargets: []*v1.DroppedTarget{},
		}
	}

	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func (a *API) rules(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	typ := strings.ToLower(r.URL.Query().Get("type"))
	if typ != "alert" {
		http.Error(w, "only type=alert query paramater supported", http.StatusBadGateway)

		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
	defer cancel()

	rules, err := a.cli.Rules(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp.Status = statusError
		resp.Error = err.Error()
		resp.ErrorType = errorInternal
	} else {
		resp.Status = statusSuccess
		resp.Data = struct {
			Groups []*v1.RuleGroup `json:"groups"`
		}{
			Groups: rules.Alerting(),
		}
	}
	enc := json.NewEncoder(w)

	if err := enc.Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func (a *API) ready(w http.ResponseWriter, r *http.Request) {
	_, _ = io.WriteString(w, "Prometheus is Ready.")
}
