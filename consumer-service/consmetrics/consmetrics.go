package consmetrics

import "github.com/prometheus/client_golang/prometheus"

var GrpcCallsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "grpc_calls_total",
		Help: "Total number of gRPC calls made by the consumer-service",
	},
	[]string{"service", "method", "status"},
)
var (
	KafkaMessagesProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_messages_processed_total",
			Help: "Total number of Kafka messages processed",
		},
		[]string{"topic", "status"},
	)
)
var (
	ConsumerRestartsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "kafka_consumer_restarts_total",
			Help: "Total number of Kafka consumer restarts",
		},
	)
)

func Init() {
	prometheus.MustRegister(GrpcCallsTotal)
	prometheus.MustRegister(ConsumerRestartsTotal)
	prometheus.MustRegister(KafkaMessagesProcessed)

}
