[![Go Report Card](https://goreportcard.com/badge/github.com/postfinance/promcli)](https://goreportcard.com/report/github.com/postfinance/promcli)
[![Coverage Status](https://coveralls.io/repos/github/postfinance/promcli/badge.svg)](https://coveralls.io/github/postfinance/promcli)
[![Build Status](https://github.com/postfinance/promcli/workflows/build/badge.svg)](https://github.com/postfinance/promcli/actions)

# promcli

CLI to query targets and alerts of multiple prometheus servers.

# Usage

Create a configuration file with required prometheus servers in `$HOME/.config/promcli/config.yaml`:

```yaml
prometheus-urls:
    - http://prometheus101.example.com
    - http://prometheus102.example.com
    - http://prometheus103.example.com
```

To list all targets run:

```console
$ promcli targets
```

```console
$ promcli targets --help
Usage: promcli targets

Show targets.

Flags:
  -h, --help                                         Show context-sensitive help ($PROMCLI_HELP).
  -u, --prometheus-urls=http://localhost:9090,...    A comma separated list of prometheus base URLs ($PROMCLI_PROMETHEUS_URLS).
      --show-config                                  Show used config files ($PROMCLI_SHOW_CONFIG)
      --version                                      Show version information ($PROMCLI_VERSION)
  -d, --debug                                        Show debug output ($PROMCLI_DEBUG).
      --timeout=10s                                  The http request timeout ($PROMCLI_TIMEOUT).

  -o, --output="table"                               Output format (table|json|yaml) ($PROMCLI_OUTPUT).
  -c, --compact                                      Do not display labels and last error ($PROMCLI_COMPACT).
  -n, --no-headers                                   Do not display headers in table output ($PROMCLI_NO_HEADERS).
  -N, --filter-name=STRING                           Filter targets by job name (regular expression) ($PROMCLI_FILTER_NAME).
  -S, --filter-server=STRING                         Filter targets by promehteus server name (regular expression) ($PROMCLI_FILTER_SERVER).
  -u, --filter-scrape-url=STRING                     Filter targets by scrape url (regular expression) ($PROMCLI_FILTER_SCRAPE_URL).
  -H, --filter-health=HEALTH-STATUS                  Filter targets by health (up|down) ($PROMCLI_FILTER_HEALTH)
  -s, --filter-selector=STRING                       Filter services by (k8s style) selector ($PROMCLI_FILTER_SELECTOR).
```

To list all alerts run:

```console
$ promcli alerts
```

```console
$ promcli alerts --help
Usage: promcli alerts

Show alerts.

Flags:
  -h, --help                                         Show context-sensitive help ($PROMCLI_HELP).
  -u, --prometheus-urls=http://localhost:9090,...    A comma separated list of prometheus base URLs ($PROMCLI_PROMETHEUS_URLS).
      --show-config                                  Show used config files ($PROMCLI_SHOW_CONFIG)
      --version                                      Show version information ($PROMCLI_VERSION)
  -d, --debug                                        Show debug output ($PROMCLI_DEBUG).
      --timeout=10s                                  The http request timeout ($PROMCLI_TIMEOUT).

  -o, --output="table"                               Output format (table|json|yaml) ($PROMCLI_OUTPUT).
  -n, --no-headers                                   Do not display headers in table output ($PROMCLI_NO_HEADERS).
  -N, --filter-name=STRING                           Filter alerts by job name (regular expression) ($PROMCLI_FILTER_NAME).
  -a, --filter-alert=STRING                          Filter alerts by alert name (regular expression) ($PROMCLI_FILTER_ALERT).
  -S, --filter-server=STRING                         Filter alerts by prometheus server name (regular expression) ($PROMCLI_FILTER_SERVER).
  -s, --filter-state=ALERT-STATE                     Filter alerts by state (pending|firing) ($PROMCLI_FILTER_STATE)
```
