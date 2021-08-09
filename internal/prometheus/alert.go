package prometheus

import (
	"context"
	"regexp"
	"time"

	"github.com/fatih/color"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"golang.org/x/sync/errgroup"
)

const (
	alertNameLabelName = "alertname"
	dlftJobName        = "default"
)

// Alerts is a slice of prometheus alerts.
type Alerts []Alert

// Alert is a prometheus alert with a server.
type Alert struct {
	v1.Alert
}

// Header represents a alert header.
func (a Alert) Header() []string {
	return []string{"SERVER", "JOB", "ALERT_NAME", "SINCE", "STATE"}
}

// Row represents a alert row.
func (a Alert) Row() []string {
	col := color.New(color.FgRed).SprintFunc()

	if a.State == v1.AlertStatePending {
		col = color.New(color.FgYellow).SprintFunc()
	}

	return []string{string(a.Labels["scraper"]), a.Job(), a.Name(), time.Since(a.ActiveAt).String(), col(string(a.State))}
}

// Job returns the job label value.
func (a Alert) Job() string {
	job, ok := a.Labels[jobLabelName]
	if !ok {
		job = dlftJobName
	}

	return string(job)
}

// Name returns the alert name label value.
func (a Alert) Name() string {
	job, ok := a.Labels[alertNameLabelName]
	if !ok {
		job = "-"
	}

	return string(job)
}

// AlertFilterFunc is a function to filter alerts. If function returns true
// service is selected else omitted.
type AlertFilterFunc func(Alert) bool

// Filter filters Alerts with AlertFilterFunc.
func (a Alerts) Filter(filters ...AlertFilterFunc) Alerts {
	alerts := Alerts{}

	for i := range a {
		selectAlert := true
		for _, f := range filters {
			selectAlert = selectAlert && f(a[i])
		}

		if selectAlert {
			alerts = append(alerts, a[i])
		}
	}

	return alerts
}

// AlertByServer filters Alerts by prometheus server.
func AlertByServer(r *regexp.Regexp) AlertFilterFunc {
	return func(a Alert) bool {
		return r.MatchString(string(a.Labels["scraper"]))
	}
}

// AlertByJob filters Alerts by job.
func AlertByJob(r *regexp.Regexp) AlertFilterFunc {
	return func(a Alert) bool {
		return r.MatchString(a.Job())
	}
}

// AlertByName filters Alerts by job.
func AlertByName(r *regexp.Regexp) AlertFilterFunc {
	return func(a Alert) bool {
		return r.MatchString(a.Name())
	}
}

// AlertByState filters Alerts by health.
func AlertByState(state v1.AlertState) AlertFilterFunc {
	return func(a Alert) bool {
		return a.State == state
	}
}

// Alerts returns all alerts.
func (c Client) Alerts(ctx context.Context) (Alerts, error) {
	g, ctx := errgroup.WithContext(ctx)
	results := make(chan alertResult, len(c.clients))

	for server, client := range c.clients {
		client := client // https://golang.org/doc/faq#closures_and_goroutines
		server := server

		g.Go(func() error {
			r, err := client.Alerts(ctx)
			if err != nil {
				return err
			}

			results <- alertResult{
				server: server,
				alert:  r,
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	close(results)

	alerts := Alerts{}

	for r := range results {
		for _, a := range r.alert.Alerts {
			a.Labels["scraper"] = model.LabelValue(r.server)

			if err := a.Labels.Validate(); err != nil {
				return nil, err
			}

			alert := Alert{
				Alert: a,
			}
			alerts = append(alerts, alert)
		}
	}

	return alerts, nil
}

type alertResult struct {
	alert  v1.AlertsResult
	server string
}
