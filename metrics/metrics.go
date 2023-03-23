package metrics

import (
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"maunium.net/go/mautrix/id"
)

var (
	knownSources = map[string]bool{
		"discord":    true,
		"facebook":   true,
		"googlechat": true,
		"hangouts":   true,
		"instagram":  true,
		"linkedin":   true,
		"signal":     true,
		"skype":      true,
		"slack":      true,
		"telegram":   true,
		"twitter":    true,
		"whatsapp":   true,
	}
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
	messagesCustomer = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "honoroit_messages_customer",
			Help: "The total number of messages, sent by customers",
		},
		[]string{"source", "sender", "domain"})
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

// MessagesCustomer increments count of messages from customers
func MessagesCustomer(sender id.UserID) {
	messagesTotal.Inc()

	messagesCustomer.With(prometheus.Labels{
		"source": getSource(sender),
		"sender": sender.String(),
		"domain": sender.Homeserver(),
	}).Inc()
}

// MessagesOperator increments count of messages from operator
func MessagesOperator() {
	messagesTotal.Inc()
	messagesOperator.Inc()
}

func getSource(sender id.UserID) string {
	parts := strings.Split(sender.Localpart(), "_")
	if len(parts) < 2 {
		return "matrix"
	}
	if knownSources[parts[0]] {
		return parts[0]
	}

	return "matrix"
}
