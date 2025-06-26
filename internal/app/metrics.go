package app

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/mechta-market/e-product/internal/config"
	"github.com/mechta-market/e-product/internal/constant"
)

var (
	metricRequestCounter   *prometheus.CounterVec
	metricResponseDuration *prometheus.HistogramVec
)

func init() {
	if !config.Conf.WithMetrics {
		return
	}

	slog.Info("metrics enabled")

	metricRequestCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: config.Conf.Namespace,
		Name:      constant.ServiceName + "_request_count",
	}, []string{
		"protocol",
		"method",
		"status",
	})

	metricResponseDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: config.Conf.Namespace,
		Name:      constant.ServiceName + "_response_duration_seconds",
	}, []string{
		"protocol",
		"method",
		"status",
	})
}
