package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var Requests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_total_requests",
		Help: "Total Requests",
	},
	[]string{"path"},
)

func init() {
	prometheus.MustRegister(Requests)
}
