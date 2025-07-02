package collectors

import (
	"log/slog"
	"transport-nsw-exporter/pkg/collectors/carpark"

	"github.com/prometheus/client_golang/prometheus"
)

func RegisterCollectors(logger *slog.Logger) {
	carParkCollector := carpark.NewCarkParkCollector(logger)
	prometheus.MustRegister(carParkCollector)
}
