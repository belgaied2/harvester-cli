package cmd

import (
	"testing"

	"github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	memoryLimit := "3G"
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

func TestMergeOptionsInUserData(t *testing.T) {
	userData := `ssh_authorized_keys:
  - ssh-rsa AAAAB3NzaC1yc2EAAA ... custom@foo
packages:
  - docker
runcmd:
  - docker run -d --restart=unless-stopped -p 80:80 rancher/hello-world
`

	sshKey := &v1beta1.KeyPair{
		ObjectMeta: v1.ObjectMeta{
			Name: "test-ssh-key",
		},
		Spec: v1beta1.KeyPairSpec{
			PublicKey: "ssh-rsa AAAAB4MabD2zd3FBBB ... predef@bar",
		},
	}

	result, err := MergeOptionsInUserData(userData, defaultCloudInitUserData, sshKey)
	if err != nil {
		t.Errorf("Error merging options in user data: %v", err)
	}

	var resultMap map[string]interface{}
	err = yaml.Unmarshal([]byte(result), &resultMap)
	if err != nil {
		t.Errorf("Error unmarshalling result: %v", err)
	}

	sshAuthorizedKeys := resultMap["ssh_authorized_keys"].([]interface{})
	if len(sshAuthorizedKeys) != 2 {
		t.Errorf("Expected 2 ssh keys, got %d", len(sshAuthorizedKeys))
	}

	packages := resultMap["packages"].([]interface{})
	if len(packages) != 2 {
		t.Errorf("Expected 2 packages, got %d", len(packages))
	}

	runcmds := resultMap["runcmd"].([]interface{})
	if len(runcmds) != 4 {
		t.Errorf("Expected 4 runcmds, got %d", len(runcmds))
	}
}
