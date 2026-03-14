package http

import (
	"net/http"

	"office/internal/query"
	"office/internal/service"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewMetricsHandler returns a Prometheus metrics endpoint handler.
func NewMetricsHandler(sessionSvc *service.SessionService, environmentSvc *service.EnvironmentService, espHealthSvc *service.ESPHealthService) http.Handler {
	registry := prometheus.NewRegistry()
	registry.MustRegister(&officeCollector{
		sessionSvc:     sessionSvc,
		environmentSvc: environmentSvc,
		espHealthSvc:   espHealthSvc,
		activeUsers: prometheus.NewDesc(
			"office_presence_active_users",
			"Number of users currently checked in",
			nil,
			nil,
		),
		temperatureC: prometheus.NewDesc(
			"office_environment_temperature_celsius",
			"Latest fresh office temperature in Celsius",
			nil,
			nil,
		),
		espDeviceUp: prometheus.NewDesc(
			"office_esp_health_up",
			"ESP device heartbeat freshness status (1=fresh, 0=stale)",
			[]string{"device_id"},
			nil,
		),
		espUptimeSeconds: prometheus.NewDesc(
			"office_esp_health_uptime_seconds",
			"ESP device uptime in seconds (fresh heartbeats only)",
			[]string{"device_id"},
			nil,
		),
		espFreeHeapBytes: prometheus.NewDesc(
			"office_esp_health_free_heap_bytes",
			"ESP device free heap in bytes (fresh heartbeats only)",
			[]string{"device_id"},
			nil,
		),
		espRSSIDbm: prometheus.NewDesc(
			"office_esp_health_rssi_dbm",
			"ESP device Wi-Fi RSSI in dBm (fresh heartbeats only)",
			[]string{"device_id"},
			nil,
		),
	})

	return promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
}

type officeCollector struct {
	sessionSvc     *service.SessionService
	environmentSvc *service.EnvironmentService
	espHealthSvc   *service.ESPHealthService

	activeUsers      *prometheus.Desc
	temperatureC     *prometheus.Desc
	espDeviceUp      *prometheus.Desc
	espUptimeSeconds *prometheus.Desc
	espFreeHeapBytes *prometheus.Desc
	espRSSIDbm       *prometheus.Desc
}

func (c *officeCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.activeUsers
	ch <- c.temperatureC
	ch <- c.espDeviceUp
	ch <- c.espUptimeSeconds
	ch <- c.espFreeHeapBytes
	ch <- c.espRSSIDbm
}

func (c *officeCollector) Collect(ch chan<- prometheus.Metric) {
	if c.sessionSvc != nil {
		count, err := c.sessionSvc.CountSessions(query.SessionFilter{ActiveOnly: true})
		if err != nil {
			log.Warn("failed to collect presence metric", "error", err)
		} else {
			ch <- prometheus.MustNewConstMetric(c.activeUsers, prometheus.GaugeValue, float64(count))
		}
	}

	if c.environmentSvc != nil {
		if reading, ok := c.environmentSvc.GetFresh(); ok {
			ch <- prometheus.MustNewConstMetric(c.temperatureC, prometheus.GaugeValue, reading.TemperatureC)
		}
	}

	if c.espHealthSvc != nil {
		for _, status := range c.espHealthSvc.GetAll() {
			isFresh := c.espHealthSvc.IsFresh(status)
			upValue := 0.0
			if isFresh {
				upValue = 1.0
			}

			ch <- prometheus.MustNewConstMetric(c.espDeviceUp, prometheus.GaugeValue, upValue, status.DeviceID)

			if !isFresh {
				continue
			}

			ch <- prometheus.MustNewConstMetric(c.espUptimeSeconds, prometheus.GaugeValue, float64(status.UptimeSeconds), status.DeviceID)
			ch <- prometheus.MustNewConstMetric(c.espFreeHeapBytes, prometheus.GaugeValue, float64(status.FreeHeapBytes), status.DeviceID)
			ch <- prometheus.MustNewConstMetric(c.espRSSIDbm, prometheus.GaugeValue, float64(status.RSSI), status.DeviceID)
		}
	}
}
