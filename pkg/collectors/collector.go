package collectors

import (
	"log/slog"
	"transport-nsw-exporter/internal/config"
	"transport-nsw-exporter/pkg/collectors/carpark"

	"github.com/prometheus/client_golang/prometheus"
)

func RegisterCollectors(logger *slog.Logger, collectors config.Collectors) {
	carParkCollector := carpark.NewCarkParkCollector(logger, collectors.Carpark)
	prometheus.MustRegister(carParkCollector)
}
