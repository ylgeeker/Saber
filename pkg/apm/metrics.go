/**
 * Copyright 2025 Saber authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
**/

// Metrics types and constructors. Define metrics with NewCounter, NewGauge,
// NewHistogram, or NewSummary; update them via WithLabelValues(...).Inc(),
// .Set(), .Observe(), etc. No need to register with HTTP—use apm.MetricsRoute()
// with sbnet.Server to expose /metrics.
package apm

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Counter is a cumulative metric. Create with NewCounter, then use
// WithLabelValues(labels...).Inc() or .Add(delta) to update.
type Counter struct {
	vec *prometheus.CounterVec
}

// CounterLabelled is a Counter with fixed label values. Use Inc or Add to update.
type CounterLabelled struct {
	c prometheus.Counter
}

// WithLabelValues returns a CounterLabelled for the given label values. Call
// Inc() or Add(delta) on it to update the metric.
func (c *Counter) WithLabelValues(lvs ...string) *CounterLabelled {
	return &CounterLabelled{c: c.vec.WithLabelValues(lvs...)}
}

// Inc increments the counter by 1.
func (c *CounterLabelled) Inc() { c.c.Inc() }

// Add adds the given value to the counter.
func (c *CounterLabelled) Add(delta float64) { c.c.Add(delta) }

// NewCounter creates and registers a Counter with the default registry.
// subsystem and name form the metric name; help is the description.
// labelNames are the label keys; use WithLabelValues to set values when updating.
func NewCounter(subsystem, name, help string, labelNames []string) *Counter {
	opts := prometheus.CounterOpts{Subsystem: subsystem, Name: name, Help: help}
	vec := prometheus.NewCounterVec(opts, labelNames)
	defaultRegistry.MustRegister(vec)
	return &Counter{vec: vec}
}

// Gauge is a value that can go up or down. Create with NewGauge, then use
// WithLabelValues(labels...).Set(v), .Inc(), .Dec(), or .Add(delta).
type Gauge struct {
	vec *prometheus.GaugeVec
}

// GaugeLabelled is a Gauge with fixed label values.
type GaugeLabelled struct {
	g prometheus.Gauge
}

// WithLabelValues returns a GaugeLabelled for the given label values.
func (g *Gauge) WithLabelValues(lvs ...string) *GaugeLabelled {
	return &GaugeLabelled{g: g.vec.WithLabelValues(lvs...)}
}

// Set sets the gauge to the given value.
func (g *GaugeLabelled) Set(v float64) { g.g.Set(v) }

// Inc increments the gauge by 1.
func (g *GaugeLabelled) Inc() { g.g.Inc() }

// Dec decrements the gauge by 1.
func (g *GaugeLabelled) Dec() { g.g.Dec() }

// Add adds the given value to the gauge.
func (g *GaugeLabelled) Add(delta float64) { g.g.Add(delta) }

// NewGauge creates and registers a Gauge with the default registry.
func NewGauge(subsystem, name, help string, labelNames []string) *Gauge {
	opts := prometheus.GaugeOpts{Subsystem: subsystem, Name: name, Help: help}
	vec := prometheus.NewGaugeVec(opts, labelNames)
	defaultRegistry.MustRegister(vec)
	return &Gauge{vec: vec}
}

// Histogram observes values and counts them in configurable buckets. Create with
// NewHistogram, then use WithLabelValues(labels...).Observe(v).
type Histogram struct {
	vec *prometheus.HistogramVec
}

// HistogramLabelled is a Histogram with fixed label values.
type HistogramLabelled struct {
	h prometheus.Observer
}

// WithLabelValues returns a HistogramLabelled for the given label values.
func (h *Histogram) WithLabelValues(lvs ...string) *HistogramLabelled {
	return &HistogramLabelled{h: h.vec.WithLabelValues(lvs...)}
}

// Observe records a value in the histogram.
func (h *HistogramLabelled) Observe(v float64) { h.h.Observe(v) }

// NewHistogram creates and registers a Histogram with the default registry.
// buckets define the upper bounds of buckets (e.g. prometheus.DefBuckets).
func NewHistogram(subsystem, name, help string, buckets []float64, labelNames []string) *Histogram {
	opts := prometheus.HistogramOpts{
		Subsystem: subsystem,
		Name:      name,
		Help:      help,
		Buckets:   buckets,
	}
	vec := prometheus.NewHistogramVec(opts, labelNames)
	defaultRegistry.MustRegister(vec)
	return &Histogram{vec: vec}
}

// Summary observes values and reports quantiles. Create with NewSummary, then
// use WithLabelValues(labels...).Observe(v).
type Summary struct {
	vec *prometheus.SummaryVec
}

// SummaryLabelled is a Summary with fixed label values.
type SummaryLabelled struct {
	s prometheus.Observer
}

// WithLabelValues returns a SummaryLabelled for the given label values.
func (s *Summary) WithLabelValues(lvs ...string) *SummaryLabelled {
	return &SummaryLabelled{s: s.vec.WithLabelValues(lvs...)}
}

// Observe records a value in the summary.
func (s *SummaryLabelled) Observe(v float64) { s.s.Observe(v) }

// NewSummary creates and registers a Summary with the default registry.
// objectives is the map of quantile -> tolerance (e.g. 0.5 -> 0.05 for median).
func NewSummary(subsystem, name, help string, objectives map[float64]float64, labelNames []string) *Summary {
	opts := prometheus.SummaryOpts{
		Subsystem:  subsystem,
		Name:      name,
		Help:      help,
		Objectives: objectives,
	}
	vec := prometheus.NewSummaryVec(opts, labelNames)
	defaultRegistry.MustRegister(vec)
	return &Summary{vec: vec}
}
