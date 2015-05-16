package aver

type BackendType string

const (
	Graphite BackendType = "graphite"
	MySQL                = "mysql"
)

type Options struct {
	// base url of backend
	Host string

	// prefix of metric names
	Prefix string

	// type of backend
	MetricBackend BackendType

	// categories
	CategoryMapping map[string]map[float64]string
}

func NewOptions() (o Options) {
	o.MetricBackend = Graphite
	o.Prefix = ""
	o.Host = "localhost"
	o.CategoryMapping = nil
	return
}
