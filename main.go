package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

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

func transformMetrics(event *types.Event) string {
	info := map[string]string{}
	p := ""
	for _, m := range event.Metrics.Points {
		mt := "gauge"
		lt := ""
		for _, t := range m.Tags {
			if t.Name == "prometheus_type" {
				mt = t.Value
				continue
			}
			if lt != "" {
				lt = lt + ","
			}
			lt = lt + fmt.Sprintf("%s=\"%s\"", t.Name, t.Value)
		}
		n := strings.Replace(m.Name, ".", "_", -1)
		if _, ok := info[n]; !ok {
			info[n] = mt
		}
		l := n
		if lt != "" {
			l = l + fmt.Sprintf("{%s}", lt)
		}
		p = p + fmt.Sprintf("%s %v\n", l, m.Value)
	}
	for n, t := range info {
		p = fmt.Sprintf("# TYPE %s %s\n", n, t) + p
	}
	log.Println(p)
	return p
}

func postMetrics(m string) error {
	url := fmt.Sprintf("%s/job/%s", plugin.URL, plugin.Job)
	resp, err := http.Post(url, "text/plain", bytes.NewBuffer([]byte(m)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Println(string(b))
	return nil
}

func executeHandler(event *types.Event) error {
	log.Println("executing handler with --url", plugin.URL)
	log.Println("executing handler with --job", plugin.Job)
	m := transformMetrics(event)
	err := postMetrics(m)
	return err
}
