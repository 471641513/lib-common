package metrics

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

/**
version: 1.0.2
*/
const ERR_SUCC = "succ"
const ERR_ERR = "err"

func CalTimecost(startTime time.Time) float64 {
	return float64(time.Now().Sub(startTime).Seconds() * 1000)
}

type MetricsBase struct {
	timecostMetricVec        *prometheus.HistogramVec
	timecostMetricSummeryVec *prometheus.SummaryVec
	MetricsCountVec          *prometheus.CounterVec
	MetricsGaugeVec          *prometheus.GaugeVec
}

func (metrics *MetricsBase) CreateMetrics(
	prefix string,
	buckets []float64,
	labels []string) *prometheus.HistogramVec {
	if nil == buckets {
		buckets = []float64{5, 10, 20, 60, 100, 200, 500}
	}

	metrics.timecostMetricVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    prefix,
			Help:    prefix + ":" + "time cost",
			Buckets: buckets,
		},
		labels)
	return metrics.timecostMetricVec
}

func (metrics *MetricsBase) CreateMetricsSummeryVec(
	prefix string,
	labels []string) *prometheus.SummaryVec {

	metrics.timecostMetricSummeryVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  prefix,
			Subsystem:  prefix,
			Name:       prefix + "time_cost",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001, 1.0: 0.0001},
		},
		labels)
	return metrics.timecostMetricSummeryVec
}

func (metrics *MetricsBase) CreateMetricsCountVec(
	prefix,
	subsystem,
	name string,
	labels []string) *prometheus.CounterVec {

	metrics.MetricsCountVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: prefix,
			Subsystem: subsystem,
			Name:      name,
		},
		labels)
	return metrics.MetricsCountVec
}

func (metrics *MetricsBase) CreateMetricsGaugeVec(
	prefix,
	subsystem,
	name string,
	labels []string) *prometheus.GaugeVec {

	metrics.MetricsGaugeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: prefix,
			Subsystem: subsystem,
			Name:      name,
		},
		labels)
	return metrics.MetricsGaugeVec
}

func (metrics *MetricsBase) GetTimeoutMetrics() *prometheus.HistogramVec {
	return metrics.timecostMetricVec
}

func (metrics *MetricsBase) Observe(timeCost float64, labels ...interface{}) {

	if metrics.timecostMetricVec == nil {
		return
	}
	labelStr := make([]string, len(labels))
	for idx, lable := range labels {
		labelStr[idx] = fmt.Sprintf("%v", lable)
	}
	metrics.timecostMetricVec.WithLabelValues(labelStr...).
		Observe(timeCost)
}

func (metrics *MetricsBase) GetTimeoutMetricsSummery() *prometheus.SummaryVec {
	return metrics.timecostMetricSummeryVec
}

func (metrics *MetricsBase) ObserveSummery(timeCost float64, labels ...string) {

	if metrics.timecostMetricSummeryVec == nil {
		return
	}
	metrics.timecostMetricSummeryVec.WithLabelValues(labels...).
		Observe(timeCost)
}

func (metrics *MetricsBase) GetTimeoutMetricsCounter() *prometheus.CounterVec {
	return metrics.MetricsCountVec
}


func (metrics *MetricsBase) ObserveCounter(count float64, labels ...interface{}) {

	if metrics.MetricsCountVec == nil {
		return
	}
	labelStr := make([]string, len(labels))
	for idx, lable := range labels {
		labelStr[idx] = fmt.Sprintf("%v", lable)
	}
	metrics.MetricsCountVec.WithLabelValues(labelStr...).Add(count)
}

func (metrics *MetricsBase) GetMetricsGauge() *prometheus.GaugeVec {
	return metrics.MetricsGaugeVec
}

func (metrics *MetricsBase) ObserveGauge(count float64, labels ...interface{}) {

	if metrics.MetricsGaugeVec == nil {
		return
	}
	labelStr := make([]string, len(labels))
	for idx, lable := range labels {
		labelStr[idx] = fmt.Sprintf("%v", lable)
	}
	metrics.MetricsGaugeVec.WithLabelValues(labelStr...).Add(count)
}

func (metrics *MetricsBase) GetMetricsVectors() (collectors []prometheus.Collector) {
	collectors = []prometheus.Collector{}
	if metrics.timecostMetricSummeryVec != nil {
		collectors = append(collectors, metrics.timecostMetricSummeryVec)
	}
	if metrics.timecostMetricVec != nil {
		collectors = append(collectors, metrics.timecostMetricVec)
	}
	if metrics.MetricsCountVec != nil {
		collectors = append(collectors, metrics.MetricsCountVec)
	}
	if metrics.MetricsGaugeVec != nil{
		collectors = append(collectors, metrics.MetricsGaugeVec)
	}
	return
}
