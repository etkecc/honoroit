package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	requestsNew = promauto.NewCounter(prometheus.CounterOpts{
		Name: "honoroit_request_new",
		Help: "The total number new/incoming requests",
	})
	requestsDone = promauto.NewCounter(prometheus.CounterOpts{
		Name: "honoroit_request_done",
		Help: "The total number of closed/done requests",
	})
	messagesTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "honoroit_messages_total",
		Help: "The total number of messages",
	})
	messagesCustomer = promauto.NewCounter(prometheus.CounterOpts{
		Name: "honoroit_messages_customer",
		Help: "The total number of messages, sent by customers",
	})
	messagesOperator = promauto.NewCounter(prometheus.CounterOpts{
		Name: "honoroit_messages_operator",
		Help: "The total number of messages, sent by operators",
	})
)

// InitMetrics registers /metrics endpoint within default http serve mux
func InitMetrics() {
	http.Handle("/metrics", promhttp.Handler())
}

// RequestNew increments count of new requests
func RequestNew() {
	requestsNew.Inc()
}

// RequestDone increments count of closed requests
func RequestDone() {
	requestsDone.Inc()
}

// Messages increments count of messages
func Messages(customer bool) {
	messagesTotal.Inc()
	if customer {
		messagesCustomer.Inc()
	} else {
		messagesOperator.Inc()
	}
}
