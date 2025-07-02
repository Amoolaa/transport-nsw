package carpark

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"transport-nsw-exporter/pkg/client"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	namespace        = "transport"
	carParkSubsystem = "car_park"
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
	carParkInfo         *prometheus.Desc
	carParkCurrentSpots *prometheus.Desc
	carParkTotalSpots   *prometheus.Desc
	carParkErrors       *prometheus.Desc
	carParkLastUpdated  *prometheus.Desc
	client              *client.Client[Facility]
	logger              *slog.Logger
}

func NewCarkParkCollector(logger *slog.Logger) *CarParkCollector {
	return &CarParkCollector{
		logger: logger,
		client: client.New[Facility](
			http.Client{
				Timeout: time.Second * 10,
			},
		),
		carParkErrors: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, carParkSubsystem, "errors"),
			"Car park scraper errors",
			nil,
			nil,
		),
		carParkInfo: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, carParkSubsystem, "info"),
			"Car park information",
			[]string{"facility_id"},
			nil,
		),
		carParkCurrentSpots: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, carParkSubsystem, "current_spots"),
			"Car park current spots",
			[]string{"facility_id"},
			nil,
		),
		carParkTotalSpots: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, carParkSubsystem, "total_spots"),
			"Car park total spots",
			[]string{"facility_id"},
			nil,
		),
		carParkLastUpdated: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, carParkSubsystem, "last_updated_ts"),
			"Car park last updated timestamp",
			[]string{"facility_id"},
			nil,
		),
	}
}

func (c *CarParkCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.carParkInfo
}

func (c *CarParkCollector) Collect(ch chan<- prometheus.Metric) {
	params := url.Values{}
	params.Add("facility", "25")
	errorCount := 0
	defer func() {
		ch <- prometheus.MustNewConstMetric(c.carParkErrors, prometheus.GaugeValue, float64(errorCount))
	}()

	resp, err := c.client.Do(context.Background(), http.MethodGet, "/carpark", params, nil)
	if err != nil {
		c.logger.Error("unable to fetch carpark data", "error", err)
		errorCount++
		return
	}
	ch <- prometheus.MustNewConstMetric(c.carParkInfo, prometheus.GaugeValue, 1, resp.FacilityID)

	currentSpots, err := strconv.Atoi(resp.Occupancy.Total)
	if err != nil {
		errorCount += 1
	}

	ch <- prometheus.MustNewConstMetric(c.carParkCurrentSpots, prometheus.GaugeValue, float64(currentSpots), resp.FacilityID)

	totalSpots, err := strconv.Atoi(resp.Spots)
	if err != nil {
		errorCount += 1
	}

	ch <- prometheus.MustNewConstMetric(c.carParkTotalSpots, prometheus.GaugeValue, float64(totalSpots), resp.FacilityID)

	t, err := time.Parse("2006-01-02T15:04:05", resp.MessageDate)
	if err != nil {
		c.logger.Error("unable to parse time", "error", err)
		errorCount += 1
	} else {
		ch <- prometheus.MustNewConstMetric(c.carParkLastUpdated, prometheus.GaugeValue, float64(t.Unix()), resp.FacilityID)
	}
}
