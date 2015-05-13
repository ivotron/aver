import (
    "strings"
    "encoding/json"
    "net/http"
    "github.com/jmoiron/jsonq"
)

type DataPoint []float64

type Metric struct {
    Target string      `json:"target"`
    Points []DataPoint `json:"datapoints"`
}

type Metrics struct {
    Names []string `json:`
}

func perror(err error) {
    if err != nil {
        panic(err)
    }
}

url := "http://" + graphiteHost + "/metrics/index.json"

res, err := http.Get(url)
perror(err)
defer res.Body.Close()
body, err := ioutil.ReadAll(res.Body)
perror(err)

dec := json.NewDecoder(strings.NewReader(jsonstring))
dec.Decode(&data)
jq := jsonq.NewQuery(data)

var results []Metric
err := json.Unmarshal([]byte(data), &results)

url = "http://" + graphiteHost + "/render/?target=" + exp + "*&format=json"

