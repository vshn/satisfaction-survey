package metrics

import "github.com/prometheus/client_golang/prometheus"
import "github.com/prometheus/client_golang/prometheus/promauto"

type PromCounter struct {
	positive prometheus.Counter
	negative prometheus.Counter
	total    prometheus.Counter
}

func NewCounter(reg prometheus.Registerer) *PromCounter {
	c := &PromCounter{
		positive: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "satisfaction_survey_positive_total",
			Help: "Total number of positive responses",
		}),
		negative: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "satisfaction_survey_negative_total",
			Help: "Total number of negative responses",
		}),
		total: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "satisfaction_survey_total",
			Help: "Total number of responses",
		}),
	}
	return c
}

func (c PromCounter) IncPositive() {
	c.positive.Inc()
	c.total.Inc()
}

func (c PromCounter) IncNegative() {
	c.negative.Inc()
	c.total.Inc()
}
