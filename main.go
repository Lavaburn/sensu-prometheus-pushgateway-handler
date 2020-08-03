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
	URL             string
	DefaultJob      string
	DefaultInstance string
	DefaultType     string
	Job             string
	Instance        string
	Debug           bool
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
			Path:      "default-job",
			Env:       "DEFAULT_PROM_JOB",
			Argument:  "default-job",
			Shorthand: "j",
			Default:   "",
			Usage:     "The Prometheus job name to use when metrics do not have a prom_job tag.",
			Value:     &plugin.DefaultJob,
		},
		&sensu.PluginConfigOption{
			Path:      "default-instance",
			Env:       "DEFAULT_PROM_INSTANCE",
			Argument:  "default-instance",
			Shorthand: "i",
			Default:   "",
			Usage:     "The Prometheus instance name to use when metrics do not have a prom_instance tag.",
			Value:     &plugin.DefaultInstance,
		},
		&sensu.PluginConfigOption{
			Path:      "default-type",
			Env:       "DEFAULT_PROM_TYPE",
			Argument:  "default-type",
			Shorthand: "t",
			Default:   "untyped",
			Usage:     "The Prometheus metric type to use when metrics do not have a prom_type tag.",
			Value:     &plugin.DefaultType,
		},
		&sensu.PluginConfigOption{
			Path:      "job",
			Env:       "PROM_JOB",
			Argument:  "job",
			Shorthand: "J",
			Default:   "",
			Usage:     "The Prometheus job name to use, ignoring metric prom_job tags.",
			Value:     &plugin.Job,
		},
		&sensu.PluginConfigOption{
			Path:      "instance",
			Env:       "PROM_INSTANCE",
			Argument:  "instance",
			Shorthand: "I",
			Default:   "",
			Usage:     "The Prometheus instance name to use, ignoring metric prom_instance tags.",
			Value:     &plugin.Instance,
		},
		&sensu.PluginConfigOption{
			Path:      "debug",
			Env:       "DEBUG",
			Argument:  "debug",
			Shorthand: "d",
			Default:   false,
			Usage:     "Turn on debug mode (i.e. print the post body metrics).",
			Value:     &plugin.Debug,
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
	// The default job name will be overriden by a metric
	// "prom_job" tag value.
	job := plugin.DefaultJob
	// A configured job name takes precedence over metric
	// "prom_job" tag values.
	if plugin.Job != "" {
		job = plugin.Job
	}
	// The default instance name will be overriden by a metric
	// "prom_instance" tag value.
	inst := plugin.DefaultInstance
	// A configured instance name takes precedence over metric
	// "prom_instance" tag values.
	if plugin.Instance != "" {
		inst = plugin.Instance
	}
	// All lines for a given metric must be provided as one single
	// group, with the TYPE token line first. A map is used to
	// create groups from unordered Sensu Go metric points.
	prom := map[string]string{}
	for _, m := range event.Metrics.Points {
		mt := plugin.DefaultType
		lt := ""
		for _, t := range m.Tags {
			// Sensu Go can collect metrics from various
			// sources (i.e. Nagios PerfData). Prometheus
			// requires a job and instance name and a
			// metric type. Metrics can specify these with
			// the "prom_job", "prom_instance", and
			// "prom_type" metric tags.
			switch t.Name {
			case "prom_job":
				if job == plugin.DefaultJob {
					job = t.Value
				}
			case "prom_instance":
				if inst == plugin.DefaultInstance {
					inst = t.Value
				}
			case "prom_type":
				mt = t.Value
			default:
				if lt != "" {
					lt = lt + ","
				}
				// Regular Prometheus label key/value pair.
				lt = lt + fmt.Sprintf("%s=\"%s\"", t.Name, t.Value)
			}
		}
		// Prometheus histograms and summaries use special
		// metric name suffixes, they need to be stripped
		// before the lines can be grouped.
		tn := strings.TrimSuffix(m.Name, "_sum")
		tn = strings.TrimSuffix(tn, "_count")
		tn = strings.TrimSuffix(tn, "_bucket")
		tn = strings.Replace(tn, ".", "_", -1)
		// A metric grouping must begin with a TYPE token line.
		if _, ok := prom[tn]; !ok {
			prom[tn] = fmt.Sprintf("# TYPE %s %s\n", tn, mt)
		}
		// Prometheus does not support dot notation pathed
		// metric names. This is not applied to the suffix
		// trimmed name, this is to support metrics from
		// sources that are not aware of Prometheus' histogram
		// and summary metric naming scheme (i.e. statsd,
		// `foo.count`).
		l := strings.Replace(m.Name, ".", "_", -1)
		if lt != "" {
			l = l + fmt.Sprintf("{%s}", lt)
		}
		prom[tn] = prom[tn] + fmt.Sprintf("%s %v\n", l, m.Value)
	}
	// Put it all together.
	m := ""
	for _, v := range prom {
		m = m + v
	}
	if plugin.Debug {
		log.Println(m)
	}
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
	if plugin.Debug {
		log.Println("executing handler with --url", plugin.URL)
		log.Println("executing handler with --default-job", plugin.DefaultJob)
		log.Println("executing handler with --default-instance", plugin.DefaultInstance)
		log.Println("executing handler with --default-type", plugin.DefaultType)
		log.Println("executing handler with --job", plugin.Job)
		log.Println("executing handler with --instance", plugin.Instance)
		log.Println("executing handler with --debug", plugin.Debug)
	}
	j, i, m := transformMetrics(event)
	err := postMetrics(j, i, m)
	return err
}
