[![Sensu Bonsai Asset](https://img.shields.io/badge/Bonsai-Download%20Me-brightgreen.svg?colorB=89C967&logo=sensu)](https://bonsai.sensu.io/assets/portertech/sensu-prometheus-pushgateway-handler)
![Go Test](https://github.com/portertech/sensu-prometheus-pushgateway-handler/workflows/Go%20Test/badge.svg)
![goreleaser](https://github.com/portertech/sensu-prometheus-pushgateway-handler/workflows/goreleaser/badge.svg)

# Prometheus Pushgateway Handler

## Table of Contents
- [Overview](#overview)
- [Files](#files)
- [Usage examples](#usage-examples)
- [Configuration](#configuration)
  - [Asset registration](#asset-registration)
  - [Handler definition](#handler-definition)
- [Installation from source](#installation-from-source)
- [Additional notes](#additional-notes)
- [Contributing](#contributing)
- [Release](#releases-with-github-actions)

## Overview

Push Sensu Go event metrics to a Prometheus Pushgateway. The
Pushgateway can then be scraped by Prometheus. This handler allows
users to collect metrics via several means, including 20 year old
Nagios plugins with perfdata, and store them in Prometheus.

This handler plugin writes Sensu Go event metrics to the Prometheus
[Pushgateway
API](https://github.com/prometheus/pushgateway#use-it). The plugin
uses the Golang Prometheus client to format and push the
metrics to a configured job (configured via a plugin CLI
argument). Metrics are expected to already have an `instance`
label. Prometheus must have `honor_labels: true` in the scrape config
for the Pushgateway.

## Files

## Usage examples

```
sensu-prometheus-pushgateway-handler -u http://pushgateway.example.org:9091/metrics -j node
```

## Configuration

### Asset registration

[Sensu Assets][10] are the best way to make use of this plugin. If you're not using an asset, please
consider doing so! If you're using sensuctl 5.13 with Sensu Backend 5.13 or later, you can use the
following command to add the asset:

```
sensuctl asset add portertech/sensu-prometheus-pushgateway-handler
```

If you're using an earlier version of sensuctl, you can find the asset on the [Bonsai Asset Index][https://bonsai.sensu.io/assets/portertech/sensu-prometheus-pushgateway-handler].

### Handler definition

```yml
---
type: Handler
api_version: core/v2
metadata:
  name: prometheus-pushgateway-handler
  namespace: default
spec:
  command: sensu-prometheus-pushgateway-handler -u http://pushgateway.example.org:9091/metrics -j node
  type: pipe
  runtime_assets:
  - portertech/sensu-prometheus-pushgateway-handler
```

#### Proxy Support

This handler supports the use of the environment variables HTTP_PROXY,
HTTPS_PROXY, and NO_PROXY (or the lowercase versions thereof). HTTPS_PROXY takes
precedence over HTTP_PROXY for https requests.  The environment values may be
either a complete URL or a "host[:port]", in which case the "http" scheme is assumed.

## Installation from source

The preferred way of installing and deploying this plugin is to use it as an Asset. If you would
like to compile and install the plugin from source or contribute to it, download the latest version
or create an executable script from this source.

From the local path of the sensu-prometheus-pushgateway-handler repository:

```
go build
```

## Additional notes

## Contributing

For more information about contributing to this plugin, see [Contributing][1].

## Releases with Github Actions

To release a version of your project, simply tag the target sha with a semver release without a `v`
prefix (ex. `1.0.0`). This will trigger the [GitHub action][5] workflow to [build and release][4]
the plugin with goreleaser. Register the asset with [Bonsai][8] to share it with the community!

[1]: https://github.com/sensu/sensu-go/blob/master/CONTRIBUTING.md
[2]: https://github.com/sensu-community/sensu-plugin-sdk
[3]: https://github.com/sensu-plugins/community/blob/master/PLUGIN_STYLEGUIDE.md
[4]: https://github.com/sensu-community/handler-plugin-template/blob/master/.github/workflows/release.yml
[5]: https://github.com/sensu-community/handler-plugin-template/actions
[6]: https://docs.sensu.io/sensu-go/latest/reference/handlers/
[7]: https://github.com/sensu-community/handler-plugin-template/blob/master/main.go
[8]: https://bonsai.sensu.io/
[9]: https://github.com/sensu-community/sensu-plugin-tool
[10]: https://docs.sensu.io/sensu-go/latest/reference/assets/
