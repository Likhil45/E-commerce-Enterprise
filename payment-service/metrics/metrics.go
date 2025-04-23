package metricsprom

import "github.com/prometheus/client_golang/prometheus"

var PaymentsProcessed = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "payments_processed_total",
		Help: "Total number of payments processed",
	},
	[]string{"status"},
)
var (
	PaymentSuccessTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "payment_success_total",
			Help: "Total number of successful payments",
		},
	)
	PaymentFailureTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "payment_failure_total",
			Help: "Total number of failed payments",
		},
	)
)

func Init() {
	prometheus.MustRegister(PaymentSuccessTotal)
	prometheus.MustRegister(PaymentFailureTotal)
	prometheus.MustRegister(PaymentsProcessed)

}
