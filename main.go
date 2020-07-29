package main

import (
	"log"
	"fmt"

	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	"github.com/sensu/sensu-go/types"
)

// Config represents the handler plugin config.
type Config struct {
	sensu.PluginConfig
	URL string
	Job string
}

var (
	plugin = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-prometheus-pushgateway-handler",
			Short:    "Send Sensu Go event metrics to the Prometheus Pushgateway.",
			Keyspace: "sensu.io/plugins/sensu-prometheus-pushgateway-handler/config",
		},
	}

	options = []*sensu.PluginConfigOption{
		&sensu.PluginConfigOption{
			Path:      "url",
			Env:       "PUSHGATEWAY_URL",
			Argument:  "url",
			Shorthand: "u",
			Default:   "http://127.0.0.1:9091/metrics",
			Usage:     "The Prometheus Pushgateway metrics API URL.",
			Value:     &plugin.URL,
		},
		&sensu.PluginConfigOption{
			Path:      "job",
			Env:       "PUSHGATEWAY_JOB",
			Argument:  "job",
			Shorthand: "j",
			Default:   "",
			Usage:     "The Prometheus Pushgateway metrics job name (required).",
			Value:     &plugin.Job,
		},
	}
)

func main() {
	handler := sensu.NewGoHandler(&plugin.PluginConfig, options, checkArgs, executeHandler)
	handler.Execute()
}

func checkArgs(_ *types.Event) error {
	if len(plugin.Job) == 0 {
		return fmt.Errorf("--job or PUSHGATEWAY_JOB environment variable is required")
	}
	return nil
}

func executeHandler(event *types.Event) error {
	log.Println("executing handler with --url", plugin.URL)
	log.Println("executing handler with --job", plugin.Job)
	return nil
}
