package aver

func NewBackendWithDefaultOptions() (backend Backend, err error) {
	opts := NewOptions()
	return NewBackend(opts)
}

func NewBackend(opts Options) (backend Backend, err error) {
	switch opts.MetricBackend {
	case Graphite:
		if backend, err = NewGraphiteBackend(opts); err != nil {
			return
		}
	default:
		return nil, AverError{"unknown backend " + string(opts.MetricBackend)}
	}

	return
}
