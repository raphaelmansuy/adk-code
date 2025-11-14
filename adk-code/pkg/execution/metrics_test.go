package execution

import (
	"testing"
	"time"
)

// TestMetric tests basic metric creation
func TestMetric(t *testing.T) {
	metric := &Metric{
		Name:      "test_metric",
		Type:      MetricTypeCounter,
		Value:     42.0,
		Timestamp: time.Now(),
		Tags: map[string]string{
			"service": "execution",
		},
	}

	if metric.Name != "test_metric" {
		t.Fatalf("Expected name 'test_metric', got %q", metric.Name)
	}

	if metric.Type != MetricTypeCounter {
		t.Fatalf("Expected type MetricTypeCounter, got %v", metric.Type)
	}

	if metric.Value != 42.0 {
		t.Fatalf("Expected value 42.0, got %f", metric.Value)
	}
}

// TestMetricsCollector tests collector creation
func TestMetricsCollector(t *testing.T) {
	collector := NewMetricsCollector()

	if collector == nil {
		t.Fatal("Failed to create collector")
	}

	metrics := collector.GetMetrics()
	if len(metrics) != 0 {
		t.Fatalf("Expected 0 metrics initially, got %d", len(metrics))
	}
}

// TestMetricsCollectorRecordMetric tests recording a metric
func TestMetricsCollectorRecordMetric(t *testing.T) {
	collector := NewMetricsCollector()

	metric := &Metric{
		Name:  "test_metric",
		Type:  MetricTypeGauge,
		Value: 100.0,
		Tags:  map[string]string{},
	}

	err := collector.RecordMetric(metric)
	if err != nil {
		t.Fatalf("Failed to record metric: %v", err)
	}

	metrics := collector.GetMetrics()
	if len(metrics) != 1 {
		t.Fatalf("Expected 1 metric, got %d", len(metrics))
	}
}

// TestMetricsCollectorIncrementCounter tests incrementing a counter
func TestMetricsCollectorIncrementCounter(t *testing.T) {
	collector := NewMetricsCollector()

	err := collector.IncrementCounter("requests", 1.0, map[string]string{})
	if err != nil {
		t.Fatalf("Failed to increment counter: %v", err)
	}

	value := collector.GetCounterValue("requests")
	if value != 1.0 {
		t.Fatalf("Expected counter value 1.0, got %f", value)
	}

	_ = collector.IncrementCounter("requests", 1.0, map[string]string{})

	value = collector.GetCounterValue("requests")
	if value != 2.0 {
		t.Fatalf("Expected counter value 2.0, got %f", value)
	}
}

// TestMetricsCollectorSetGauge tests setting a gauge
func TestMetricsCollectorSetGauge(t *testing.T) {
	collector := NewMetricsCollector()

	err := collector.SetGauge("memory_usage", 512.5, map[string]string{})
	if err != nil {
		t.Fatalf("Failed to set gauge: %v", err)
	}

	value := collector.GetGaugeValue("memory_usage")
	if value != 512.5 {
		t.Fatalf("Expected gauge value 512.5, got %f", value)
	}

	_ = collector.SetGauge("memory_usage", 768.0, map[string]string{})

	value = collector.GetGaugeValue("memory_usage")
	if value != 768.0 {
		t.Fatalf("Expected gauge value 768.0, got %f", value)
	}
}

// TestMetricsCollectorClear tests clearing metrics
func TestMetricsCollectorClear(t *testing.T) {
	collector := NewMetricsCollector()

	_ = collector.IncrementCounter("test", 1.0, map[string]string{})
	_ = collector.SetGauge("test", 100.0, map[string]string{})

	collector.Clear()

	metrics := collector.GetMetrics()
	if len(metrics) != 0 {
		t.Fatalf("Expected 0 metrics after clear, got %d", len(metrics))
	}
}

// TestMetricsCollectorSummary tests getting metrics summary
func TestMetricsCollectorSummary(t *testing.T) {
	collector := NewMetricsCollector()

	_ = collector.IncrementCounter("test", 1.0, map[string]string{})

	summary := collector.Summary()
	if summary == nil {
		t.Fatal("Expected non-nil summary")
	}
}

// TestExecutionMetrics tests execution metrics
func TestExecutionMetrics(t *testing.T) {
	metrics := &ExecutionMetrics{
		ExecutionID:        "exec-1",
		StartTime:          time.Now().Add(-10 * time.Second),
		EndTime:            time.Now(),
		CommandsExecuted:   10,
		SuccessfulCommands: 8,
		FailedCommands:     2,
		BytesTransferred:   1024,
		ErrorCount:         1,
	}

	metrics.Duration = metrics.EndTime.Sub(metrics.StartTime)

	if metrics.ExecutionID != "exec-1" {
		t.Fatalf("Expected ID 'exec-1', got %q", metrics.ExecutionID)
	}

	if metrics.Duration <= 0 {
		t.Fatal("Expected positive duration")
	}
}

// TestExecutionMetricsSuccessRate tests success rate calculation
func TestExecutionMetricsSuccessRate(t *testing.T) {
	metrics := &ExecutionMetrics{
		CommandsExecuted:   10,
		SuccessfulCommands: 8,
	}

	rate := metrics.SuccessRate()
	expected := 80.0
	if rate != expected {
		t.Fatalf("Expected success rate %.1f, got %.1f", expected, rate)
	}
}

// TestExecutionMetricsAverageTime tests average command time
func TestExecutionMetricsAverageTime(t *testing.T) {
	metrics := &ExecutionMetrics{
		Duration:         10 * time.Second,
		CommandsExecuted: 5,
	}

	avgTime := metrics.AverageCommandTime()
	expected := 2 * time.Second
	if avgTime != expected {
		t.Fatalf("Expected average time %v, got %v", expected, avgTime)
	}
}

// TestExecutionMetricsTracker tests tracker creation
func TestExecutionMetricsTracker(t *testing.T) {
	tracker := NewExecutionMetricsTracker()

	if tracker == nil {
		t.Fatal("Failed to create tracker")
	}

	executions := tracker.GetExecutions()
	if len(executions) != 0 {
		t.Fatalf("Expected 0 executions initially, got %d", len(executions))
	}
}

// TestTracer tests tracer creation
func TestTracer(t *testing.T) {
	tracer := NewTracer(true)

	if tracer == nil {
		t.Fatal("Failed to create tracer")
	}

	if !tracer.IsEnabled() {
		t.Fatal("Expected tracer to be enabled")
	}
}

// TestTracerDisabled tests disabled tracer
func TestTracerDisabled(t *testing.T) {
	tracer := NewTracer(false)

	if tracer.IsEnabled() {
		t.Fatal("Expected tracer to be disabled")
	}
}

// TestTracerRecord tests recording a trace
func TestTracerRecord(t *testing.T) {
	tracer := NewTracer(true)

	tracer.Record("command_execute", 100*time.Millisecond, "success", map[string]interface{}{
		"command": "echo test",
	})

	traces := tracer.GetTraces()
	if len(traces) != 1 {
		t.Fatalf("Expected 1 trace, got %d", len(traces))
	}

	if traces[0].Operation != "command_execute" {
		t.Fatalf("Expected operation 'command_execute', got %q", traces[0].Operation)
	}
}

// TestTracerDisabledRecord tests that disabled tracer doesn't record
func TestTracerDisabledRecord(t *testing.T) {
	tracer := NewTracer(false)

	tracer.Record("command_execute", 100*time.Millisecond, "success", map[string]interface{}{})

	traces := tracer.GetTraces()
	if len(traces) != 0 {
		t.Fatalf("Expected 0 traces when disabled, got %d", len(traces))
	}
}

// TestTracerClear tests clearing traces
func TestTracerClear(t *testing.T) {
	tracer := NewTracer(true)

	tracer.Record("test", 100*time.Millisecond, "success", map[string]interface{}{})
	tracer.Clear()

	traces := tracer.GetTraces()
	if len(traces) != 0 {
		t.Fatalf("Expected 0 traces after clear, got %d", len(traces))
	}
}
