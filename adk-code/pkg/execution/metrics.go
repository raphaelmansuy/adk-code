package execution

import (
	"fmt"
	"sync"
	"time"
)

// MetricType represents the type of metric
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeTimer     MetricType = "timer"
)

// Metric represents a single metric data point
type Metric struct {
	// Name is the metric name
	Name string

	// Type is the metric type
	Type MetricType

	// Value is the metric value
	Value float64

	// Timestamp is when the metric was recorded
	Timestamp time.Time

	// Tags are additional metadata
	Tags map[string]string
}

// MetricsCollector collects execution metrics
type MetricsCollector struct {
	// mu protects the metrics map
	mu sync.RWMutex

	// metrics stores collected metrics
	metrics []*Metric

	// counters tracks counter values
	counters map[string]float64

	// gauges tracks gauge values
	gauges map[string]float64

	// startTime is when the collector was started
	startTime time.Time
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics:   []*Metric{},
		counters:  make(map[string]float64),
		gauges:    make(map[string]float64),
		startTime: time.Now(),
	}
}

// RecordMetric records a metric
func (mc *MetricsCollector) RecordMetric(metric *Metric) error {
	if metric == nil {
		return fmt.Errorf("metric is nil")
	}

	if metric.Name == "" {
		return fmt.Errorf("metric name is required")
	}

	metric.Timestamp = time.Now()

	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.metrics = append(mc.metrics, metric)
	return nil
}

// IncrementCounter increments a counter metric
func (mc *MetricsCollector) IncrementCounter(name string, value float64, tags map[string]string) error {
	metric := &Metric{
		Name:  name,
		Type:  MetricTypeCounter,
		Value: value,
		Tags:  tags,
	}

	mc.mu.Lock()
	mc.counters[name] += value
	mc.mu.Unlock()

	return mc.RecordMetric(metric)
}

// SetGauge sets a gauge metric
func (mc *MetricsCollector) SetGauge(name string, value float64, tags map[string]string) error {
	metric := &Metric{
		Name:  name,
		Type:  MetricTypeGauge,
		Value: value,
		Tags:  tags,
	}

	mc.mu.Lock()
	mc.gauges[name] = value
	mc.mu.Unlock()

	return mc.RecordMetric(metric)
}

// GetCounterValue gets the current value of a counter
func (mc *MetricsCollector) GetCounterValue(name string) float64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return mc.counters[name]
}

// GetGaugeValue gets the current value of a gauge
func (mc *MetricsCollector) GetGaugeValue(name string) float64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return mc.gauges[name]
}

// GetMetrics returns all collected metrics
func (mc *MetricsCollector) GetMetrics() []*Metric {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	// Return a copy
	result := make([]*Metric, len(mc.metrics))
	copy(result, mc.metrics)
	return result
}

// GetMetricsByType returns metrics of a specific type
func (mc *MetricsCollector) GetMetricsByType(metricType MetricType) []*Metric {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	var result []*Metric
	for _, m := range mc.metrics {
		if m.Type == metricType {
			result = append(result, m)
		}
	}
	return result
}

// Clear clears all metrics
func (mc *MetricsCollector) Clear() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.metrics = []*Metric{}
	mc.counters = make(map[string]float64)
	mc.gauges = make(map[string]float64)
}

// Summary returns a summary of collected metrics
func (mc *MetricsCollector) Summary() map[string]interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return map[string]interface{}{
		"total_metrics": len(mc.metrics),
		"counters":      len(mc.counters),
		"gauges":        len(mc.gauges),
		"uptime":        time.Since(mc.startTime).Seconds(),
	}
}

// ExecutionMetrics tracks metrics for a single execution
type ExecutionMetrics struct {
	// ExecutionID is the execution identifier
	ExecutionID string

	// StartTime is when execution started
	StartTime time.Time

	// EndTime is when execution ended
	EndTime time.Time

	// Duration is the total execution duration
	Duration time.Duration

	// CPUUsage is CPU usage percentage (0-100)
	CPUUsage float64

	// MemoryUsage is memory usage in MB
	MemoryUsage float64

	// CommandsExecuted is the number of commands executed
	CommandsExecuted int

	// SuccessfulCommands is the number of successful commands
	SuccessfulCommands int

	// FailedCommands is the number of failed commands
	FailedCommands int

	// BytesTransferred is the number of bytes transferred
	BytesTransferred int64

	// ErrorCount is the total number of errors
	ErrorCount int
}

// SuccessRate returns the success rate as a percentage
func (em *ExecutionMetrics) SuccessRate() float64 {
	if em.CommandsExecuted == 0 {
		return 0
	}
	return float64(em.SuccessfulCommands) / float64(em.CommandsExecuted) * 100
}

// AverageCommandTime returns the average time per command
func (em *ExecutionMetrics) AverageCommandTime() time.Duration {
	if em.CommandsExecuted == 0 {
		return 0
	}
	return em.Duration / time.Duration(em.CommandsExecuted)
}

// ExecutionMetricsTracker tracks metrics across multiple executions
type ExecutionMetricsTracker struct {
	// mu protects the metrics map
	mu sync.RWMutex

	// executions maps execution ID to metrics
	executions map[string]*ExecutionMetrics

	// aggregated stores aggregated metrics
	aggregated map[string]float64
}

// NewExecutionMetricsTracker creates a new execution metrics tracker
func NewExecutionMetricsTracker() *ExecutionMetricsTracker {
	return &ExecutionMetricsTracker{
		executions: make(map[string]*ExecutionMetrics),
		aggregated: make(map[string]float64),
	}
}

// RecordExecution records metrics for an execution
func (emt *ExecutionMetricsTracker) RecordExecution(metrics *ExecutionMetrics) error {
	if metrics == nil {
		return fmt.Errorf("metrics is nil")
	}

	if metrics.ExecutionID == "" {
		return fmt.Errorf("execution ID is required")
	}

	emt.mu.Lock()
	defer emt.mu.Unlock()

	emt.executions[metrics.ExecutionID] = metrics

	// Update aggregated metrics
	emt.aggregated["total_executions"] += 1
	emt.aggregated["total_duration"] += metrics.Duration.Seconds()
	emt.aggregated["total_commands"] += float64(metrics.CommandsExecuted)
	emt.aggregated["total_bytes"] += float64(metrics.BytesTransferred)
	emt.aggregated["total_errors"] += float64(metrics.ErrorCount)

	return nil
}

// GetExecution gets metrics for a specific execution
func (emt *ExecutionMetricsTracker) GetExecution(id string) (*ExecutionMetrics, error) {
	emt.mu.RLock()
	defer emt.mu.RUnlock()

	metrics, exists := emt.executions[id]
	if !exists {
		return nil, fmt.Errorf("execution %q not found", id)
	}

	return metrics, nil
}

// GetExecutions returns all recorded execution metrics
func (emt *ExecutionMetricsTracker) GetExecutions() []*ExecutionMetrics {
	emt.mu.RLock()
	defer emt.mu.RUnlock()

	var result []*ExecutionMetrics
	for _, m := range emt.executions {
		result = append(result, m)
	}
	return result
}

// GetAggregatedMetrics returns aggregated metrics across all executions
func (emt *ExecutionMetricsTracker) GetAggregatedMetrics() map[string]float64 {
	emt.mu.RLock()
	defer emt.mu.RUnlock()

	// Return a copy
	result := make(map[string]float64)
	for k, v := range emt.aggregated {
		result[k] = v
	}
	return result
}

// GetAverageMetrics returns average metrics across all executions
func (emt *ExecutionMetricsTracker) GetAverageMetrics() map[string]float64 {
	emt.mu.RLock()
	defer emt.mu.RUnlock()

	count := float64(len(emt.executions))
	if count == 0 {
		return make(map[string]float64)
	}

	return map[string]float64{
		"avg_duration":        emt.aggregated["total_duration"] / count,
		"avg_commands":        emt.aggregated["total_commands"] / count,
		"avg_bytes":           emt.aggregated["total_bytes"] / count,
		"avg_errors":          emt.aggregated["total_errors"] / count,
		"avg_memory_usage":    emt.aggregated["total_memory"] / count,
		"avg_cpu_usage":       emt.aggregated["total_cpu"] / count,
	}
}

// TraceEntry represents a single trace entry
type TraceEntry struct {
	// Timestamp is when the trace was recorded
	Timestamp time.Time

	// Operation is the operation being traced
	Operation string

	// Duration is how long the operation took
	Duration time.Duration

	// Status is the operation status (success, error, etc.)
	Status string

	// Details contains operation-specific details
	Details map[string]interface{}
}

// Tracer records execution traces
type Tracer struct {
	// mu protects the traces
	mu sync.RWMutex

	// traces stores trace entries
	traces []*TraceEntry

	// enabled indicates if tracing is enabled
	enabled bool
}

// NewTracer creates a new tracer
func NewTracer(enabled bool) *Tracer {
	return &Tracer{
		traces:  []*TraceEntry{},
		enabled: enabled,
	}
}

// Record records a trace entry
func (t *Tracer) Record(operation string, duration time.Duration, status string, details map[string]interface{}) {
	if !t.enabled {
		return
	}

	entry := &TraceEntry{
		Timestamp: time.Now(),
		Operation: operation,
		Duration:  duration,
		Status:    status,
		Details:   details,
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.traces = append(t.traces, entry)
}

// GetTraces returns all recorded traces
func (t *Tracer) GetTraces() []*TraceEntry {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Return a copy
	result := make([]*TraceEntry, len(t.traces))
	copy(result, t.traces)
	return result
}

// GetTracesByOperation returns traces for a specific operation
func (t *Tracer) GetTracesByOperation(operation string) []*TraceEntry {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var result []*TraceEntry
	for _, entry := range t.traces {
		if entry.Operation == operation {
			result = append(result, entry)
		}
	}
	return result
}

// Clear clears all traces
func (t *Tracer) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.traces = []*TraceEntry{}
}

// IsEnabled returns whether tracing is enabled
func (t *Tracer) IsEnabled() bool {
	return t.enabled
}
