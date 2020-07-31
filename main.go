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
	}
)

func main() {
	handler := sensu.NewGoHandler(&plugin.PluginConfig, options, checkArgs, executeHandler)
	handler.Execute()
}

func checkArgs(_ *types.Event) error {
	return nil
}

func transformMetrics(event *types.Event) (string, string, string) {
	job := ""
	inst := ""
	info := map[string]string{}
	prom := map[string]string{}
	for _, m := range event.Metrics.Points {
		mt := "untyped"
		lt := ""
		for _, t := range m.Tags {
			switch t.Name {
			case "prom_job":
				if job == "" {
					job = t.Value
				}
			case "prom_instance":
				if inst == "" {
					inst = t.Value
				}
			case "prom_type":
				mt = t.Value
			default:
				if lt != "" {
					lt = lt + ","
				}
				lt = lt + fmt.Sprintf("%s=\"%s\"", t.Name, t.Value)
			}
		}
		l := strings.Replace(m.Name, ".", "_", -1)
		if lt != "" {
			l = l + fmt.Sprintf("{%s}", lt)
		}
		tn := strings.TrimSuffix(m.Name, "_sum")
		tn = strings.TrimSuffix(tn, "_count")
		tn = strings.Replace(tn, ".", "_", -1)
		if _, ok := info[tn]; !ok {
			info[tn] = mt
		}
		prom[tn] = prom[tn] + fmt.Sprintf("%s %v\n", l, m.Value)
	}
	m := ""
	for n, t := range info {
		m = fmt.Sprintf("# TYPE %s %s\n", n, t) + prom[n] + m
	}
	log.Println(m)
	return job, inst, m
}

func postMetrics(j string, i string, m string) error {
	url := fmt.Sprintf("%s/job/%s/instance/%s", plugin.URL, j, i)
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
	j, i, m := transformMetrics(event)
	err := postMetrics(j, i, m)
	return err
}
