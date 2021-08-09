[![Go Report Card](https://goreportcard.com/badge/github.com/postfinance/promi)](https://goreportcard.com/report/github.com/postfinance/promi)
[![Coverage Status](https://coveralls.io/repos/github/postfinance/promi/badge.svg)](https://coveralls.io/github/postfinance/promi)
[![Build Status](https://github.com/postfinance/promi/workflows/build/badge.svg)](https://github.com/postfinance/promi/actions)

# promi

CLI and Web UI to view targets and alerts of multiple prometheus servers.

# Usage

Create a configuration file with required prometheus servers in `$HOME/.config/promi/config.yaml`:

```yaml
prometheus-urls:
    - http://prometheus101.example.com
    - http://prometheus102.example.com
    - http://prometheus103.example.com
```

Or you can create an environment variable:

```console
export PROMI_PROMETHEUS_URLS=http://prometheus101.example.com,http://prometheus102.example.com,http://prometheus103.example.com
```

## Web UI
If you run the command:

```console
$ promi server
```

and point your browser to http://localhost:8080 you get the official prometheus web ui with all consolidated targets. The source code
is taken from prometheus [react-app](https://github.com/prometheus/prometheus/tree/main/web/ui/react-app). Only the targets endpoint work and
the classic ui is omitted.

If multiple prometheus servers scrape the same endpoint you can run the server with the option `--deduplicate`.

## CLI
To list all targets run:

```console
$ promi targets
```

```console
$ promi targets --help
Usage: promi targets

Show targets.

Flags:
  -h, --help                                         Show context-sensitive help ($PROMI_HELP).
  -u, --prometheus-urls=http://localhost:9090,...    A comma separated list of prometheus base URLs ($PROMI_PROMETHEUS_URLS).
      --show-config                                  Show used config files ($PROMI_SHOW_CONFIG)
      --version                                      Show version information ($PROMI_VERSION)
  -d, --debug                                        Show debug output ($PROMI_DEBUG).
      --timeout=10s                                  The http request timeout ($PROMI_TIMEOUT).

  -o, --output="table"                               Output format (table|json|yaml) ($PROMI_OUTPUT).
  -c, --compact                                      Do not display labels and last error ($PROMI_COMPACT).
  -n, --no-headers                                   Do not display headers in table output ($PROMI_NO_HEADERS).
  -N, --filter-name=STRING                           Filter targets by job name (regular expression) ($PROMI_FILTER_NAME).
  -S, --filter-server=STRING                         Filter targets by promehteus server name (regular expression) ($PROMI_FILTER_SERVER).
  -u, --filter-scrape-url=STRING                     Filter targets by scrape url (regular expression) ($PROMI_FILTER_SCRAPE_URL).
  -H, --filter-health=HEALTH-STATUS                  Filter targets by health (up|down) ($PROMI_FILTER_HEALTH)
  -s, --filter-selector=STRING                       Filter services by (k8s style) selector ($PROMI_FILTER_SELECTOR).
```

To list all alerts run:

```console
$ promi alerts
```

```console
$ promi alerts --help
Usage: promi alerts

Show alerts.

Flags:
  -h, --help                                         Show context-sensitive help ($PROMI_HELP).
  -u, --prometheus-urls=http://localhost:9090,...    A comma separated list of prometheus base URLs ($PROMI_PROMETHEUS_URLS).
      --show-config                                  Show used config files ($PROMI_SHOW_CONFIG)
      --version                                      Show version information ($PROMI_VERSION)
  -d, --debug                                        Show debug output ($PROMI_DEBUG).
      --timeout=10s                                  The http request timeout ($PROMI_TIMEOUT).

  -o, --output="table"                               Output format (table|json|yaml) ($PROMI_OUTPUT).
  -n, --no-headers                                   Do not display headers in table output ($PROMI_NO_HEADERS).
  -N, --filter-name=STRING                           Filter alerts by job name (regular expression) ($PROMI_FILTER_NAME).
  -a, --filter-alert=STRING                          Filter alerts by alert name (regular expression) ($PROMI_FILTER_ALERT).
  -S, --filter-server=STRING                         Filter alerts by prometheus server name (regular expression) ($PROMI_FILTER_SERVER).
  -s, --filter-state=ALERT-STATE                     Filter alerts by state (pending|firing) ($PROMI_FILTER_STATE)
```
