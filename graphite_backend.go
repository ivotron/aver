package aver

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

type GraphiteBackend struct {
	prefix     string
	host       string
	categories map[string]map[float64]string
}

func NewGraphiteBackend(o Options) (b *GraphiteBackend, err error) {
	return &GraphiteBackend{prefix: o.Prefix, host: o.Host, categories: o.CategoryMapping}, nil
}

func getData(url string) (data []byte, err error) {
	res, err := http.Get(url)

	if err != nil {
		return
	}

	defer res.Body.Close()

	data, err = ioutil.ReadAll(res.Body)

	return
}

func (b *GraphiteBackend) GetMetrics() (metrics map[string]Metric, err error) {
	data, err := getData("http://" + b.host + "/metrics/index.json")

	if err != nil {
		return
	}

	var names []string

	if err = json.Unmarshal([]byte(data), &names); err != nil {
		return
	}

	if len(names) == 0 {
		return nil, AverError{"no metrics in graphite index"}
	}

	metrics = make(map[string]Metric)

	for _, name := range names {
		if strings.HasPrefix(name, b.prefix) {
			data, err = getData("http://" + b.host + "/render?target=" + name + "&format=json")

			if err != nil {
				return
			}

			var metric []Metric

			if err = json.Unmarshal([]byte(data), &metric); err != nil {
				return
			}

			if len(metric) != 1 {
				return nil, AverError{"expecting 1 metric only"}
			}

			metric[0].Categories = b.categories[name]
			metrics[name] = metric[0]
		}
	}

	return
}
