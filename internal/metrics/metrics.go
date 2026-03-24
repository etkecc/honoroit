package metrics

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/VictoriaMetrics/metrics"
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
	requestsNew      = metrics.NewCounter("honoroit_request_new")
	requestsDone     = metrics.NewCounter("honoroit_request_done")
	messagesTotal    = metrics.NewCounter("honoroit_messages_total")
	messagesOperator = metrics.NewCounter("honoroit_messages_operator")
)

// Handler for metrics
type Handler struct{}

func (h *Handler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	metrics.WritePrometheus(w, false)
}

// InitMetrics registers /metrics endpoint within default http serve mux
func InitMetrics() {
	http.Handle("/metrics", &Handler{})
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

	metrics.GetOrCreateCounter(
		fmt.Sprintf(
			"honoroit_messages_customer{source=%q,sender=%q,domain=%q}",
			getSource(sender), sender.String(), sender.Homeserver(),
		),
	).Inc()
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
