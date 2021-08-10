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
	jobLabelName    = "job"
	sourcesJobName  = "promi_scrape_sources"
	sourceLabelName = "promi_scrape_src"
)

// Targets is a slice of prometheus targets.
type Targets []Target

// Target is prometheus Target.
type Target struct {
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
	col := color.New(color.FgGreen).SprintFunc()

	if health == v1.HealthBad {
		col = color.New(color.FgRed).SprintFunc()
	}

	if health == v1.HealthUnknown {
		col = color.New(color.FgYellow).SprintFunc()
	}

	server := string(t.Labels[sourceLabelName])

	return []string{server, t.Job(), t.ScrapeURL, time.Since(t.LastScrape).String(), t.Labels.String(), t.ActiveTarget.LastError, col(string(t.Health))}
}

// Targets returns all active targets. If appendScraperAsTarget is true the scraper status
// is added to the active targets and all errors accessing scrapers are ignored.
func (c Client) Targets(ctx context.Context, appendScraperAsTarget bool) (Targets, error) {
	g, ctx := errgroup.WithContext(ctx)
	results := make(chan result, len(c.clients))

	for server, client := range c.clients {
		client := client
		server := server

		g.Go(func() error {
			start := time.Now()
			r, err := client.Targets(ctx)
			if appendScraperAsTarget {
				a := v1.ActiveTarget{
					ScrapeURL:  server,
					ScrapePool: sourcesJobName,
					GlobalURL:  server,
					Labels: model.LabelSet{
						"instance": model.LabelValue(server),
					},
					DiscoveredLabels:   map[string]string{},
					LastScrape:         start,
					LastScrapeDuration: time.Since(start).Seconds(),
					Health:             v1.HealthGood,
				}

				if err != nil {
					a.Health = v1.HealthBad
					a.LastError = err.Error()
				}

				r.Active = append(r.Active, a)

				err = nil // ingoring errors if we append scrapers as targets
			}

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
		for i := range r.target.Active {
			activeTarget := r.target.Active[i]
			if activeTarget.ScrapePool != sourcesJobName {
				activeTarget.Labels[sourceLabelName] = model.LabelValue(r.server)
			}

			if err := activeTarget.Labels.Validate(); err != nil {
				return nil, err
			}

			target := Target{
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
		return r.MatchString(string(t.Labels[sourceLabelName]))
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
		switch strings.Compare(string(t[i].Labels[sourceLabelName]), string(t[j].Labels[sourceLabelName])) {
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

// Active returns the active targets.
func (t Targets) Active() []*v1.ActiveTarget {
	active := []*v1.ActiveTarget{}

	for i := range t {
		active = append(active, &t[i].ActiveTarget)
	}

	return active
}

// Deduplicate deduplicates targets by scrape url. If identical targets
// have different health status, then HealthBad status has highest priority.
// Multiple sources are appended in the sourceLabelName label.
func (t Targets) Deduplicate() Targets {
	m := map[string]Target{}

	for i := range t {
		target := t[i]
		if tg, ok := m[target.ScrapeURL]; ok {
			tg.addSource(target.getSource())

			if target.Health == tg.Health || tg.Health == v1.HealthBad {
				target.addSource(tg.getSource())
				continue
			}
		}

		m[target.ScrapeURL] = target
	}

	targets := make(Targets, 0, len(m))

	for s := range m {
		targets = append(targets, m[s])
	}

	return targets
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

func (t Target) addSource(src string) {
	s := string(t.Labels[sourceLabelName])

	if s == "" || src == "" || strings.Contains(s, src) {
		return
	}

	l := []string{s}
	if strings.Contains(s, ",") {
		l = strings.Split(s, ",")
	}

	l = append(l, src)

	sort.Strings(l)

	t.Labels[sourceLabelName] = model.LabelValue(strings.Join(l, ","))
}

func (t Target) getSource() string {
	return string(t.Labels[sourceLabelName])
}
