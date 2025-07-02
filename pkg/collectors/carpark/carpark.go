package carpark

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"transport-nsw-exporter/internal/config"
	"transport-nsw-exporter/pkg/client"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	namespace        = "transport"
	carParkSubsystem = "car_park"
	commonLabels     = []string{"facility_id", "facility_name", "suburb", "address"}
)

type Facility struct {
	TSN             string    `json:"tsn"`
	Time            string    `json:"time"`
	Spots           string    `json:"spots"`
	Zones           []Zone    `json:"zones"`
	ParkID          string    `json:"ParkID"`
	Location        Location  `json:"location"`
	Occupancy       Occupancy `json:"occupancy"`
	MessageDate     string    `json:"MessageDate"`
	FacilityID      string    `json:"facility_id"`
	FacilityName    string    `json:"facility_name"`
	TFNSWFacilityID string    `json:"tfnsw_facility_id"`
}

type Zone struct {
	Spots        string    `json:"spots"`
	ZoneID       string    `json:"zone_id"`
	Occupancy    Occupancy `json:"occupancy"`
	ZoneName     string    `json:"zone_name"`
	ParentZoneID string    `json:"parent_zone_id"`
}

type Location struct {
	Suburb    string `json:"suburb"`
	Address   string `json:"address"`
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}

type Occupancy struct {
	Loop       *string `json:"loop"`
	Total      string  `json:"total"`
	Monthlies  *string `json:"monthlies"`
	OpenGate   *string `json:"open_gate"`
	Transients *string `json:"transients"`
}

type CarParkCollector struct {
	carParkInfo                *prometheus.Desc
	carParkCurrentVehicleCount *prometheus.Desc
	carParkTotalSpots          *prometheus.Desc
	carParkErrors              *prometheus.Desc
	carParkLastUpdated         *prometheus.Desc
	client                     *client.Client[Facility]
	logger                     *slog.Logger
	cfg                        config.Carpark
}

func NewCarkParkCollector(logger *slog.Logger, carparkCfg config.Carpark) *CarParkCollector {
	return &CarParkCollector{
		cfg:    carparkCfg,
		logger: logger,
		client: client.New[Facility](
			http.Client{
				Timeout: time.Second * 10,
			},
		),
		carParkErrors: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, carParkSubsystem, "errors"),
			"Number of collector errors",
			[]string{"facility_id"},
			nil,
		),
		carParkInfo: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, carParkSubsystem, "info"),
			"Car park information",
			commonLabels,
			nil,
		),
		carParkCurrentVehicleCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, carParkSubsystem, "vehicle_count"),
			"Car park current vehicle count",
			commonLabels,
			nil,
		),
		carParkTotalSpots: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, carParkSubsystem, "spots_total"),
			"Car park total spots",
			commonLabels,
			nil,
		),
		carParkLastUpdated: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, carParkSubsystem, "last_updated_seconds"),
			"Car park last updated timestamp",
			commonLabels,
			nil,
		),
	}
}

func (c *CarParkCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.carParkInfo
}

func (c *CarParkCollector) CollectFromID(ch chan<- prometheus.Metric, id string) {
	params := url.Values{}
	params.Add("facility", id)
	errorCount := 0
	defer func() {
		ch <- prometheus.MustNewConstMetric(c.carParkErrors, prometheus.GaugeValue, float64(errorCount), id)
	}()

	resp, err := c.client.Do(context.Background(), http.MethodGet, "/carpark", params, nil)
	if err != nil {
		c.logger.Error("unable to fetch carpark data", "error", err)
		errorCount++
		return
	}

	commonLabelValues := []string{id, resp.FacilityName, resp.Location.Suburb, resp.Location.Address}

	ch <- prometheus.MustNewConstMetric(c.carParkInfo, prometheus.GaugeValue, 1, commonLabelValues...)

	currentSpots, err := strconv.Atoi(resp.Occupancy.Total)
	if err != nil {
		errorCount += 1
	}

	ch <- prometheus.MustNewConstMetric(c.carParkCurrentVehicleCount, prometheus.GaugeValue, float64(currentSpots), commonLabelValues...)

	totalSpots, err := strconv.Atoi(resp.Spots)
	if err != nil {
		errorCount += 1
	}

	ch <- prometheus.MustNewConstMetric(c.carParkTotalSpots, prometheus.GaugeValue, float64(totalSpots), commonLabelValues...)

	t, err := time.Parse("2006-01-02T15:04:05", resp.MessageDate)
	if err != nil {
		c.logger.Error("unable to parse time", "error", err)
		errorCount += 1
	} else {
		ch <- prometheus.MustNewConstMetric(c.carParkLastUpdated, prometheus.GaugeValue, float64(t.Unix()), commonLabelValues...)
	}
}

func (c *CarParkCollector) Collect(ch chan<- prometheus.Metric) {
	for _, id := range c.cfg.FacilityIDs {
		c.CollectFromID(ch, id)
	}
}
