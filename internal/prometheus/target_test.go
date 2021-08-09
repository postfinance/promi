package prometheus

import (
	"testing"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/stretchr/testify/assert"
)

func TestDeduplicate(t *testing.T) {
	targets := Targets{
		Target{
			ActiveTarget: v1.ActiveTarget{
				ScrapeURL: "http://url1.example.com",
				Health:    v1.HealthGood,
			},
		},
		Target{
			ActiveTarget: v1.ActiveTarget{
				ScrapeURL: "http://url1.example.com",
				Health:    v1.HealthGood,
			},
		},
		Target{
			ActiveTarget: v1.ActiveTarget{
				ScrapeURL: "http://url2.example.com",
				Health:    v1.HealthBad,
				LastError: "scrapeErr",
			},
		},
		Target{
			ActiveTarget: v1.ActiveTarget{
				ScrapeURL: "http://url2.example.com",
				Health:    v1.HealthGood,
			},
		},
		Target{
			ActiveTarget: v1.ActiveTarget{
				ScrapeURL: "http://url3.example.com",
				Health:    v1.HealthBad,
				LastError: "scrapeErr",
			},
		},
		Target{
			ActiveTarget: v1.ActiveTarget{
				ScrapeURL: "http://url3.example.com",
				Health:    v1.HealthGood,
			},
		},
	}

	dedup := targets.Deduplicate()

	dedup.Sort()

	expected := Targets{
		Target{
			ActiveTarget: v1.ActiveTarget{
				ScrapeURL: "http://url1.example.com",
				Health:    v1.HealthGood,
			},
		},
		Target{
			ActiveTarget: v1.ActiveTarget{
				ScrapeURL: "http://url2.example.com",
				Health:    v1.HealthBad,
				LastError: "scrapeErr",
			},
		},
		Target{
			ActiveTarget: v1.ActiveTarget{
				ScrapeURL: "http://url3.example.com",
				Health:    v1.HealthBad,
				LastError: "scrapeErr",
			},
		},
	}

	assert.Equal(t, expected, dedup)

}
