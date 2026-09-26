package observability

// nopMeter 空实现：observability 未启用（或单测中）时零开销
type nopMeter struct{}

type nopCounter struct{}
type nopGauge struct{}
type nopHistogram struct{}

func (nopCounter) Inc(...string)                 {}
func (nopCounter) Add(float64, ...string)        {}
func (nopGauge) Set(float64, ...string)          {}
func (nopGauge) Inc(...string)                   {}
func (nopGauge) Dec(...string)                   {}
func (nopHistogram) Observe(float64, ...string)  {}

func (nopMeter) NewCounter(string, string, []string) Counter {
	return nopCounter{}
}

func (nopMeter) NewGauge(string, string, []string) Gauge {
	return nopGauge{}
}

func (nopMeter) NewHistogram(string, string, []float64, []string) Histogram {
	return nopHistogram{}
}
