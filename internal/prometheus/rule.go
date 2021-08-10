package prometheus

import (
	"context"
	"strings"
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"golang.org/x/sync/errgroup"
)

// Rules is a slice of prometheus rules.
type Rules []Rule

// Alerting returns all alerting rules.
func (r Rules) Alerting() []*v1.RuleGroup {
	rules := []*v1.RuleGroup{}

	for i := range r {
		rules = append(rules, &r[i].RuleGroup)
	}

	return rules

	/*
		for i := range r {
			alertRules := []interface{}{}

			for j := range r[i].RuleGroup.Rules {
				if a, ok := r[i].RuleGroup.Rules[j].(v1.AlertingRule); ok {
					alertRules = append(alertRules, a)
				}
			}

			if len(alertRules) > 0 {
				rg := r[i].RuleGroup
				rg.Rules = alertRules

				rules = append(rules, &rg)
			}
		}

		return rules
	*/
}

// Rule is a prometheus roule group.
type Rule struct {
	v1.RuleGroup
}

// Rules returns all rules.
func (c Client) Rules(ctx context.Context) (Rules, error) {
	g, ctx := errgroup.WithContext(ctx)
	results := make(chan v1.RulesResult, len(c.clients))

	for server, client := range c.clients {
		client := client
		server := server
		srvName := strings.Split(server, ":")[0]

		g.Go(func() error {
			r, err := client.Rules(ctx)
			if err != nil {
				ng := v1.RuleGroup{
					Name:     srvName + "-promi.rules",
					File:     "promy.yaml",
					Interval: time.Second.Seconds(),
					Rules: v1.Rules{
						v1.AlertingRule{
							Name:     "PromiSourceUnreachable",
							Query:    "na",
							Duration: time.Second.Seconds(),
							Health:   v1.RuleHealthBad,
						},
					},
				}

				results <- v1.RulesResult{
					Groups: []v1.RuleGroup{ng},
				}
				return nil
			}

			/*
				for i := range r.Groups {
					for j := range r.Groups[i].Rules {
						if alertingRule, ok := r.Groups[i].Rules[j].(v1.AlertingRule); ok {
							alertingRule.Labels = model.LabelSet{
								sourceLabelName: model.LabelValue(server),
							}
						}
						if recordingRule, ok := r.Groups[i].Rules[j].(v1.RecordingRule); ok {
							recordingRule.Labels = model.LabelSet{
								sourceLabelName: model.LabelValue(server),
							}
						}
					}
				}
			*/
			results <- r

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	close(results)

	rules := Rules{}

	for r := range results {
		for _, g := range r.Groups {
			rule := Rule{
				RuleGroup: g,
			}
			rules = append(rules, rule)
		}
	}

	return rules, nil
}
