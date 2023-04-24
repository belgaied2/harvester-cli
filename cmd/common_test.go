package cmd

import (
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
)

func TestHandleCPUOverCommittment(t *testing.T) {
	cpuNumber := int64(4)
	overCommitSettingMap := map[string]int{
		"cpu":    1600,
		"memory": 150,
		"disk":   200,
	}

	result := HandleCPUOverCommittment(overCommitSettingMap, cpuNumber)
	if result.MilliValue() != 250 {
		t.Errorf("Expected 250m, got %dm", result.MilliValue())
	}
}

func TestHandleMemoryOverCommittment(t *testing.T) {
	memoryLimit := "4Gi"
	overCommitSettingMap := map[string]int{
		"cpu":    1600,
		"memory": 150,
		"disk":   200,
	}

	result := HandleMemoryOverCommittment(overCommitSettingMap, memoryLimit)
	if result.ScaledValue(resource.Giga) != 2 {
		t.Errorf("Expected 2G, got %dM", result.ScaledValue(resource.Mega))
	}
}
