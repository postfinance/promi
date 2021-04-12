package prometheus

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTarget(t *testing.T) {
	s1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, targets1)
	}))
	s2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, targets2)
	}))

	cli, err := New(s1.URL, s2.URL)
	require.NoError(t, err)
	targets, err := cli.Targets(context.Background())
	require.NoError(t, err)

	t.Run("the number of targets must be 4", func(t *testing.T) {
		assert.Len(t, targets, 4)
	})

	t.Run("the number of returned prometheus servers must be 2", func(t *testing.T) {
		m := map[string]bool{}

		for _, target := range targets {
			m[target.Server] = true
		}

		assert.Len(t, m, 2)
	})

	t.Run("the filtered targets must contain exactly one target", func(t *testing.T) {
		filters := []TargetFilterFunc{
			TargetByServer(regexp.MustCompile("127.0.0.1" + ":" + port(t, s1.URL))),
			TargetByJob(regexp.MustCompile("jobname$")),
			TargetByHealth("up"),
		}

		targets = targets.Filter(filters...)
		assert.Len(t, targets, 1)
	})
}

func TestAlert(t *testing.T) {
	s1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, alerts1)
	}))
	s2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, alerts2)
	}))

	cli, err := New(s1.URL, s2.URL)
	require.NoError(t, err)
	alerts, err := cli.Alerts(context.Background())
	require.NoError(t, err)

	t.Run("the number of alerts must be 4", func(t *testing.T) {
		assert.Len(t, alerts, 2)
	})

	t.Run("the number of returned prometheus servers must be 2", func(t *testing.T) {
		m := map[string]bool{}

		for _, target := range alerts {
			m[target.Server] = true
		}

		assert.Len(t, m, 2)
	})

	t.Run("the filtered alerts must contain exactly one target", func(t *testing.T) {
		filters := []AlertFilterFunc{
			AlertByState("firing"),
		}

		alerts = alerts.Filter(filters...)
		assert.Len(t, alerts, 1)
	})
}

func port(t *testing.T, u string) string {
	parsed, err := url.Parse(u)
	require.NoError(t, err)

	return parsed.Port()
}

var (
	targets1 = `
{
  "status": "success",
  "data": {
    "activeTargets": [
      {
        "discoveredLabels": {
          "__address__": "example101.com",
          "__metrics_path__": "/metrics",
          "__scheme__": "http",
          "job": "jobname"
        },
        "labels": {
          "instance": "example101",
          "job": "jobname"
        },
        "scrapePool": "jobname",
        "scrapeUrl": "http://example101.com/metrics",
        "globalUrl": "http://example101.com/metrics",
        "lastError": "",
        "lastScrape": "2021-04-12T08:41:32.051367968+02:00",
        "lastScrapeDuration": 0.003451212,
        "health": "up"
      },
      {
        "discoveredLabels": {
          "__address__": "example102.com",
          "__metrics_path__": "/metrics",
          "__scheme__": "http",
          "job": "jobname"
        },
        "labels": {
          "instance": "example102",
          "job": "jobname"
        },
        "scrapePool": "jobname",
        "scrapeUrl": "http://example102.com/metrics",
        "globalUrl": "http://example102.com/metrics",
        "lastError": "",
        "lastScrape": "2021-04-12T08:41:32.051367968+02:00",
        "lastScrapeDuration": 0.003451212,
        "health": "down"
      },
      {
        "discoveredLabels": {
          "__address__": "example103.com",
          "__metrics_path__": "/metrics",
          "__scheme__": "http",
          "job": "jobname2"
        },
        "labels": {
          "instance": "example103",
          "job": "jobname2"
        },
        "scrapePool": "jobname2",
        "scrapeUrl": "http://example103.com/metrics",
        "globalUrl": "http://example103.com/metrics",
        "lastError": "",
        "lastScrape": "2021-04-12T08:41:32.051367968+02:00",
        "lastScrapeDuration": 0.003451212,
        "health": "up"
      }
    ]
  }
}
`
	targets2 = `
{
  "status": "success",
  "data": {
    "activeTargets": [
      {
        "discoveredLabels": {
          "__address__": "example101.com",
          "__metrics_path__": "/metrics",
          "__scheme__": "http",
          "job": "jobname"
        },
        "labels": {
          "instance": "example101",
          "job": "jobname"
        },
        "scrapePool": "jobname",
        "scrapeUrl": "http://example101.com/metrics",
        "globalUrl": "http://example101.com/metrics",
        "lastError": "",
        "lastScrape": "2021-04-12T08:41:32.051367968+02:00",
        "lastScrapeDuration": 0.003451212,
        "health": "up"
      }
    ]
  }
}
`

	alerts1 = `
{
  "status": "success",
  "data": {
    "alerts": [
      {
        "labels": {
          "alertname": "PrometheusNotRunning",
          "instance": "example101",
          "job": "prometheus",
          "severity": "major"
        },
        "annotations": {
          "description": "Check service with systemctl status prometheus",
          "summary": "Prometheus server is not running on example101"
        },
        "state": "pending",
        "activeAt": "2021-04-12T08:14:19.574495804Z",
        "value": "0e+00"
      }
    ]
  }
}
`
	alerts2 = `
{
  "status": "success",
  "data": {
    "alerts": [
      {
        "labels": {
          "alertname": "PrometheusNotRunning",
          "instance": "example102",
          "job": "prometheus",
          "severity": "major"
        },
        "annotations": {
          "description": "Check service with systemctl status prometheus",
          "summary": "Prometheus server is not running on example102"
        },
        "state": "firing",
        "activeAt": "2021-04-12T08:14:19.574495804Z",
        "value": "0e+00"
      }
    ]
  }
}
`
)
