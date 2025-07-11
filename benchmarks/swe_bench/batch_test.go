package swe_bench

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultBatchConfig(t *testing.T) {
	config := DefaultBatchConfig()
	
	if config == nil {
		t.Fatal("DefaultBatchConfig returned nil")
	}
	
	// Test default values
	if config.Agent.Model.Name == "" {
		t.Error("Default model name should not be empty")
	}
	
	if config.Agent.Model.Temperature < 0 || config.Agent.Model.Temperature > 2 {
		t.Error("Default temperature should be between 0 and 2")
	}
	
	if config.NumWorkers <= 0 {
		t.Error("Default workers should be positive")
	}
	
	if config.Instances.Type == "" {
		t.Error("Default dataset type should not be empty")
	}
	
	if config.OutputPath == "" {
		t.Error("Default output path should not be empty")
	}
}

func TestConfigManager(t *testing.T) {
	cm := NewConfigManager()
	
	// Test validation with default config
	config := DefaultBatchConfig()
	err := cm.ValidateConfig(config)
	if err != nil {
		t.Errorf("Default config should be valid: %v", err)
	}
	
	// Test validation with invalid config
	invalidConfig := &BatchConfig{}
	err = cm.ValidateConfig(invalidConfig)
	if err == nil {
		t.Error("Empty config should be invalid")
	}
	
	// Test config save and load
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")
	
	err = cm.SaveConfig(config, configPath)
	if err != nil {
		t.Errorf("Failed to save config: %v", err)
	}
	
	loadedConfig, err := cm.LoadConfig(configPath)
	if err != nil {
		t.Errorf("Failed to load config: %v", err)
	}
	
	if loadedConfig.Agent.Model.Name != config.Agent.Model.Name {
		t.Error("Loaded config does not match saved config")
	}
}

func TestDatasetLoader(t *testing.T) {
	loader := NewDatasetLoader()
	
	// Test config validation
	validConfig := DatasetConfig{
		Type:   "swe_bench",
		Subset: "lite",
		Split:  "dev",
	}
	
	err := loader.ValidateConfig(validConfig)
	if err != nil {
		t.Errorf("Valid config should pass validation: %v", err)
	}
	
	// Test invalid config
	invalidConfig := DatasetConfig{
		Type:   "swe_bench",
		Subset: "invalid",
		Split:  "dev",
	}
	
	err = loader.ValidateConfig(invalidConfig)
	if err == nil {
		t.Error("Invalid config should fail validation")
	}
	
	// Test file config validation
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.json")
	os.WriteFile(testFile, []byte("[]"), 0644)
	
	fileConfig := DatasetConfig{
		Type:     "file",
		FilePath: testFile,
	}
	
	err = loader.ValidateConfig(fileConfig)
	if err != nil {
		t.Errorf("File config should be valid: %v", err)
	}
}

func TestResultWriter(t *testing.T) {
	writer := NewResultWriter()
	tempDir := t.TempDir()
	
	// Create test results
	results := []WorkerResult{
		{
			InstanceID: "test_1",
			Status:     StatusCompleted,
			Solution:   "test solution",
			Duration:   time.Second,
			TokensUsed: 100,
			Cost:       0.01,
		},
		{
			InstanceID: "test_2",
			Status:     StatusFailed,
			Error:      "test error",
			Duration:   time.Second * 2,
			TokensUsed: 50,
			Cost:       0.005,
		},
	}
	
	// Test writing partial results
	err := writer.WritePartialResults(context.Background(), results, tempDir)
	if err != nil {
		t.Errorf("Failed to write partial results: %v", err)
	}
	
	// Check that files were created
	predsFile := filepath.Join(tempDir, "preds_partial.json")
	if _, err := os.Stat(predsFile); os.IsNotExist(err) {
		t.Error("Partial predictions file was not created")
	}
	
	// Test appending result
	err = writer.AppendResult(context.Background(), results[0], tempDir)
	if err != nil {
		t.Errorf("Failed to append result: %v", err)
	}
	
	streamFile := filepath.Join(tempDir, "streaming_results.jsonl")
	if _, err := os.Stat(streamFile); os.IsNotExist(err) {
		t.Error("Streaming results file was not created")
	}
	
	// Test validation
	err = writer.ValidateResults(results)
	if err != nil {
		t.Errorf("Valid results should pass validation: %v", err)
	}
	
	// Test invalid results
	invalidResults := []WorkerResult{
		{
			// Missing required fields
			Status: StatusCompleted,
		},
	}
	
	err = writer.ValidateResults(invalidResults)
	if err == nil {
		t.Error("Invalid results should fail validation")
	}
	
	// Test results stats
	stats := writer.GetResultsStats(results)
	if stats.Total != 2 {
		t.Errorf("Expected 2 total results, got %d", stats.Total)
	}
	if stats.Completed != 1 {
		t.Errorf("Expected 1 completed result, got %d", stats.Completed)
	}
	if stats.Failed != 1 {
		t.Errorf("Expected 1 failed result, got %d", stats.Failed)
	}
	if stats.SuccessRate != 50.0 {
		t.Errorf("Expected 50%% success rate, got %.1f%%", stats.SuccessRate)
	}
}

func TestWorkerPool(t *testing.T) {
	pool := NewWorkerPool(2)
	
	// Test initial state
	status := pool.GetStatus()
	if status.ActiveWorkers != 0 {
		t.Error("Initial active workers should be 0")
	}
	
	// Test worker count limits
	largPool := NewWorkerPool(100)
	if largPool.numWorkers > 20 {
		t.Error("Worker count should be limited to 20")
	}
	
	smallPool := NewWorkerPool(-1)
	if smallPool.numWorkers != 1 {
		t.Error("Worker count should default to 1 for invalid values")
	}
	
	// Test setting max workers
	err := pool.SetMaxWorkers(5)
	if err != nil {
		t.Errorf("Should be able to set max workers when pool is stopped: %v", err)
	}
	
	if pool.numWorkers != 5 {
		t.Errorf("Expected 5 workers, got %d", pool.numWorkers)
	}
}

func TestProgressReporter(t *testing.T) {
	reporter := NewProgressReporter()
	
	// Test start and stop
	ctx := context.Background()
	err := reporter.Start(ctx)
	if err != nil {
		t.Errorf("Failed to start progress reporter: %v", err)
	}
	
	// Test update
	update := ProgressUpdate{
		Timestamp: time.Now(),
		Total:     10,
		Completed: 5,
		Failed:    1,
		Running:   2,
		Remaining: 2,
	}
	
	err = reporter.Update(update)
	if err != nil {
		t.Errorf("Failed to update progress: %v", err)
	}
	
	err = reporter.Stop()
	if err != nil {
		t.Errorf("Failed to stop progress reporter: %v", err)
	}
	
	// Test double start
	err = reporter.Start(ctx)
	if err != nil {
		t.Errorf("Failed to restart progress reporter: %v", err)
	}
	
	err = reporter.Start(ctx)
	if err == nil {
		t.Error("Starting already running reporter should fail")
	}
	
	reporter.Stop()
}

func TestMonitor(t *testing.T) {
	monitor := NewMonitor()
	
	// Test start and stop
	ctx := context.Background()
	err := monitor.StartMonitoring(ctx)
	if err != nil {
		t.Errorf("Failed to start monitor: %v", err)
	}
	
	// Test recording metrics
	tags := map[string]string{"test": "value"}
	err = monitor.RecordMetric("test_metric", 42.0, tags)
	if err != nil {
		t.Errorf("Failed to record metric: %v", err)
	}
	
	// Test recording events
	eventData := map[string]interface{}{"status": "ok"}
	err = monitor.RecordEvent("test_event", eventData)
	if err != nil {
		t.Errorf("Failed to record event: %v", err)
	}
	
	// Test getting metrics
	metrics := monitor.GetMetrics()
	if len(metrics) == 0 {
		t.Error("Should have recorded metrics")
	}
	
	// Test getting events
	events := monitor.GetEvents()
	if len(events) == 0 {
		t.Error("Should have recorded events")
	}
	
	err = monitor.StopMonitoring()
	if err != nil {
		t.Errorf("Failed to stop monitor: %v", err)
	}
}

func TestAgentFactory(t *testing.T) {
	factory := NewAgentFactory()
	
	// Test config validation
	config := DefaultBatchConfig()
	err := factory.ValidateConfig(config)
	if err != nil {
		t.Errorf("Default config should be valid: %v", err)
	}
	
	// Test invalid config
	invalidConfig := &BatchConfig{}
	err = factory.ValidateConfig(invalidConfig)
	if err == nil {
		t.Error("Empty config should be invalid")
	}
}

func BenchmarkConfigValidation(b *testing.B) {
	cm := NewConfigManager()
	config := DefaultBatchConfig()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.ValidateConfig(config)
	}
}

func BenchmarkResultWriting(b *testing.B) {
	writer := NewResultWriter()
	tempDir := b.TempDir()
	
	result := WorkerResult{
		InstanceID: "bench_test",
		Status:     StatusCompleted,
		Solution:   "benchmark solution",
		Duration:   time.Millisecond,
		TokensUsed: 100,
		Cost:       0.01,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writer.AppendResult(context.Background(), result, tempDir)
	}
}

// Helper function to create test instances
func createTestInstances(count int) []Instance {
	instances := make([]Instance, count)
	for i := 0; i < count; i++ {
		instances[i] = Instance{
			ID:               fmt.Sprintf("test_%d", i),
			RepoURL:          "https://github.com/test/repo",
			BaseCommit:       "abc123",
			ProblemStatement: fmt.Sprintf("Test problem %d", i),
		}
	}
	return instances
}

func TestCreateTestInstances(t *testing.T) {
	instances := createTestInstances(3)
	
	if len(instances) != 3 {
		t.Errorf("Expected 3 instances, got %d", len(instances))
	}
	
	for i, instance := range instances {
		if instance.ID != fmt.Sprintf("test_%d", i) {
			t.Errorf("Instance %d has wrong ID: %s", i, instance.ID)
		}
		
		if instance.RepoURL == "" {
			t.Errorf("Instance %d missing repo URL", i)
		}
		
		if instance.ProblemStatement == "" {
			t.Errorf("Instance %d missing problem statement", i)
		}
	}
}