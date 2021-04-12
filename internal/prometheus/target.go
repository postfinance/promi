package prometheus

import (
	"context"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	jobLabelName = "job"
)

// Targets is a slice of prometheus targets.
type Targets []Target

// Target is prometheus Target.
type Target struct {
	Server string
	v1.ActiveTarget
}

// Job returns the job label value.
func (t Target) Job() string {
	job, ok := t.Labels[jobLabelName]
	if !ok {
		job = dlftJobName
	}

	return string(job)
}

// Header represents a target header.
func (t Target) Header() []string {
	return []string{"SERVER", "JOB", "SCRAPE_URL", "LAST_SCRAPE", "LABELS", "LAST_ERROR", "HEALTH"}
}

// Row represents a target row.
func (t Target) Row() []string {
	health := t.Health
	col := color.New(color.FgRed).SprintFunc()

	if health == v1.HealthBad {
		col = color.New(color.FgGreen).SprintFunc()
	}

	if health == v1.HealthUnknown {
		col = color.New(color.FgYellow).SprintFunc()
	}

	return []string{t.Server, t.Job(), t.ScrapeURL, time.Since(t.LastScrape).String(), t.Labels.String(), t.ActiveTarget.LastError, col(string(t.Health))}
}

// Targets returns all active targets.
func (c Client) Targets(ctx context.Context) (Targets, error) {
	g, ctx := errgroup.WithContext(ctx)
	results := make(chan result, len(c.clients))

	for server, client := range c.clients {
		client := client // https://golang.org/doc/faq#closures_and_goroutines
		server := server

		g.Go(func() error {
			r, err := client.Targets(ctx)
			if err != nil {
				return err
			}

			results <- result{
				server: server,
				target: r,
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	close(results)

	targets := Targets{}

	for r := range results {
		for _, activeTarget := range r.target.Active {
			target := Target{
				Server:       r.server,
				ActiveTarget: activeTarget,
			}

			targets = append(targets, target)
		}
	}

	return targets, nil
}

// TargetFilterFunc is a function to filter targets. If function returns true
// service is selected else omitted.
type TargetFilterFunc func(Target) bool

// Filter filters Targets with TargetFilterFunc.
func (t Targets) Filter(filters ...TargetFilterFunc) Targets {
	targets := Targets{}

	for i := range t {
		selectTarget := true
		for _, f := range filters {
			selectTarget = selectTarget && f(t[i])
		}

		if selectTarget {
			targets = append(targets, t[i])
		}
	}

	return targets
}

// TargetByServer filters Targets by prometheus server.
func TargetByServer(r *regexp.Regexp) TargetFilterFunc {
	return func(t Target) bool {
		return r.MatchString(t.Server)
	}
}

// TargetByJob filters Targets by job.
func TargetByJob(r *regexp.Regexp) TargetFilterFunc {
	return func(t Target) bool {
		return r.MatchString(t.Job())
	}
}

// TargetByScrapeURL filters Targets by scrape url.
func TargetByScrapeURL(r *regexp.Regexp) TargetFilterFunc {
	return func(t Target) bool {
		return r.MatchString(t.ScrapeURL)
	}
}

// TargetByHealth filters Targets by health.
func TargetByHealth(health v1.HealthStatus) TargetFilterFunc {
	return func(t Target) bool {
		return t.Health == health
	}
}

// TargetBySelector filters Services by Selector.
func TargetBySelector(selector labels.Selector) TargetFilterFunc {
	return func(t Target) bool {
		return selector.Matches(t.labels())
	}
}

// Sort sorts targets by job and scrape url.
func (t Targets) Sort() {
	sort.Slice(t, func(i, j int) bool {
		switch strings.Compare(t[i].Job(), t[j].Job()) {
		case -1:
			return true
		case 1:
			return false
		}
		switch strings.Compare(t[i].Server, t[j].Server) {
		case -1:
			return true
		case 1:
			return false
		}
		return t[i].ScrapeURL < t[j].ScrapeURL
	})
}

// Compact removes labels and last error information.
func (t Targets) Compact() {
	for i := range t {
		t[i].Labels = model.LabelSet{}
		t[i].LastError = ""
	}
}

func (t Target) labels() k8sLabels {
	return k8sLabels{
		LabelSet: t.Labels,
	}
}

type result struct {
	target v1.TargetsResult
	server string
}

type k8sLabels struct {
	model.LabelSet
}

// Has returs true if Labels contains key.
func (l k8sLabels) Has(key string) bool {
	_, ok := l.LabelSet[model.LabelName(key)]
	return ok
}

// Get gets the value for key.
func (l k8sLabels) Get(key string) string {
	return string(l.LabelSet[model.LabelName(key)])
}
