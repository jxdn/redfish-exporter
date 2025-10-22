package exporter

import (
	"context"
	"log"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type RedfishCollector struct {
	cfg    RedfishCollectorConfig
	client *RedfishClient

	powerConsumed *prometheus.Desc
	powerAvg      *prometheus.Desc
	powerMin      *prometheus.Desc
	powerMax      *prometheus.Desc
}

func NewRedfishCollector(cfg RedfishCollectorConfig) (*RedfishCollector, error) {
	cli, err := NewRedfishClient(cfg.BaseURL, cfg.Username, cfg.Password, cfg.InsecureTLS, cfg.ScrapeTimeout)
	if err != nil {
		return nil, err
	}
	labels := []string{"chassis", "control_name"}
	return &RedfishCollector{
		cfg:    cfg,
		client: cli,
		powerConsumed: prometheus.NewDesc(
			"redfish_power_consumed_watts",
			"Instantaneous power consumption in watts",
			labels, nil,
		),
		powerAvg: prometheus.NewDesc(
			"redfish_power_average_watts",
			"Average power consumption over the controller-provided interval in watts",
			labels, nil,
		),
		powerMin: prometheus.NewDesc(
			"redfish_power_min_watts",
			"Minimum power consumption over the controller-provided interval in watts",
			labels, nil,
		),
		powerMax: prometheus.NewDesc(
			"redfish_power_max_watts",
			"Maximum power consumption over the controller-provided interval in watts",
			labels, nil,
		),
	}, nil
}

func (c *RedfishCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.powerConsumed
	ch <- c.powerAvg
	ch <- c.powerMin
	ch <- c.powerMax
}

func (c *RedfishCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), c.cfg.ScrapeTimeout)
	defer cancel()

	chassisIDs := []string{}
	if c.cfg.ChassisID != "" {
		// If ChassisID is a full odata id, use as-is; else construct
		if strings.HasPrefix(c.cfg.ChassisID, "/redfish/") {
			chassisIDs = []string{c.cfg.ChassisID}
		} else {
			chassisIDs = []string{"/redfish/v1/Chassis/" + c.cfg.ChassisID}
		}
	} else {
		ids, err := c.client.ListChassis()
		if err != nil {
			log.Printf("error listing chassis: %v", err)
			return
		}
		chassisIDs = ids
	}

	_ = ctx // currently unused, but kept for future request scoping

	for _, oid := range chassisIDs {
		power, err := c.client.GetChassisPower(oid)
		if err != nil {
			log.Printf("error getting power for %s: %v", oid, err)
			continue
		}
		chassisLabel := lastPathComponent(oid)
		for _, pc := range power.PowerControl {
			controlName := pc.Name
			labels := []string{chassisLabel, controlName}
			if pc.PowerConsumedWatts != nil {
				ch <- prometheus.MustNewConstMetric(c.powerConsumed, prometheus.GaugeValue, *pc.PowerConsumedWatts, labels...)
			}
			if pc.PowerMetrics.AverageConsumedWatts != nil {
				ch <- prometheus.MustNewConstMetric(c.powerAvg, prometheus.GaugeValue, *pc.PowerMetrics.AverageConsumedWatts, labels...)
			}
			if pc.PowerMetrics.MinConsumedWatts != nil {
				ch <- prometheus.MustNewConstMetric(c.powerMin, prometheus.GaugeValue, *pc.PowerMetrics.MinConsumedWatts, labels...)
			}
			if pc.PowerMetrics.MaxConsumedWatts != nil {
				ch <- prometheus.MustNewConstMetric(c.powerMax, prometheus.GaugeValue, *pc.PowerMetrics.MaxConsumedWatts, labels...)
			}
		}
	}
}

func lastPathComponent(odataID string) string {
	if odataID == "" {
		return odataID
	}
	s := strings.TrimSuffix(odataID, "/")
	idx := strings.LastIndex(s, "/")
	if idx == -1 || idx+1 >= len(s) {
		return s
	}
	return s[idx+1:]
}
