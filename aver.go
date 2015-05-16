package aver

type AverError struct {
	Msg string
}

func (e AverError) Error() string {
	return "aver: " + e.Msg
}

type DataPoint []float64

type Metric struct {
	Target     string      `json:"target"`
	Points     []DataPoint `json:"datapoints"`
	Categories map[float64]string
}

type Backend interface {
	// obtains metrics
	GetMetrics() (map[string]Metric, error)
}

// checks metrics against a validation string
func Aver(validation string, metrics map[string]Metric) (bool, error) {
	return true, nil
}
