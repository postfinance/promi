package prometheus

import (
	"fmt"
	"net/url"
	"testing"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeduplicate(t *testing.T) {
	targets := Targets{
		Target{
			ActiveTarget: v1.ActiveTarget{
				ScrapeURL: "http://url1.example.com",
				Health:    v1.HealthGood,
				Labels: model.LabelSet{
					sourceLabelName: "src1",
				},
			},
		},
		Target{
			ActiveTarget: v1.ActiveTarget{
				ScrapeURL: "http://url1.example.com",
				Health:    v1.HealthGood,
				Labels: model.LabelSet{
					sourceLabelName: "src2",
				},
			},
		},
		Target{
			ActiveTarget: v1.ActiveTarget{
				ScrapeURL: "http://url2.example.com",
				Health:    v1.HealthBad,
				LastError: "scrapeErr",
				Labels: model.LabelSet{
					sourceLabelName: "src1",
				},
			},
		},
		Target{
			ActiveTarget: v1.ActiveTarget{
				ScrapeURL: "http://url2.example.com",
				Health:    v1.HealthGood,
				Labels: model.LabelSet{
					sourceLabelName: "src2",
				},
			},
		},
		Target{
			ActiveTarget: v1.ActiveTarget{
				ScrapeURL: "http://url3.example.com",
				Health:    v1.HealthBad,
				LastError: "scrapeErr",
				Labels: model.LabelSet{
					sourceLabelName: "src1",
				},
			},
		},
		Target{
			ActiveTarget: v1.ActiveTarget{
				ScrapeURL: "http://url3.example.com",
				Health:    v1.HealthGood,
				Labels: model.LabelSet{
					sourceLabelName: "src2",
				},
			},
		},
	}

	dedup := targets.Deduplicate()
	assert.NoError(t, dedup[0].Labels.Validate())

	dedup.Sort()

	expected := Targets{
		Target{
			ActiveTarget: v1.ActiveTarget{
				ScrapeURL: "http://url1.example.com",
				Health:    v1.HealthGood,
				Labels: model.LabelSet{
					sourceLabelName: "src1,src2",
				},
			},
		},
		Target{
			ActiveTarget: v1.ActiveTarget{
				ScrapeURL: "http://url2.example.com",
				Health:    v1.HealthBad,
				LastError: "scrapeErr",
				Labels: model.LabelSet{
					sourceLabelName: "src1,src2",
				},
			},
		},
		Target{
			ActiveTarget: v1.ActiveTarget{
				ScrapeURL: "http://url3.example.com",
				Health:    v1.HealthBad,
				LastError: "scrapeErr",
				Labels: model.LabelSet{
					sourceLabelName: "src1,src2",
				},
			},
		},
	}

	assert.Equal(t, expected, dedup)

}

func TestRene(t *testing.T) {
	s := "www.reen.ch:7002"

	u, err := url.ParseRequestURI(s)
	require.NoError(t, err)

	fmt.Println(u.Hostname())
	fmt.Println(u.Host)
	fmt.Printf("%#v", u)

}
