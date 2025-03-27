package resource

import (
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/enrayga/omc-o2ims/internal/service/omc_rest"
	"github.com/stretchr/testify/assert"
)

func TestInitReconciliationInfo(t *testing.T) {

	t.Run("Test case: Check for InitReconciliationInfo when status is empty", func(t *testing.T) {

	})
}

func TestProvisioningRequestCreation(t *testing.T) {

	t.Run("Test case: Check for SetInitFields when status is not empty", func(t *testing.T) {
		// Arrange
		name := "test-provisioning-001"
		content := `
		{
			"apiVersion": "o2ims.provisioning.oran.org/v1alpha1",
			"kind": "ProvisioningRequest",
			"metadata": {
				"creationTimestamp": "2025-01-05T16:09:12Z",
				"generation": 1,
				"name": "test-provisioning-001",
				"resourceVersion": "365239",
				"uid": "3378091a-8260-4b84-9373-eb9b6831c3a3"
			},
			"spec": {
				"description": "Test provisioning request",
				"name": "Test Cluster Provisioning",
				"templateName": "cluster-template",
				"templateParameters": {
					"cluster_name": "test-cluster",
					"node_count": 3,
					"node_type": "worker"
				},
				"templateVersion": "v1.0"
			},
			"status": {
			}	
		}
		`
		var data map[string]interface{}
		_ = json.Unmarshal([]byte(content), &data)

		newResource := ProvisioningRequest{}
		err := newResource.SetInitFields(name, data)

		if err != nil {
			t.Fatal("Failed to set provisioning request fields:", err)
		}

		_, ok := newResource.Status["extensions"]
		assert.True(t, ok, "Status extensions not created")

		_, ok = newResource.Status["extensions"].(map[string]interface{})["reconciliationInfo"]
		assert.True(t, ok, "Status reconcilliationInfo not created")

		_, ok = newResource.Status["extensions"].(map[string]interface{})["reconciliationInfo"].(map[string]interface{})
		assert.True(t, ok, "Status reconcilliationInfo is not of map type")

	})
	t.Run("Test case: Check for SetInitFields when status is empty", func(t *testing.T) {
		var err error
		name := "test-provisioning-001"
		content := `
		{
			"apiVersion": "o2ims.provisioning.oran.org/v1alpha1",
			"kind": "ProvisioningRequest",
			"metadata": {
				"creationTimestamp": "2025-01-05T16:09:12Z",
				"generation": 1,
				"name": "test-provisioning-001",
				"resourceVersion": "365239",
				"uid": "3378091a-8260-4b84-9373-eb9b6831c3a3"
			},
			"spec": {
				"description": "Test provisioning request",
				"name": "Test Cluster Provisioning",
				"templateName": "cluster-template",
				"templateParameters": {
					"cluster_name": "test-cluster",
					"node_count": 3,
					"node_type": "worker"
				},
				"templateVersion": "v1.0"
			}
		}
		`
		var data map[string]interface{}
		_ = json.Unmarshal([]byte(content), &data)

		newResource := ProvisioningRequest{}
		err = newResource.SetInitFields(name, data)

		if err != nil {
			t.Fatal("Failed to set provisioning request fields:", err)
		}

		_, ok := newResource.Status["extensions"]
		assert.True(t, ok, "Status extensions not created")

		_, ok = newResource.Status["extensions"].(map[string]interface{})["reconciliationInfo"]
		assert.True(t, ok, "Status reconcilliationInfo not created")

		_, ok = newResource.Status["extensions"].(map[string]interface{})["reconciliationInfo"].(map[string]interface{})
		assert.True(t, ok, "Status reconcilliationInfo is not of map type")
	})

	t.Run("Test case: Check for SetInitFields when name is empty", func(t *testing.T) {
		var err error
		name := ""
		content := `
		{
			"apiVersion": "o2ims.provisioning.oran.org/v1alpha1",
			"kind": "ProvisioningRequest",
			"metadata": {
				"creationTimestamp": "2025-01-05T16:09:12Z",
				"generation": 1,
				"name": "test-provisioning-001",
				"resourceVersion": "365239",
				"uid": "3378091a-8260-4b84-9373-eb9b6831c3a3"
			},
			"spec": {
				"description": "Test provisioning request",
				"name": "Test Cluster Provisioning",
				"templateName": "cluster-template",
				"templateParameters": {
					"cluster_name": "test-cluster",
					"node_count": 3,
					"node_type": "worker"
				},
				"templateVersion": "v1.0"
			}
		}
		`
		var data map[string]interface{}
		_ = json.Unmarshal([]byte(content), &data)

		newResource := ProvisioningRequest{}
		err = newResource.SetInitFields(name, data)

		if err == nil {
			t.Error("Expected an error but got none")
		}
	})

	t.Run("Test case: Check for SetInitFields when Spec is empty or template parameters is empty", func(t *testing.T) {
		var err error
		name := "test-provisioning-001"
		content := `
		{
			"apiVersion": "o2ims.provisioning.oran.org/v1alpha1",
			"kind": "ProvisioningRequest",
			"metadata": {
				"creationTimestamp": "2025-01-05T16:09:12Z",
				"generation": 1,
				"name": "test-provisioning-001",
				"resourceVersion": "365239",
				"uid": "3378091a-8260-4b84-9373-eb9b6831c3a3"
			},
			"spec2": {
				"description": "Test provisioning request",
				"name": "Test Cluster Provisioning",
				"templateName": "cluster-template",
				"templateParameters": {
					"cluster_name": "test-cluster",
					"node_count": 3,
					"node_type": "worker"
				},
				"templateVersion": "v1.0"
			}
		}
		`
		var data map[string]interface{}
		_ = json.Unmarshal([]byte(content), &data)
		newResource := ProvisioningRequest{}
		err = newResource.SetInitFields(name, data)

		assert.Contains(t, err.Error(), "missing Spec in ProvisioningRequest")

		content = `
		{
			"apiVersion": "o2ims.provisioning.oran.org/v1alpha1",
			"kind": "ProvisioningRequest",
			"metadata": {
				"creationTimestamp": "2025-01-05T16:09:12Z",
				"generation": 1,
				"name": "test-provisioning-001",
				"resourceVersion": "365239",
				"uid": "3378091a-8260-4b84-9373-eb9b6831c3a3"
			},
			"spec": {
				"description": "Test provisioning request",
				"name": "Test Cluster Provisioning",
				"templateName2": "cluster-template",
				"templateParameters": {
					"cluster_name": "test-cluster",
					"node_count": 3,
					"node_type": "worker"
				},
				"templateVersion": "v1.0"
			}
		}
		`

		_ = json.Unmarshal([]byte(content), &data)
		newResource = ProvisioningRequest{}
		err = newResource.SetInitFields(name, data)
		assert.Contains(t, err.Error(), "missing templateName in spec")

		content = `
		{
			"apiVersion": "o2ims.provisioning.oran.org/v1alpha1",
			"kind": "ProvisioningRequest",
			"metadata": {
				"creationTimestamp": "2025-01-05T16:09:12Z",
				"generation": 1,
				"name": "test-provisioning-001",
				"resourceVersion": "365239",
				"uid": "3378091a-8260-4b84-9373-eb9b6831c3a3"
			},
			"spec": {
				"description": "Test provisioning request",
				"name": "Test Cluster Provisioning",
				"templateName": "cluster-template",
				"templateParameters": {
					"cluster_name": "test-cluster",
					"node_count": 3,
					"node_type": "worker"
				},
				"templateVersion1": "v1.0"
			}
		}
		`

		_ = json.Unmarshal([]byte(content), &data)
		newResource = ProvisioningRequest{}
		err = newResource.SetInitFields(name, data)
		assert.Contains(t, err.Error(), "missing templateVersion in spec")

		content = `
		{
			"apiVersion": "o2ims.provisioning.oran.org/v1alpha1",
			"kind": "ProvisioningRequest",
			"metadata": {
				"creationTimestamp": "2025-01-05T16:09:12Z",
				"generation": 1,
				"name": "test-provisioning-001",
				"resourceVersion": "365239",
				"uid": "3378091a-8260-4b84-9373-eb9b6831c3a3"
			},
			"spec": {
				"description": "Test provisioning request",
				"name": "Test Cluster Provisioning",
				"templateName": "cluster-template",
				"templateParameters1": {
					"cluster_name": "test-cluster",
					"node_count": 3,
					"node_type": "worker"
				},
				"templateVersion": "v1.0"
			}
		}
		`
		_ = json.Unmarshal([]byte(content), &data)
		newResource = ProvisioningRequest{}
		err = newResource.SetInitFields(name, data)
		assert.Contains(t, err.Error(), "missing templateParameters in spec")

	})
}

func TestProvisioningRequestCompare(t *testing.T) {

	t.Run("Test case: Compare when nothing is changed", func(t *testing.T) {
		var err error
		name := "test-provisioning-001"
		content := `
		{
			"apiVersion": "o2ims.provisioning.oran.org/v1alpha1",
			"kind": "ProvisioningRequest",
			"metadata": {
				"creationTimestamp": "2025-01-05T16:09:12Z",
				"generation": 1,
				"name": "test-provisioning-001",
				"resourceVersion": "365239",
				"uid": "3378091a-8260-4b84-9373-eb9b6831c3a3"
			},
			"spec": {
				"description": "Test provisioning request",
				"name": "Test Cluster Provisioning",
				"templateName": "cluster-template",
				"templateParameters": {
					"cluster_name": "test-cluster",
					"node_count": 3,
					"node_type": "worker"
				},
				"templateVersion": "v1.0"
			}
		}
		`
		var data map[string]interface{}
		_ = json.Unmarshal([]byte(content), &data)

		newResource := ProvisioningRequest{}
		err = newResource.SetInitFields(name, data)

		if err != nil {
			t.Fatal("Failed to set provisioning request fields:", err)
		}

		new_content := `
		{
			"apiVersion": "o2ims.provisioning.oran.org/v1alpha1",
			"kind": "ProvisioningRequest",
			"metadata": {
				"creationTimestamp": "2025-01-05T16:09:12Z",
				"generation": 1,
				"name": "test-provisioning-001",
				"resourceVersion": "365239",
				"uid": "3378091a-8260-4b84-9373-eb9b6831c3a3"
			},
			"spec": {
				"description": "Test provisioning request",
				"name": "Test Cluster Provisioning",
				"templateName": "cluster-template",
				"templateParameters": {
					"cluster_name": "test-cluster",
					"node_count": 3,
					"node_type": "worker"
				},
				"templateVersion": "v1.0"
			}
		}
		`
		var new_data map[string]interface{}
		_ = json.Unmarshal([]byte(new_content), &new_data)
		changed, _ := newResource.Compare(name, new_data, false)

		if changed {
			t.Error("Unexpected change detected. Expected no changes.")
		}
	})

	t.Run("Test case: Compare should detect template version change", func(t *testing.T) {
		name := "test-provisioning-001"
		content := `
		{
			"apiVersion": "o2ims.provisioning.oran.org/v1alpha1",
			"kind": "ProvisioningRequest",
			"metadata": {
				"creationTimestamp": "2025-01-05T16:09:12Z",
				"generation": 1,
				"name": "test-provisioning-001",
				"resourceVersion": "365239",
				"uid": "3378091a-8260-4b84-9373-eb9b6831c3a3"
			},
			"spec": {
				"description": "Test provisioning request",
				"name": "Test Cluster Provisioning",
				"templateName": "cluster-template",
				"templateParameters": {
					"cluster_name": "test-cluster",
					"node_count": 3,
					"node_type": "worker"
				},
				"templateVersion": "v1.0"
			}
		}
		`
		var data map[string]interface{}
		_ = json.Unmarshal([]byte(content), &data)

		newResource := ProvisioningRequest{}
		_ = newResource.SetInitFields(name, data)

		new_content := `
		{
			"apiVersion": "o2ims.provisioning.oran.org/v1alpha1",
			"kind": "ProvisioningRequest",
			"metadata": {
				"creationTimestamp": "2025-01-05T16:09:12Z",
				"generation": 1,
				"name": "test-provisioning-001",
				"resourceVersion": "365239",
				"uid": "3378091a-8260-4b84-9373-eb9b6831c3a3"
			},
			"spec": {
				"description": "Test provisioning request",
				"name": "Test Cluster Provisioning",
				"templateName": "cluster-template",
				"templateParameters": {
					"cluster_name": "test-cluster",
					"node_count": 3,
					"node_type": "worker"
				},
				"templateVersion": "v1.1"
			}
		}
		`
		var new_data map[string]interface{}
		_ = json.Unmarshal([]byte(new_content), &new_data)
		new_val := new_data["spec"].(map[string]interface{})["templateVersion"].(string)
		changed, _ := newResource.Compare(name, new_data, false)
		if !changed {
			t.Errorf("Expected Template Version change detected to be true, but got no change. Current version: %s, new version: %s", newResource.Spec["templateVersion"], new_val)
		}
	})
	t.Run("Test case: Compare should return true when templateParameters is/are changed", func(t *testing.T) {
		name := "test-provisioning-001"
		content := `
		{
			"apiVersion": "o2ims.provisioning.oran.org/v1alpha1",
			"kind": "ProvisioningRequest",
			"metadata": {
				"creationTimestamp": "2025-01-05T16:09:12Z",
				"generation": 1,
				"name": "test-provisioning-001",
				"resourceVersion": "365239",
				"uid": "3378091a-8260-4b84-9373-eb9b6831c3a3"
			},
			"spec": {
				"description": "Test provisioning request",
				"name": "Test Cluster Provisioning",
				"templateName": "cluster-template",
				"templateParameters": {
					"cluster_name": "test-cluster",
					"node_count": 3,
					"node_type": "worker"
				},
				"templateVersion": "v1.0"
			}
		}
		`
		var data map[string]interface{}
		_ = json.Unmarshal([]byte(content), &data)

		newResource := ProvisioningRequest{}
		_ = newResource.SetInitFields(name, data)

		new_content := `
		{
			"apiVersion": "o2ims.provisioning.oran.org/v1alpha1",
			"kind": "ProvisioningRequest",
			"metadata": {
				"creationTimestamp": "2025-01-05T16:09:12Z",
				"generation": 1,
				"name": "test-provisioning-001",
				"resourceVersion": "365239",
				"uid": "3378091a-8260-4b84-9373-eb9b6831c3a3"
			},
			"spec": {
				"description": "Test provisioning request",
				"name": "Test Cluster Provisioning",
				"templateName": "cluster-template",
				"templateParameters": {
					"cluster_name": "test-cluster",
					"node_count": 3,
					"node_type": "worker2"
				},
				"templateVersion": "v1.0"
			}
		}
		`
		var new_data map[string]interface{}
		_ = json.Unmarshal([]byte(new_content), &new_data)
		new_val := new_data["spec"].(map[string]interface{})["templateParameters"]

		changed, _ := newResource.Compare(name, new_data, false)
		if !changed {
			t.Errorf("Expected param 'node_type' to be changed from '%v' to '%v'", newResource.Spec["templateParameters"], new_val)
		}
	})
	t.Run("Check for Compare when template name is changed", func(t *testing.T) {
		name := "test-provisioning-001"
		content := `
		{
			"apiVersion": "o2ims.provisioning.oran.org/v1alpha1",
			"kind": "ProvisioningRequest",
			"metadata": {
				"creationTimestamp": "2025-01-05T16:09:12Z",
				"generation": 1,
				"name": "test-provisioning-001",
				"resourceVersion": "365239",
				"uid": "3378091a-8260-4b84-9373-eb9b6831c3a3"
			},
			"spec": {
				"description": "Test provisioning request",
				"name": "Test Cluster Provisioning",
				"templateName": "cluster-template",
				"templateParameters": {
					"cluster_name": "test-cluster",
					"node_count": 3,
					"node_type": "worker"
				},
				"templateVersion": "v1.0"
			}
		}
		`
		var data map[string]interface{}
		_ = json.Unmarshal([]byte(content), &data)
		old_val := data["spec"].(map[string]interface{})["templateName"]

		newResource := ProvisioningRequest{}
		_ = newResource.SetInitFields(name, data)

		new_content := `
		{
			"apiVersion": "o2ims.provisioning.oran.org/v1alpha1",
			"kind": "ProvisioningRequest",
			"metadata": {
				"creationTimestamp": "2025-01-05T16:09:12Z",
				"generation": 1,
				"name": "test-provisioning-001",
				"resourceVersion": "365239",
				"uid": "3378091a-8260-4b84-9373-eb9b6831c3a3"
			},
			"spec": {
				"description": "Test provisioning request",
				"name": "Test Cluster Provisioning",
				"templateName": "cluster-template2",
				"templateParameters": {
					"cluster_name": "test-cluster",
					"node_count": 3,
					"node_type": "worker"
				},
				"templateVersion": "v1.0"
			}
		}
		`
		var new_data map[string]interface{}
		_ = json.Unmarshal([]byte(new_content), &new_data)
		new_val := new_data["spec"].(map[string]interface{})["templateName"]
		changed, _ := newResource.Compare(name, new_data, false)
		if !changed {
			t.Errorf("Expected template name to be changed, but got no change.  old value: %v, new value %v", old_val, new_val)
		}
	})

	t.Run("Test case: Compare should detect template version change and apply change", func(t *testing.T) {

		name := "test-provisioning-001"

		content := `
		{
			"apiVersion": "o2ims.provisioning.oran.org/v1alpha1",
			"kind": "ProvisioningRequest",
			"metadata": {
				"creationTimestamp": "2025-01-05T16:09:12Z",
				"generation": 1,
				"name": "test-provisioning-001",
				"resourceVersion": "365239",
				"uid": "3378091a-8260-4b84-9373-eb9b6831c3a3"
			},
			"spec": {
				"description": "Test provisioning request",
				"name": "Test Cluster Provisioning",
				"templateName": "cluster-template",
				"templateParameters": {
					"cluster_name": "test-cluster",
					"node_count": 3,
					"node_type": "worker"
				},
				"templateVersion": "v1.0"
			}
		}
		`

		var data map[string]interface{}
		_ = json.Unmarshal([]byte(content), &data)
		old_val := data["spec"].(map[string]interface{})["templateVersion"]

		newResource := ProvisioningRequest{}

		_ = newResource.SetInitFields(name, data)

		new_content := `
		{
			"apiVersion": "o2ims.provisioning.oran.org/v1alpha1",
			"kind": "ProvisioningRequest",
			"metadata": {
				"creationTimestamp": "2025-01-05T16:09:12Z",
				"generation": 1,
				"name": "test-provisioning-001",
				"resourceVersion": "365239",
				"uid": "3378091a-8260-4b84-9373-eb9b6831c3a3"
			},
			"spec": {
				"description": "Test provisioning request",
				"name": "Test Cluster Provisioning",
				"templateName": "cluster-template",
				"templateParameters": {
					"cluster_name": "test-cluster",
					"node_count": 3,
					"node_type": "worker"
				},
				"templateVersion": "v1.1"
			}
		}
		`
		var new_data map[string]interface{}

		_ = json.Unmarshal([]byte(new_content), &new_data)
		new_val := new_data["spec"].(map[string]interface{})["templateVersion"]

		changed, _ := newResource.Compare(name, new_data, true)
		if !changed {
			t.Errorf("Expected template version to be changed, but got no change. %v -> %v", old_val, new_val)
		}

		if newResource.Spec["templateVersion"] != new_val {
			t.Errorf("Expected template version to be updated, but got not updated. %v -> %v", old_val, new_val)
		}

	})

	t.Run("Test case: Compare should return true when templateParameters is changed (and Apply the changes)", func(t *testing.T) {

		name := "test-provisioning-001"
		content := `
		{
			"apiVersion": "o2ims.provisioning.oran.org/v1alpha1",
			"kind": "ProvisioningRequest",
			"metadata": {
				"creationTimestamp": "2025-01-05T16:09:12Z",
				"generation": 1,
				"name": "test-provisioning-001",
				"resourceVersion": "365239",
				"uid": "3378091a-8260-4b84-9373-eb9b6831c3a3"
			},
			"spec": {
				"description": "Test provisioning request",
				"name": "Test Cluster Provisioning",
				"templateName": "cluster-template",
				"templateParameters": {
					"cluster_name": "test-cluster",
					"node_count": 3,
					"node_type": "worker"
				},
				"templateVersion": "v1.0"
			}
		}
		`

		var data map[string]interface{}
		_ = json.Unmarshal([]byte(content), &data)
		old_val := data["spec"].(map[string]interface{})["templateParameters"]

		newResource := ProvisioningRequest{}
		_ = newResource.SetInitFields(name, data)

		new_content := `
		{
			"apiVersion": "o2ims.provisioning.oran.org/v1alpha1",
			"kind": "ProvisioningRequest",
			"metadata": {
				"creationTimestamp": "2025-01-05T16:09:12Z",
				"generation": 1,
				"name": "test-provisioning-001",
				"resourceVersion": "365239",
				"uid": "3378091a-8260-4b84-9373-eb9b6831c3a3"
			},
			"spec": {
				"description": "Test provisioning request",
				"name": "Test Cluster Provisioning",
				"templateName": "cluster-template",
				"templateParameters": {
					"cluster_name": "test-cluster",
					"node_count": 4,
					"node_type": "worker"
				},
				"templateVersion": "v1.0"
			}
		}
		`
		var new_data map[string]interface{}
		_ = json.Unmarshal([]byte(new_content), &new_data)
		new_val := new_data["spec"].(map[string]interface{})["templateParameters"]

		changed, _ := newResource.Compare(name, new_data, true)
		if !changed {
			t.Errorf("Expected templateParameters to be changed, but got no change %v\n -> %v\n", old_val, new_val)
		}

		if !reflect.DeepEqual(newResource.Spec["templateParameters"], data["spec"].(map[string]interface{})["templateParameters"]) {
			t.Errorf("Expected template parameters to be updated,  but got updated %v\n -> %v\n", old_val, new_val)
		}
	})

	t.Run("Check for Compare when template name is changed (and Apply is true)", func(t *testing.T) {
		name := "test-provisioning-001"
		content := `
		{
			"apiVersion": "o2ims.provisioning.oran.org/v1alpha1",
			"kind": "ProvisioningRequest",
			"metadata": {
				"creationTimestamp": "2025-01-05T16:09:12Z",
				"generation": 1,
				"name": "test-provisioning-001",
				"resourceVersion": "365239",
				"uid": "3378091a-8260-4b84-9373-eb9b6831c3a3"
			},
			"spec": {
				"description": "Test provisioning request",
				"name": "Test Cluster Provisioning",
				"templateName": "cluster-template",
				"templateParameters": {
					"cluster_name": "test-cluster",
					"node_count": 3,
					"node_type": "worker"
				},
				"templateVersion": "v1.0"
			}
		}
		`

		var data map[string]interface{}
		_ = json.Unmarshal([]byte(content), &data)
		newResource := ProvisioningRequest{}
		_ = newResource.SetInitFields(name, data)
		old_val := data["spec"].(map[string]interface{})["templateName"]

		new_content := `
		{
			"apiVersion": "o2ims.provisioning.oran.org/v1alpha1",
			"kind": "ProvisioningRequest",
			"metadata": {
				"creationTimestamp": "2025-01-05T16:09:12Z",
				"generation": 1,
				"name": "test-provisioning-001",
				"resourceVersion": "365239",
				"uid": "3378091a-8260-4b84-9373-eb9b6831c3a3"
			},
			"spec": {
				"description": "Test provisioning request",
				"name": "Test Cluster Provisioning",
				"templateName": "cluster-template-new",
				"templateParameters": {
					"cluster_name": "test-cluster",
					"node_count": 3,
					"node_type": "worker"
				},
				"templateVersion": "v1.0"
			}
		}
		`
		var new_data map[string]interface{}

		_ = json.Unmarshal([]byte(new_content), &new_data)
		new_val := new_data["spec"].(map[string]interface{})["templateName"]

		fmt.Printf("old_val: %v, new_val: %v\n", old_val, new_val)

		changed, err := newResource.Compare(name, new_data, true)
		if !changed {
			t.Errorf("Expected template name to be changed, but got no change. %v -> %v , %v", old_val, new_val, err)
		}

		if newResource.Spec["templateName"] != new_val {
			t.Errorf("Expected template name to be updated, but got not updated. %v -> %v", old_val, newResource.Spec["templateName"])
		}
	})

}

func TestAllSmallFunctions(t *testing.T) {
	name := "test"

	provisioningRequest := NewProvisioningRequest(
		name,
		"o2ims.provisioning.oran.org/v1alpha1",
		"ProvisioningRequest",
		map[string]interface{}{
			"creationTimestamp": "2025-01-05T16:09:12Z",
			"generation":        1,
			"name":              "test-provisioning-001",
			"resourceVersion":   "365239",
			"uid":               "3378091a-8260-4b84-9373-eb9b6831c3a3",
		},
		map[string]interface{}{
			"description":  "Test provisioning request",
			"name":         "Test Cluster Provisioning",
			"templateName": "cluster-template2",
			"templateParameters": map[string]interface{}{
				"cluster_name": "test-cluster",
				"node_count":   3,
				"node_type":    "worker",
			},
			"templateVersion": "v1.0",
		},
		map[string]interface{}{},
	)

	//check provisioning request to be non empty
	if provisioningRequest == nil {
		t.Errorf("Expected provisioning request to be non-empty")
	}

	// test GetID
	id := provisioningRequest.GetID()
	if id != name {
		t.Errorf("Expected ID to be %s, got %s", name, id)
	}

	now_time := time.Now().Format(time.RFC3339)
	_ = now_time

	fields := map[string]interface{}{
		"creationTimestamp": "2025-01-05T16:09:12Z",
		"generation":        1,
		"name":              "test-provisioning-001",
		"resourceVersion":   "365239",
		"uid":               "3378091a-8260-4b84-9373-eb9b6831c3a3",
		"spec": map[string]interface{}{
			"description":  "Test provisioning request",
			"name":         "Test Cluster Provisioning",
			"templateName": "cluster-template2",
			"templateParameters": map[string]interface{}{
				"cluster_name": "test-cluster",
				"node_count":   3,
				"node_type":    "worker",
			},
			"templateVersion": "v1.0",
		},
		"status": map[string]interface{}{
			"provisioningStatus": map[string]interface{}{
				"provisioningMessage":    "",
				"provisioningState":      "",
				"provisioningUpdateTime": "",
			},
			"extensions": map[string]interface{}{
				"reconciliationInfo": map[string]interface{}{
					"configSetName":         "",
					"configSetCrc":          "",
					"endTime":               "",
					"lastUpdateTime":        now_time,
					"meDescription":         "test",
					"meFlavorType":          "single-server",
					"meName":                "test",
					"meProductType":         "CNIS",
					"meSwVer":               "1.15",
					"omcOperation":          "",
					"reconciliationState":   "init",
					"reconciliationTimeout": "",
					"startTime":             now_time,
					"templateName":          "",
					"templateParamsCRC":     "",
					"templateParamsApplied": false,
					"templateVersion":       "",
					"markedForDeletion":     false,
					"transitionTime":        now_time,
					"workflowId":            "",
				},
			},
		},
	}

	provisioningRequest.SetInitFields(name, fields)

	// pretty print the request
	pretty_print, err := json.MarshalIndent(provisioningRequest, "", "    ")
	_ = pretty_print
	if err != nil {
		t.Errorf("Failed to pretty print: %s", err)
	}
	//fmt.Printf("pretty print: %s\n", pretty_print)

	// test GetStatus
	status, err := provisioningRequest.GetStatus()

	//fmt.Printf("status: %v\n", status)

	_ = status
	if err != nil {
		t.Errorf("Failed to get status: %s", err)
	}

	expectedStatus := map[string]interface{}{
		"provisioningStatus": map[string]interface{}{
			"provisioningMessage":    "",
			"provisioningState":      "",
			"provisioningUpdateTime": "",
		},
		"extensions": map[string]interface{}{
			"reconciliationInfo": map[string]interface{}{
				"apiFailure":            "",
				"backOffTime":           now_time,
				"configSetCrc":          "",
				"configSetName":         "",
				"endTime":               "",
				"lastUpdateTime":        now_time,
				"meDescription":         "test",
				"meFlavorType":          "single-server",
				"meName":                "test",
				"meProductType":         "CNIS",
				"meSwVer":               "1.15",
				"omcOperation":          "",
				"reconciliationState":   "init",
				"reconciliationTimeout": "",
				"apiRetryCount":         "0",
				"startTime":             now_time,
				"subState":              "",
				"templateName":          "",
				"templateParamsCRC":     "",
				"templateParamsApplied": false,
				"templateVersion":       "",
				"markedForDeletion":     false,
				"transitionTime":        now_time,
				"workflowId":            "",
			},
		},
	}

	if !reflect.DeepEqual(status, expectedStatus) {
		t.Errorf("Expected status to be \n%v, got \n%v", expectedStatus, status)
	}

}

func TestProvisioningRequestSpecChanged(t *testing.T) {

	t.Run("Test case: specChanged - spec changed)", func(t *testing.T) {
		name := "test-provisioning-001"
		content := `
		{
			"apiVersion": "o2ims.provisioning.oran.org/v1alpha1",
			"kind": "ProvisioningRequest",
			"metadata": {
				"creationTimestamp": "2025-01-05T16:09:12Z",
				"generation": 1,
				"name": "test-provisioning-001",
				"resourceVersion": "365239",
				"uid": "3378091a-8260-4b84-9373-eb9b6831c3a3"
			},
			"spec": {
				"description": "Test provisioning request",
				"name": "Test Cluster Provisioning",
				"templateName": "cluster-template",
				"templateParameters": {
					"cluster_name": "test-cluster",
					"node_count": 3,
					"node_type": "worker"
				},
				"templateVersion": "v1.0"
			}
		}
		`

		var data map[string]interface{}
		_ = json.Unmarshal([]byte(content), &data)
		

		newResource := ProvisioningRequest{}
		_ = newResource.SetInitFields(name, data)

	})

	//FIXME TBA

}

func TestProvisioningRequestCheckManadatoryFields(t *testing.T) {

	var err error

	// Test case: checkManadatoryFields - success case

	name := "test-provisioning-001"
	content := `
	{
		"apiVersion": "o2ims.provisioning.oran.org/v1alpha1",
		"kind": "ProvisioningRequest",
		"metadata": {
			"name": "test-provisioning-001",
			"namespace": "o2ims",
			"labels": {
				"app": "o2ims-provisioning"
			}
		},
		"spec": {
			"templateName": "test-template",
			"templateVersion": "1.0.0",
			"templateParameters": {
				"param1": "value1"
			}
		},
		"status": {
			"extensions": {
				"reconciliationInfo": {
					"meName": "test",
					"meDescription": "test",
					"meProductType": "CNIS",
					"meFlavorType": "single-server",
					"meSwVer": "1.15",
					"markedForDeletion": false,
					"reconciliationState": "init"
				}
			},
			"provisioningStatus": {
				"provisioningMessage": "",
				"provisioningState": "",
				"provisioningUpdateTime": ""
			}
		}
	}
	`

	var data map[string]interface{}
	_ = json.Unmarshal([]byte(content), &data)

	pr := ProvisioningRequest{}
	err = pr.SetInitFields(name, data)

	err = pr.checkManadatoryFields()
	if err != nil {
		t.Fatalf("Expected no error  but  got %s", err)
	}

	// test delete and add templateName and version

	keyList := []string{"templateName", "templateVersion", "templateParameters"}

	for _, key := range keyList {
		tmp, _ := pr.Spec[key]
		delete(pr.Spec, key)
		err = pr.checkManadatoryFields()
		if err == nil {
			t.Fatalf("Expected error but got none")
		}
		pr.Spec[key] = tmp

		err = pr.checkManadatoryFields()
		if err != nil {
			t.Fatalf("Expected no error but got %s", err)
		}
	}

	// test delete and add extensions
	tmp, _ := pr.Status["extensions"]
	delete(pr.Status, "extensions")
	err = pr.checkManadatoryFields()
	if err == nil {
		t.Fatalf("Expected error but got none")
	}

	// test when extensions is not of type map interface
	pr.Status["extensions"] = []string{"test"}
	if err == nil {
		t.Fatalf("Expected error but got none")
	}
	pr.Status["extensions"] = tmp
	err = pr.checkManadatoryFields()
	if err != nil {
		t.Fatalf("Expected no error but got %s", err)
	}

	// test delete and add reconciliationInfo
	tmp, _ = pr.Status["extensions"].(map[string]interface{})["reconciliationInfo"]
	delete(pr.Status["extensions"].(map[string]interface{}), "reconciliationInfo")
	pr.Status["extensions"].(map[string]interface{})["reconciliationInfo"] = []string{"test"}
	err = pr.checkManadatoryFields()
	if err == nil {
		t.Fatalf("Expected error but got none")
	}

	delete(pr.Status["extensions"].(map[string]interface{}), "reconciliationInfo")
	pr.Status["extensions"].(map[string]interface{})["reconciliationInfo"] = tmp
	err = pr.checkManadatoryFields()
	if err != nil {
		t.Fatalf("Expected no error but got %s", err)
	}

	keylist := []string{"reconciliationState", "markedForDeletion",
		"templateName",
		"templateVersion",
		"templateParamsCRC",
		"templateParamsApplied",
		"workflowId",
		"configSetName",
		"configSetCrc",
		"meName",
		"meDescription",
		"meProductType",
		"meFlavorType",
		"meSwVer",
		"omcOperation",
		"startTime",
		"endTime",
		"lastUpdateTime",
		"transitionTime",
		"reconciliationTimeout",
		"markedForDeletion",
	}

	for _, key := range keylist {
		tmp = pr.Status["extensions"].(map[string]interface{})["reconciliationInfo"].(map[string]interface{})[key]
		delete(pr.Status["extensions"].(map[string]interface{})["reconciliationInfo"].(map[string]interface{}), key)

		// /fmt.Printf("key %v \n recon %v\n", key, pr.Status["extensions"].(map[string]interface{})["reconciliationInfo"].(map[string]interface{}))
		err = pr.checkManadatoryFields()
		if err == nil {
			t.Fatalf("Expected error but got none")
		}
		pr.Status["extensions"].(map[string]interface{})["reconciliationInfo"].(map[string]interface{})[key] = tmp
	}

	//check if to be deleted is of type bool
	tmp = pr.Status["extensions"].(map[string]interface{})["reconciliationInfo"].(map[string]interface{})["markedForDeletion"]
	delete(pr.Status["extensions"].(map[string]interface{})["reconciliationInfo"].(map[string]interface{}), "markedForDeletion")
	pr.Status["extensions"].(map[string]interface{})["reconciliationInfo"].(map[string]interface{})["markedForDeletion"] = "string"
	err = pr.checkManadatoryFields()
	if err == nil {
		t.Fatalf("Expected error but got none")
	}
	delete(pr.Status["extensions"].(map[string]interface{})["reconciliationInfo"].(map[string]interface{}), "markedForDeletion")
	pr.Status["extensions"].(map[string]interface{})["reconciliationInfo"].(map[string]interface{})["markedForDeletion"] = tmp

	tmp = pr.Status["provisioningStatus"]
	delete(pr.Status, "provisioningStatus")
	err = pr.checkManadatoryFields()
	if err == nil {
		t.Fatalf("Expected error but got none")
	}

	pr.Status["provisioningStatus"] = []string{"test"}
	err = pr.checkManadatoryFields()
	if err == nil {
		t.Fatalf("Expected error but got none")
	}
	delete(pr.Status, "provisioningStatus")
	pr.Status["provisioningStatus"] = tmp

	keylist = []string{"provisioningMessage", "provisioningState", "provisioningUpdateTime"}

	for _, key := range keylist {
		tmp = pr.Status["provisioningStatus"].(map[string]interface{})[key]
		delete(pr.Status["provisioningStatus"].(map[string]interface{}), key)
		err = pr.checkManadatoryFields()
		if err == nil {
			t.Fatalf("Expected error but got none")
		}
		pr.Status["provisioningStatus"].(map[string]interface{})[key] = tmp
	}
}

func TestProvisioningRequestGet(t *testing.T) {

	t.Run("Test case: GetNew returns a new instance", func(t *testing.T) {
		pr := ProvisioningRequest{}

		res := pr.GetNew()
		if res == nil {
			t.Errorf("Expected a new instance, got nil")
		}
		if _, ok := res.(*ProvisioningRequest); !ok {
			t.Errorf("Expected instance to be of type *ProvisioningRequest, got %T", res)
		}
	})

	t.Run("Test case: UpdateProvisioningStatus updates the status and GetStatus retrieves it", func(t *testing.T) {
		pr := ProvisioningRequest{
			Status: make(map[string]interface{}),
		}

		newProvisioningStatus := map[string]interface{}{
			"provisioningMessage":    "Test Message",
			"provisioningState":      "Test State",
			"provisioningUpdateTime": "2025-01-05T17:19:11Z",
		}

		err := pr.UpdateProvisioningStatus(newProvisioningStatus)
		if err != nil {
			t.Fatalf("Failed to update provisioning status: %v", err)
		}

		status, err := pr.GetStatus()
		if err != nil {
			t.Fatalf("Failed to get provisioning status: %v", err)
		}

		if status["provisioningStatus"].(map[string]interface{})["provisioningMessage"] != "Test Message" {
			t.Errorf("Expected provisioningMessage to be %s, but got %s", "Test Message", status["provisioningStatus"].(map[string]interface{})["provisioningMessage"])
		}
		if status["provisioningStatus"].(map[string]interface{})["provisioningState"] != "Test State" {
			t.Errorf("Expected provisioningState to be %s, but got %s", "Test State", status["provisioningStatus"].(map[string]interface{})["provisioningState"])
		}
		if status["provisioningStatus"].(map[string]interface{})["provisioningUpdateTime"] != "2025-01-05T17:19:11Z" {
			t.Errorf("Expected provisioningUpdateTime to be %s, but got %s", "2025-01-05T17:19:11Z", status["provisioningStatus"].(map[string]interface{})["provisioningUpdateTime"])
		}

	})

	t.Run("Test case: GetReconciliationInfo retrieves the reconciliation info", func(t *testing.T) {
		pr := ProvisioningRequest{
			Status: map[string]interface{}{
				"extensions": map[string]interface{}{
					"reconciliationInfo": map[string]interface{}{
						ReconciliationState:   "Progressing",
						MarkedForDeletion:     false,
						TemplateName:          "test-template",
						TemplateVersion:       "1.0",
						TemplateParamsCRC:     "test-crc",
						TemplateParamsApplied: false,
						WorkflowId:            "test-workflow",
						ConfigSetName:         "test-configset",
						ConfigSetCrc:          "1.0",
						MeName:                "test-me",
						MeDescription:         "test-me",
						MeProductType:         "test-type",
						MeFlavorType:          "test-flavor",
						MeSwVer:               "1.0",
						OmcOperation:          "test-op",
						StartTime:             "2025-01-05T17:19:11Z",
						EndTime:               "2025-01-05T17:19:11Z",
						LastUpdateTime:        "2025-01-05T17:19:11Z",
						TransitionTime:        "2025-01-05T17:19:11Z",
						ReconciliationTimeout: "2025-01-05T17:19:11Z",
					},
				},
			},
		}

		reconciliationInfo, err := pr.GetReconciliationInfo()
		if err != nil {
			t.Fatalf("Failed to get reconciliation info: %v", err)
		}
		if reconciliationInfo[ReconciliationState] != "Progressing" {
			t.Errorf("Expected ReconciliationState to be %s, but got %s", "Progressing", reconciliationInfo[ReconciliationState])
		}
		if reconciliationInfo[MarkedForDeletion] != false {
			t.Errorf("Expected MarkedForDeletion to be %v, but got %v", false, reconciliationInfo[MarkedForDeletion])
		}
		if reconciliationInfo[TemplateName] != "test-template" {
			t.Errorf("Expected TemplateName to be %s, but got %s", "test-template", reconciliationInfo[TemplateName])
		}
		if reconciliationInfo[TemplateVersion] != "1.0" {
			t.Errorf("Expected TemplateVersion to be %s, but got %s", "1.0", reconciliationInfo[TemplateVersion])
		}
		if reconciliationInfo[TemplateParamsCRC] != "test-crc" {
			t.Errorf("Expected TemplateParamsCRC to be %s, but got %s", "test-crc", reconciliationInfo[TemplateParamsCRC])
		}
		if reconciliationInfo[TemplateParamsApplied] != false {
			t.Errorf("Expected TemplateParamsApplied to be %v, but got %v", false, reconciliationInfo[TemplateParamsApplied])
		}
		if reconciliationInfo[WorkflowId] != "test-workflow" {
			t.Errorf("Expected WorkflowId to be %s, but got %s", "test-workflow", reconciliationInfo[WorkflowId])
		}
		if reconciliationInfo[ConfigSetName] != "test-configset" {
			t.Errorf("Expected ConfigSetName to be %s, but got %s", "test-configset", reconciliationInfo[ConfigSetName])
		}
		if reconciliationInfo[ConfigSetCrc] != "1.0" {
			t.Errorf("Expected ConfigSetCrc to be %s, but got %s", "1.0", reconciliationInfo[ConfigSetCrc])
		}
		if reconciliationInfo[MeName] != "test-me" {
			t.Errorf("Expected MeName to be %s, but got %s", "test-me", reconciliationInfo[MeName])
		}
		if reconciliationInfo[MeDescription] != "test-me" {
			t.Errorf("Expected MeDescription to be %s, but got %s", "test-me", reconciliationInfo[MeDescription])
		}
		if reconciliationInfo[MeProductType] != "test-type" {
			t.Errorf("Expected MeProductType to be %s, but got %s", "test-type", reconciliationInfo[MeProductType])
		}
		if reconciliationInfo[MeFlavorType] != "test-flavor" {
			t.Errorf("Expected MeFlavorType to be %s, but got %s", "test-flavor", reconciliationInfo[MeFlavorType])
		}
		if reconciliationInfo[MeSwVer] != "1.0" {
			t.Errorf("Expected MeSwVer to be %s, but got %s", "1.0", reconciliationInfo[MeSwVer])
		}
		if reconciliationInfo[OmcOperation] != "test-op" {
			t.Errorf("Expected OmcOperation to be %s, but got %s", "test-op", reconciliationInfo[OmcOperation])
		}
		if reconciliationInfo[StartTime] != "2025-01-05T17:19:11Z" {
			t.Errorf("Expected StartTime to be %s, but got %s", "2025-01-05T17:19:11Z", reconciliationInfo[StartTime])
		}
		if reconciliationInfo[EndTime] != "2025-01-05T17:19:11Z" {
			t.Errorf("Expected EndTime to be %s, but got %s", "2025-01-05T17:19:11Z", reconciliationInfo[EndTime])
		}
		if reconciliationInfo[TransitionTime] != "2025-01-05T17:19:11Z" {
			t.Errorf("Expected TransitionTime to be %s, but got %s", "2025-01-05T17:19:11Z", reconciliationInfo[TransitionTime])
		}

		reconciliationInfo2 := map[string]interface{}{
			ReconciliationState:   "test-state",
			MarkedForDeletion:     true,
			TemplateName:          "test-template2",
			TemplateVersion:       "2.0",
			TemplateParamsCRC:     "test-crc2",
			TemplateParamsApplied: false,
			WorkflowId:            "test-workflow2",
			ConfigSetName:         "test-configset2",
			ConfigSetCrc:          "2.0",
			MeName:                "test-me2",
			MeDescription:         "test-me2",
			MeProductType:         "test-type2",
			MeFlavorType:          "test-flavor2",
			MeSwVer:               "2.0",
			OmcOperation:          "test-op2",
			StartTime:             "2025-01-05T17:19:12Z",
			EndTime:               "2025-01-05T17:19:12Z",
			TransitionTime:        "2025-01-05T17:19:12Z",
			ReconciliationTimeout: "10m",
			LastUpdateTime:        "2025-01-05T17:19:12Z",
		}

		err = pr.updateReconcileInfo(reconciliationInfo2)
		if err != nil {
			t.Errorf("Expected nil, but got %v", err)
		}
		status, err := pr.GetStatus()
		if err != nil {
			t.Errorf("Expected nil, but got %v", err)
		}
		reconciliationInfo3, err := pr.GetReconciliationInfo()
		if err != nil {
			t.Errorf("Expected nil, but got %v", err)
		}
		if !reflect.DeepEqual(reconciliationInfo3, reconciliationInfo2) {
			t.Errorf("Expected %v, but got %v", reconciliationInfo2, reconciliationInfo3)
		}
		if !reflect.DeepEqual(status["extensions"].(map[string]interface{})["reconciliationInfo"], reconciliationInfo2) {
			t.Errorf("Expected %v, but got %v", reconciliationInfo2, status["extensions"].(map[string]interface{})["reconciliationInfo"])
		}

		pr = ProvisioningRequest{
			Status: map[string]interface{}{
				"extensions": map[string]interface{}{},
			},
		}

		reconciliationInfo4 := map[string]interface{}{
			ReconciliationState:   "test-state",
			MarkedForDeletion:     true,
			TemplateName:          "test-template2",
			TemplateVersion:       "2.0",
			TemplateParamsCRC:     "test-crc2",
			TemplateParamsApplied: false,
			WorkflowId:            "test-workflow2",
			ConfigSetName:         "test-configset2",
			ConfigSetCrc:          "2.0",
			MeName:                "test-me2",
			MeDescription:         "test-me2",
			MeProductType:         "test-type2",
			MeFlavorType:          "test-flavor2",
			MeSwVer:               "2.0",
			OmcOperation:          "test-op2",
			StartTime:             "2025-01-05T17:19:12Z",
			EndTime:               "2025-01-05T17:19:12Z",
			TransitionTime:        "2025-01-05T17:19:12Z",
			ReconciliationTimeout: "10m",
			LastUpdateTime:        "2025-01-05T17:19:12Z",
		}

		err = pr.updateReconcileInfo(reconciliationInfo4)
		if err != nil {
			t.Errorf("Expected nil, but got %v", err)
		}

		reconciliationInfo5, err := pr.GetReconciliationInfo()
		if err != nil {
			t.Errorf("Expected nil, but got %v", err)
		}

		if !reflect.DeepEqual(reconciliationInfo5, reconciliationInfo4) {
			t.Errorf("Expected %v, but got %v", reconciliationInfo2, reconciliationInfo3)
		}

		// Extension is not present as well
		pr = ProvisioningRequest{
			Status: map[string]interface{}{},
		}

		reconciliationInfo4 = map[string]interface{}{
			ReconciliationState:   "test-state",
			MarkedForDeletion:     true,
			TemplateName:          "test-template2",
			TemplateVersion:       "2.0",
			TemplateParamsCRC:     "test-crc2",
			TemplateParamsApplied: false,
			WorkflowId:            "test-workflow2",
			ConfigSetName:         "test-configset2",
			ConfigSetCrc:          "2.0",
			MeName:                "test-me2",
			MeDescription:         "test-me2",
			MeProductType:         "test-type2",
			MeFlavorType:          "test-flavor2",
			MeSwVer:               "2.0",
			OmcOperation:          "test-op2",
			StartTime:             "2025-01-05T17:19:12Z",
			EndTime:               "2025-01-05T17:19:12Z",
			TransitionTime:        "2025-01-05T17:19:12Z",
			ReconciliationTimeout: "10m",
			LastUpdateTime:        "2025-01-05T17:19:12Z",
		}

		err = pr.updateReconcileInfo(reconciliationInfo4)
		if err != nil {
			t.Errorf("Expected nil, but got %v", err)
		}

		reconciliationInfo5, err = pr.GetReconciliationInfo()
		if err != nil {
			t.Errorf("Expected nil, but got %v", err)
		}

		if !reflect.DeepEqual(reconciliationInfo5, reconciliationInfo4) {
			t.Errorf("Expected %v, but got %v", reconciliationInfo2, reconciliationInfo3)
		}

	})

	t.Run("Test case: GetTemplateNameVersionAndParams retrieves the template name, version and params", func(t *testing.T) {
		pr := ProvisioningRequest{
			Spec: map[string]interface{}{
				"templateName":       "test-template",
				"templateVersion":    "1.0",
				"templateParameters": map[string]interface{}{"param1": "value1"},
			},
		}

		templateName, templateVersion, templateParameters, err := pr.GetTemplateNameVersionAndParams(pr.Spec)
		if err != nil {
			t.Errorf("Expected nil, but got %v", err)
		}
		if templateName != "test-template" {
			t.Errorf("Expected templateName to be %s, but got %s", "test-template", templateName)
		}
		if templateVersion != "1.0" {
			t.Errorf("Expected templateVersion to be %s, but got %s", "1.0", templateVersion)
		}
		if !reflect.DeepEqual(templateParameters, map[string]interface{}{"param1": "value1"}) {
			t.Errorf("Expected templateParameters to be %v, but got %v", map[string]interface{}{"param1": "value1"}, templateParameters)
		}

		pr = ProvisioningRequest{
			Spec: map[string]interface{}{
				"templateVersion":    "1.0",
				"templateParameters": map[string]interface{}{"param1": "value1"},
			},
		}

		templateName, templateVersion, templateParameters, err = pr.GetTemplateNameVersionAndParams(pr.Spec)
		expected_error := "templateName not found"
		if err == nil || err.Error() != expected_error {
			t.Errorf("Expected error %s, but got %v", expected_error, err)
		}

		pr = ProvisioningRequest{
			Spec: map[string]interface{}{
				"templateName":       "test-template",
				"templateParameters": map[string]interface{}{"param1": "value1"},
			},
		}

		templateName, templateVersion, templateParameters, err = pr.GetTemplateNameVersionAndParams(pr.Spec)
		expected_error = "templateVersion not found"
		if err == nil || err.Error() != expected_error {
			t.Errorf("Expected error %s, but got %v", expected_error, err)
		}

		pr = ProvisioningRequest{
			Spec: map[string]interface{}{
				"templateName":    "test-template",
				"templateVersion": "1.0",
			},
		}

		templateName, templateVersion, templateParameters, err = pr.GetTemplateNameVersionAndParams(pr.Spec)
		expected_error = "templateParameters not found"
		if err == nil || err.Error() != expected_error {
			t.Errorf("Expected error %s, but got %v", expected_error, err)
		}

	})

	t.Run("Test case: GetDeleteFlag returns the value of MarkedForDeletion in the reconciliation info", func(t *testing.T) {

		pr := ProvisioningRequest{
			Status: map[string]interface{}{
				"extensions": map[string]interface{}{
					"reconciliationInfo": map[string]interface{}{
						"markedForDeletion": true,
					},
				},
			},
		}

		if pr.GetDeleteFlag() != true {
			t.Errorf("Expected markedForDeletion to be true, but got false")
		}

	})

	t.Run("Test case: SetDeleteFlag sets the value of markedForDeletion in the reconciliation info", func(t *testing.T) {

		pr := ProvisioningRequest{
			Status: map[string]interface{}{
				"extensions": map[string]interface{}{
					"reconciliationInfo": map[string]interface{}{
						"markedForDeletion": true,
					},
				},
			},
		}

		if err := pr.SetDeleteFlag(); err != nil {
			t.Errorf("Expected SetDeleteFlag to return nil, but got %v", err)
		}

		if pr.GetDeleteFlag() != true {
			t.Errorf("Expected markedForDeletion to be true, but got false")
		}

		/* delete the tobedleted flag */
		delete(pr.Status["extensions"].(map[string]interface{})["reconciliationInfo"].(map[string]interface{}), "markedForDeletion")

		if pr.GetDeleteFlag() != false {
			t.Errorf("Expected markedForDeletion to be true, but got false")
		}

	})

	t.Run("Test case: GetReconciliationInfo should return empty", func(t *testing.T) {

		pr := ProvisioningRequest{
			Status: map[string]interface{}{},
		}

		recon, err := pr.GetReconciliationInfo()

		if err != nil {
			t.Errorf("Expected GetReconciliationInfo succeed, but got err %v", err)
		}

		if len(recon) != 0 {
			t.Errorf("Expected GetReconciliationInfo to return an empty map, but got %v", recon)
		}
	})

	t.Run("Test case: GetReconciliationInfo should should return empty reconciliationInfo map is missing", func(t *testing.T) {

		pr := ProvisioningRequest{
			Status: map[string]interface{}{
				"extensions": map[string]interface{}{},
			},
		}

		recon, err := pr.GetReconciliationInfo()

		if err != nil {
			t.Errorf("Expected GetReconciliationInfo succeed, but got err %v", err)
		}

		if len(recon) != 0 {
			t.Errorf("Expected GetReconciliationInfo to return an empty map, but got %v", recon)
		}

	})

	t.Run("Test case: GetReconciliationInfo with empty string even if the field is empty", func(t *testing.T) {

		pr := ProvisioningRequest{
			Status: map[string]interface{}{
				"extensions": map[string]interface{}{},
			},
		}
		reconciliationInfo, err := pr.GetReconciliationInfo()
		if err != nil {
			t.Errorf("Expected GetReconciliationInfo to return nil, but got %v", err)
		}

		if reconciliationInfo == nil {
			t.Errorf("Expected GetReconciliationInfo to return a non-nil map, but got nil")
		}

		if len(reconciliationInfo) != 0 {
			t.Errorf("Expected GetReconciliationInfo to return an empty map but got %v", reconciliationInfo)
		}

	})
}

func TestProvisioningRequestReconcileInit(t *testing.T) {

	t.Run("Test case: ME Create when API error is there for unknown reason", func(t *testing.T) {

		var err error
		_ = err
		name := "test-provisioning-001"
		content := `
			{
				"apiVersion": "o2ims.provisioning.oran.org/v1alpha1",
				"kind": "ProvisioningRequest",
				"metadata": {
					"creationTimestamp": "2025-01-05T16:09:12Z",
					"generation": 1,
					"name": "test-provisioning-001",
					"resourceVersion": "365239",
					"uid": "3378091a-8260-4b84-9373-eb9b6831c3a3"
				},
				"spec": {
					"description": "Test provisioning request",
					"name": "Test Cluster Provisioning",
					"templateName": "cluster-template",
					"templateParameters": {
						"cluster_name": "test-cluster",
						"node_count": 3,
						"node_type": "worker",
						"meAddtionalParams": {
							"meCreateFailRetry": 3
						}
					},
					"templateVersion": "v1.0"
				}
			}
			`
		var data map[string]interface{}
		_ = json.Unmarshal([]byte(content), &data)

		newResource := ProvisioningRequest{}

		err = newResource.SetInitFields(name, data)
		if err != nil {
			t.Errorf("SetInitFields() error = %v", err)
		}

		MockOMCRemoteService := omc_rest.NewMockOMCRemoteService()
		newResource.remoteService = MockOMCRemoteService

		var ErrServiceUnavailable = errors.New("service unavailable")
		MockOMCRemoteService.SetAPIError("CreateME", errors.New("service unavailable"), 2, 0)

		_ = newResource.Reconcile()
		rInfo, _ := newResource.GetReconciliationInfo()
		retryCount, _ := strconv.Atoi(rInfo[ApiRetryCount].(string))
		assert.Equal(t, retryCount, 1, "retry count should be 1")
		assert.Equal(t, rInfo[SubState], "CreateME", "substate should be CreateME")
		assert.Equal(t, rInfo[ApiFailure], ErrServiceUnavailable.Error(), "service unavailable")
		assert.Equal(t, rInfo[ReconciliationState], "init", "reconciliation state should be init")

		_ = newResource.Reconcile()
		rInfo, _ = newResource.GetReconciliationInfo()
		retryCount, _ = strconv.Atoi(rInfo[ApiRetryCount].(string))
		assert.Equal(t, retryCount, 2, "retry count should be 2")
		assert.Equal(t, rInfo[SubState], "CreateME", "substate should be CreateME")
		assert.Equal(t, rInfo[ApiFailure], ErrServiceUnavailable.Error(), "service unavailable")
		assert.Equal(t, rInfo[ReconciliationState], "init", "reconciliation state should be init")

		MockOMCRemoteService.SetAPIError("GetME", errors.New("service unavailable"), 1, 1)
		//Note: createMEWithParams will be called getME to we want it to pass hense we skip once
		_ = newResource.Reconcile()
		rInfo, _ = newResource.GetReconciliationInfo()
		retryCount, _ = strconv.Atoi(rInfo[ApiRetryCount].(string))
		assert.Equal(t, retryCount, 1, "retry count should be 1")
		assert.Equal(t, rInfo[SubState], "WaitingOnME", "substate should be WaitingOnME")
		assert.Equal(t, rInfo[ApiFailure], ErrServiceUnavailable.Error(), "service unavailable")
		assert.Equal(t, rInfo[ReconciliationState], "init", "reconciliation state should be init")

		MockOMCRemoteService.SetAPIError("GetME", errors.New("service unavailable"), 1, 0)
		//Note: createMEWithParams is not invoked this time so no need to skip
		_ = newResource.Reconcile()
		rInfo, _ = newResource.GetReconciliationInfo()
		retryCount, _ = strconv.Atoi(rInfo[ApiRetryCount].(string))
		assert.Equal(t, retryCount, 2, "retry count should be 2")
		assert.Equal(t, rInfo[SubState], "WaitingOnME", "substate should be WaitingOnME")
		assert.Equal(t, rInfo[ApiFailure], ErrServiceUnavailable.Error(), "service unavailable")
		assert.Equal(t, rInfo[ReconciliationState], "init", "reconciliation state should be init")

		_ = newResource.Reconcile()
		rInfo, _ = newResource.GetReconciliationInfo()
		retryCount, _ = strconv.Atoi(rInfo[ApiRetryCount].(string))
		assert.Equal(t, retryCount, 0, "retry count should be 0")
		assert.Equal(t, rInfo[SubState], "", "substate should be \"\"")
		assert.Equal(t, rInfo[ApiFailure], "", "api failure should be \"\"")
		assert.Equal(t, rInfo[ReconciliationState], "provisioning", "reconciliation state should be init")
		_ = newResource.Reconcile()
	})

	t.Run("Test case: ME Create is alredy present ", func(t *testing.T) {
		//ADD ME
	})

	t.Run("Test case: ME Create is deleted even before it goes out of the init state ", func(t *testing.T) {
		//ADD ME
	})

}

func TestProvisioningRequestReconcileInstallUpgrade(t *testing.T) {

	var enableDumpMessageAndInfo = true

	t.Run("Check for Entry into Reconcile Install from Init", func(t *testing.T) {

		var err error
		_ = err
		name := "test-provisioning-001"
		content := `
		{
			"apiVersion": "o2ims.provisioning.oran.org/v1alpha1",
			"kind": "ProvisioningRequest",
			"metadata": {
				"creationTimestamp": "2025-01-05T16:09:12Z",
				"generation": 1,
				"name": "test-provisioning-001",
				"resourceVersion": "365239",
				"uid": "3378091a-8260-4b84-9373-eb9b6831c3a3"
			},
			"spec": {
				"description": "Test provisioning request",
				"name": "Test Cluster Provisioning",
				"templateName": "cluster-template",
				"templateParameters": {
					"cluster_name": "test-cluster",
					"node_count": 3,
					"node_type": "worker",
					"meAddtionalParams": {
						"meCreateFailRetry": 3
					}
				},
				"templateVersion": "v1.0"
			}
		}
		`
		var data map[string]interface{}
		_ = json.Unmarshal([]byte(content), &data)

		newResource := ProvisioningRequest{}

		err = newResource.SetInitFields(name, data)
		if err != nil {
			t.Errorf("SetInitFields() error = %v", err)
		}

		MockOMCRemoteService := omc_rest.NewMockOMCRemoteService()
		newResource.remoteService = MockOMCRemoteService
		fmt.Printf("\n\n")

		// state:installing -- Begins from init state
		_ = newResource.Reconcile()
		rInfo, _ := newResource.GetReconciliationInfo()
		//We just exited from init state and entering into Installing State
		assert.Equal(t, "", rInfo["subState"].(string), "subState is not empty")
		assert.Equal(t, "provisioning", rInfo["reconciliationState"].(string), "reconciliationState is not Provisioning")
		assert.Equal(t, "", rInfo["apiFailure"].(string), "apiFailure is not empty")
		assert.Equal(t, "0", rInfo["apiRetryCount"].(string), "apiRetryCount is not 0")
		assert.Equal(t, "", rInfo["workflowId"].(string), "workflowId is not empty")
		if enableDumpMessageAndInfo {
			fmt.Printf("Info: test to check if entered the Provisioning state (that's why subState is empty)\n")
			dumpMessageAndInfo(&newResource, t)
		}

		// In this case both CreateConfig and PushConfig are executed in the same
		// reconcile iteration. This is because the CreateConfig and PushConfig
		// happen in the same iteration. Hence we do not expect the subState to
		// move to PushConfig.
		//
		// state:installing -- InstallUpdateCreateConfig & InstallUpdatePushConfig are
		//                     executed in the same reconcile iteration. and Proceeds to
		//                     InstallUpdatePushConfig
		_ = newResource.Reconcile()
		rInfo, _ = newResource.GetReconciliationInfo()
		assert.Equal(t, "provisioning", rInfo["reconciliationState"].(string), "reconciliationState is not Provisioning")
		assert.Equal(t, "OperationStart", rInfo["subState"].(string), "subState is not empty")
		assert.Equal(t, "", rInfo["omcOperation"].(string), "omcOperation is not Deploy")
		if enableDumpMessageAndInfo {
			fmt.Printf("Info: test case to check ProvisioningCreateConfig and ProvisioningPushConfig " +
				" are completed in the same reconcile iteration\n")
			dumpMessageAndInfo(&newResource, t)
		}

		// state:installing - OperationStart completes and enters into OperationMonitor
		// After this reconcile, we have performed the first Deploy operation
		_ = newResource.Reconcile()
		rInfo, _ = newResource.GetReconciliationInfo()
		assert.Equal(t, "provisioning", rInfo["reconciliationState"].(string), "reconciliationState is not Provisioning")
		assert.Equal(t, "OperationMonitor", rInfo["subState"].(string), "subState is not empty")
		assert.Equal(t, "deploy", rInfo["omcOperation"].(string), "omcOperation is not Deploy")
		if enableDumpMessageAndInfo {
			fmt.Printf("Info: test case to check OperationStart (deploy) is completed and next state is OperationMonitor \n")
			dumpMessageAndInfo(&newResource, t)
		}

		state := map[string]interface{}{
			"administrative": "locked",
			"operational":    "install",
		}
		//After the OperationStart MEs Admin State will change to locked and Operational State will change to install
		MockOMCRemoteService.SetMeStatus("test-provisioning-001", state)

		// state:installing - OperationStart completes and enters into OperationMonitor
		// we continue to monitor the operation
		_ = newResource.Reconcile()
		_ = newResource.Reconcile()
		rInfo, _ = newResource.GetReconciliationInfo()
		assert.Equal(t, "provisioning", rInfo["reconciliationState"].(string), "reconciliationState is not Provisioning")
		assert.Equal(t, "OperationMonitor", rInfo["subState"].(string), "subState is not empty")
		assert.Equal(t, "deploy", rInfo["omcOperation"].(string), "omcOperation is not Deploy")
		if enableDumpMessageAndInfo {
			fmt.Printf("Info:: test case to check OperationMonitor (deploy) continues unless managed element is unlocked\n")
			dumpMessageAndInfo(&newResource, t)
		}

		state = map[string]interface{}{
			"administrative": "locked",
			"operational":    "ready",
		}
		//
		MockOMCRemoteService.SetMeStatus("test-provisioning-001", state)

		// we continue to monitor the operation
		// state:installing - OperationMonitor
		_ = newResource.Reconcile()
		rInfo, _ = newResource.GetReconciliationInfo()
		assert.Equal(t, "provisioning", rInfo["reconciliationState"].(string), "reconciliationState is not Provisioning")
		assert.Equal(t, "OperationMonitor", rInfo["subState"].(string), "subState is not empty")
		assert.Equal(t, "deploy", rInfo["omcOperation"].(string), "omcOperation is not Deploy")
		if enableDumpMessageAndInfo {
			fmt.Printf("Info: In this test, we check if managed element is locked and Ready is still being monitored we need unlocked and ready\n")
			dumpMessageAndInfo(&newResource, t)
		}
		state = map[string]interface{}{
			"administrative": "unlocked",
			"operational":    "ready",
		}
		MockOMCRemoteService.SetMeStatus("test-provisioning-001", state)

		// we continue to monitor the operation
		// state:installing - OperationMonitor -> completed
		_ = newResource.Reconcile()
		//Now we should be in completed state
		rInfo, _ = newResource.GetReconciliationInfo()
		assert.Equal(t, "completed", rInfo["reconciliationState"].(string), "reconciliationState is not completed")
		assert.Equal(t, "", rInfo["subState"].(string), "subState is not empty")
		assert.Equal(t, "", rInfo["omcOperation"].(string), "omcOperation is not Deploy")
		if enableDumpMessageAndInfo {
			fmt.Printf("Info: In this test, we check if managed element is unlocked & Install is completed\n")
			dumpMessageAndInfo(&newResource, t)
		}

		//if nothing changes we will remain in completed state
		// we continue to remain in completed
		// state:completed - completed
		_ = newResource.Reconcile()
		_ = newResource.Reconcile()
		rInfo, _ = newResource.GetReconciliationInfo()
		assert.Equal(t, "completed", rInfo["reconciliationState"].(string), "reconciliationState is not completed")
		assert.Equal(t, "", rInfo["subState"].(string), "subState is not empty")
		assert.Equal(t, "", rInfo["omcOperation"].(string), "omcOperation is not empty")
		if enableDumpMessageAndInfo {
			fmt.Printf("Info: In this test, we check if multiple reconcile calls result in the same state\n")
			dumpMessageAndInfo(&newResource, t)
		}

		//Simulate Spec Change and  while in stable state
		newResource.Spec["templateParameters"].(map[string]interface{})["node_count"] = 5

		// we continue to transtion to updating
		// state:completed -(spec updated)  updating
		_ = newResource.Reconcile()
		rInfo, _ = newResource.GetReconciliationInfo()
		//We just exited from completed state and entering into Updating State
		assert.Equal(t, "", rInfo["subState"].(string), "subState is not empty")
		assert.Equal(t, "provisioning", rInfo["reconciliationState"].(string), "reconciliationState is not provisioning")
		assert.Equal(t, "", rInfo["apiFailure"].(string), "apiFailure is not empty")
		assert.Equal(t, "0", rInfo["apiRetryCount"].(string), "apiRetryCount is not 0")
		assert.Equal(t, "", rInfo["workflowId"].(string), "workflowId is not empty")
		if enableDumpMessageAndInfo {
			fmt.Println("Info: In this test, we changed the spec while we were in the completed state." +
				" (transitioning to provisioning state)")
			dumpMessageAndInfo(&newResource, t)
		}

		//state:updating -(spec updated) InstallUpdateCreateConfig & InstallUpdatePushConfig are over
		_ = newResource.Reconcile()
		rInfo, _ = newResource.GetReconciliationInfo()
		assert.Equal(t, "OperationStart", rInfo["subState"].(string), "subState is empty")
		assert.Equal(t, "provisioning", rInfo["reconciliationState"].(string), "reconciliationState is not provisioning")
		assert.Equal(t, "", rInfo["apiFailure"].(string), "apiFailure is not empty")
		assert.Equal(t, "0", rInfo["apiRetryCount"].(string), "apiRetryCount is not 0")
		assert.Equal(t, "", rInfo["workflowId"].(string), "workflowId is not empty")
		if enableDumpMessageAndInfo {
			fmt.Println("Info: in this test, we simulate a successful configuration upload and then " +
				"the next reconciliation will trigger the managed element run operation")
			dumpMessageAndInfo(&newResource, t)
		}

		state = map[string]interface{}{
			"administrative": "unlocked",
			"operational":    "ready",
		}
		MockOMCRemoteService.SetMeStatus("test-provisioning-001", state)

		//state:updating  OperationStart -- Begins
		_ = newResource.Reconcile()
		rInfo, _ = newResource.GetReconciliationInfo()
		assert.Equal(t, "update", rInfo["omcOperation"].(string), "omcOperation is not Update")
		assert.Equal(t, "OperationMonitor", rInfo["subState"].(string), "subState is empty")
		assert.Equal(t, "provisioning", rInfo["reconciliationState"].(string), "reconciliationState is not provisioning")
		assert.Equal(t, "", rInfo["apiFailure"].(string), "apiFailure is not empty")
		assert.Equal(t, "0", rInfo["apiRetryCount"].(string), "apiRetryCount is not 0")
		assert.NotEqual(t, "", rInfo["workflowId"].(string), "workflowId is empty")
		if enableDumpMessageAndInfo {
			fmt.Printf("Info: in this test, we simulate the managed element udpate has started\n")
			dumpMessageAndInfo(&newResource, t)
		}

		//Simulate managed element is in progress
		state = map[string]interface{}{
			"administrative": "locked",
			"operational":    "upgrade",
		}
		MockOMCRemoteService.SetMeStatus("test-provisioning-001", state)

		//state:updating OperationMonitor -- Begins
		_ = newResource.Reconcile()
		_ = newResource.Reconcile()
		rInfo, _ = newResource.GetReconciliationInfo()

		assert.Equal(t, "OperationMonitor", rInfo["subState"].(string), "subState is empty")
		assert.Equal(t, "provisioning", rInfo["reconciliationState"].(string), "reconciliationState is not provisioning")
		assert.Equal(t, "", rInfo["apiFailure"].(string), "apiFailure is not empty")
		assert.Equal(t, "0", rInfo["apiRetryCount"].(string), "apiRetryCount is not 0")
		assert.NotEqual(t, "", rInfo["workflowId"].(string), "workflowId is empty")

		if enableDumpMessageAndInfo {
			fmt.Printf("Info: In this test, we check if multiple reconcile calls result in the same state\n")
			dumpMessageAndInfo(&newResource, t)
		}

		// User changes the spec again :(
		// While update is ongoing user tries to change the spec again so we have to chuck out to pendingForPrevious state
		//state:updating -> PendingForPrevious
		// The Updated CRC is picked by the reconcile info but not used
		newResource.Spec["templateParameters"].(map[string]interface{})["node_count"] = 10
		crc := paramCRC(newResource.Spec["templateParameters"].(map[string]interface{}))
		_ = crc

		//OperationMonitor -- finds spec change and next is set to PendingForPrevious
		_ = newResource.Reconcile()
		rInfo, _ = newResource.GetReconciliationInfo()
		assert.Equal(t, "", rInfo["subState"].(string), "subState is empty")
		assert.Equal(t, "pendingForPrevious", rInfo["reconciliationState"].(string), "reconciliationState is not updating")
		assert.Equal(t, "", rInfo["apiFailure"].(string), "apiFailure is not empty")
		assert.Equal(t, "0", rInfo["apiRetryCount"].(string), "apiRetryCount is not 0")
		assert.NotEqual(t, "", rInfo["workflowId"].(string), "workflowId is empty")
		assert.Equal(t, crc, rInfo["templateParamsCRC"].(string), "template param crc")

		if enableDumpMessageAndInfo {
			fmt.Printf("Info: User changes the spec again :( -- While update is ongoing user tries to change the spec again so we have to chuck out to pendingForPrevious state\n")
			dumpMessageAndInfo(&newResource, t)
		}

		//state:PendingForPrevious --> PendingForPrevious
		// multiple reconcile calls result in the same state
		// even is this stae more spec changes are ignored as they
		// they may have not value unless we start update again
		newResource.Spec["templateParameters"].(map[string]interface{})["node_count"] = 23
		old_crc := crc
		_ = old_crc
		//crc = paramCRC(newResource.Spec["templateParameters"].(map[string]interface{}))

		//state:PendingForPrevious --> PendingForPrevious
		_ = newResource.Reconcile()
		_ = newResource.Reconcile()
		rInfo, _ = newResource.GetReconciliationInfo()

		assert.Equal(t, "", rInfo["subState"].(string), "subState is empty")
		assert.Equal(t, "pendingForPrevious", rInfo["reconciliationState"].(string), "reconciliationState is not updating")
		assert.Equal(t, "", rInfo["apiFailure"].(string), "apiFailure is not empty")
		assert.Equal(t, "0", rInfo["apiRetryCount"].(string), "apiRetryCount is not 0")
		assert.NotEqual(t, "", rInfo["workflowId"].(string), "workflowId is empty")
		assert.Equal(t, old_crc, rInfo["templateParamsCRC"].(string), "template param crc")

		if enableDumpMessageAndInfo {
			fmt.Printf("Info: in this test we check if we retain in the pendingForPrevious state. even if there are changes in the spec we ignore \n")
			dumpMessageAndInfo(&newResource, t)
		}

		newResource.Spec["templateParameters"].(map[string]interface{})["node_count"] = 24
		crc = paramCRC(newResource.Spec["templateParameters"].(map[string]interface{}))
		_ = crc

		state = map[string]interface{}{
			"administrative": "unlocked",
			"operational":    "error",
		}
		MockOMCRemoteService.SetMeStatus("test-provisioning-001", state)

		//PendingForPrevious -- install
		_ = newResource.Reconcile()
		rInfo, _ = newResource.GetReconciliationInfo()

		assert.Equal(t, "", rInfo["subState"].(string), "subState is empty")
		assert.Equal(t, "provisioning", rInfo["reconciliationState"].(string), "reconciliationState is not provisioning")
		assert.Equal(t, "", rInfo["apiFailure"].(string), "apiFailure is not empty")
		assert.Equal(t, "0", rInfo["apiRetryCount"].(string), "apiRetryCount is not 0")
		assert.Equal(t, "", rInfo["workflowId"].(string), "workflowId is empty")

		if enableDumpMessageAndInfo {
			fmt.Printf("Info: in this test, we check ot te take us to either delete or isntall state and the latest configset is used \n")
			dumpMessageAndInfo(&newResource, t)
		}

		//installing -- ensure latest spec is used
		_ = newResource.Reconcile()
		rInfo, _ = newResource.GetReconciliationInfo()

		assert.Equal(t, "OperationStart", rInfo["subState"].(string), "subState is empty")
		assert.Equal(t, "provisioning", rInfo["reconciliationState"].(string), "reconciliationState is not provisioning")
		assert.Equal(t, "", rInfo["apiFailure"].(string), "apiFailure is not empty")
		assert.Equal(t, "0", rInfo["apiRetryCount"].(string), "apiRetryCount is not 0")
		assert.Equal(t, "", rInfo["workflowId"].(string), "workflowId is empty")
		assert.Equal(t, crc, rInfo["templateParamsCRC"].(string), "template param crc")
		if enableDumpMessageAndInfo {
			fmt.Printf("Info: in this test, we check to check if the laest configset is used once we come from pending for previois state \n")
			dumpMessageAndInfo(&newResource, t)
		}

		//installing -- next is operation monitor
		_ = newResource.Reconcile()
		if enableDumpMessageAndInfo {
			fmt.Printf("Info: tXXXXn  \n")
			dumpMessageAndInfo(&newResource, t)
		}

		state = map[string]interface{}{
			"administrative": "locked",
			"operational":    "upgrade",
		}
		MockOMCRemoteService.SetMeStatus("test-provisioning-001", state)

		//installing -- in operation monitor
		_ = newResource.Reconcile()
		_ = newResource.Reconcile()
		rInfo, _ = newResource.GetReconciliationInfo()
		_ = rInfo

		if enableDumpMessageAndInfo {
			fmt.Printf("Info: in this test we just monitor the operation  \n")
			dumpMessageAndInfo(&newResource, t)
		}

		//installing -- failed
		state = map[string]interface{}{
			"administrative": "unlocked",
			"operational":    "error",
		}
		MockOMCRemoteService.SetMeStatus("test-provisioning-001", state)

		//installing -- in detected failure
		_ = newResource.Reconcile()
		rInfo, _ = newResource.GetReconciliationInfo()
		_ = rInfo

		if enableDumpMessageAndInfo {
			fmt.Printf("Info: asfadfasdfd \n")
			dumpMessageAndInfo(&newResource, t)
		}
	})
	//TBD add some tests on config failures etc
}

func paramCRC(params map[string]interface{}) string {
	table := crc32.MakeTable(crc32.IEEE)
	crc := crc32.Checksum([]byte(fmt.Sprintf("%v", params)), table)
	return strconv.FormatUint(uint64(crc), 10)
}

func dumpMessageAndInfo(newResource *ProvisioningRequest, t *testing.T) {

	status := newResource.Status
	provisioningStatus := status["provisioningStatus"].(map[string]interface{})
	message := provisioningStatus["provisioningMessage"].(string)

	rInfo, _ := newResource.GetReconciliationInfo()
	jsonString, err := json.MarshalIndent(rInfo, "", "    ")
	if err != nil {
		t.Fatalf("json.MarshalIndent() error = %v", err)
	}
	fmt.Printf("rInfo:\n%s\n", jsonString)
	fmt.Printf("message: %s\n\n", message)

}
