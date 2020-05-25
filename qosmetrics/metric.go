package qosmetrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type IGRPCMetric interface {
	IncConsumed(methodName string)
	IncErrored(methodName string)
	ObserveLatency(methodName string, start time.Time)
}

type grpcMetrics struct {
	errorCount    *prometheus.CounterVec
	consumedCount *prometheus.CounterVec
	latencyHisto  *prometheus.HistogramVec
}

func New(nameSpace string) (grpcMetrics, error) {

	m := grpcMetrics{}
	m.errorCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: nameSpace,
		Name:      "grpc_method_errored_count",
		Help:      "gRPC method errored count",
	}, []string{"method"})

	m.consumedCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: nameSpace,
		Name:      "grpc_method_consumed_count",
		Help:      "gRPC method consumed count",
	}, []string{"method"})

	m.latencyHisto = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: nameSpace,
			Name:      "grpc_method_latency_seconds",
			Help:      "gRPC method invokation duration",
		},
		[]string{"method"})

	return m, m.registerMetrics()
}

func (m grpcMetrics) IncConsumed(methodName string) {
	m.consumedCount.WithLabelValues(methodName).Inc()
}

func (m grpcMetrics) IncErrored(methodName string) {
	m.errorCount.WithLabelValues(methodName).Inc()
}

func (m grpcMetrics) ObserveLatency(methodName string, start time.Time) {
	m.latencyHisto.WithLabelValues(methodName).Observe(time.Since(start).Seconds())
}

func (m grpcMetrics) registerMetrics() error {

	if err := ProcessPrometheusError(prometheus.Register(m.consumedCount)); err != nil {
		return err
	}
	if err := ProcessPrometheusError(prometheus.Register(m.errorCount)); err != nil {
		return err
	}
	if err := ProcessPrometheusError(prometheus.Register(m.latencyHisto)); err != nil {
		return err
	}

	return nil
}

func ProcessPrometheusError(err error) error {
	if err == nil {
		return nil
	}

	switch err.(type) {
	case prometheus.AlreadyRegisteredError:
		return nil
	default:
		return err
	}
}
