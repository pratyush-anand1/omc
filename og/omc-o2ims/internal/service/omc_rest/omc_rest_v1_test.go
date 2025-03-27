package omc_rest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v2"
)

const (
	username      = "apiadmin"
	password      = "Ericsson@123"
	url22         = "http://localhost:8080"
	skipTLSVerify = true
)

func TestOMCRemoteService(t *testing.T) {
	t.Skip("Skipping")
	testME := "testme123789"
	configsetName := "test-config-set"

	service := NewOMCRemoteService(url22, username, password, skipTLSVerify)
	t.Run("TestFetchAuthToken", func(t *testing.T) {
		err := service.FetchAuthToken()
		if err != nil {
			t.Errorf("FetchAuthToken failed: %v", err)
		}
	})
	t.Run("TestCreateME", func(t *testing.T) {

		testMeDescription := "This is a Test ME for API testing"
		testMeproductType := "CNIS"
		testMeflavor := "single-server"

		payload := make(map[string]string)
		var err error
		payload["name"] = testME
		payload["description"] = testMeDescription
		payload["product"] = testMeproductType
		payload["flavor"] = testMeflavor

		err = service.CreateME(payload)
		if err != nil {
			t.Errorf("CreateME failed: %v", err)
		}
		fmt.Printf("testME: %s\n", testME)
	})

	t.Run("TestGetME", func(t *testing.T) {
		me, err := service.GetME(testME)
		_ = me

		if err != nil {
			t.Errorf("GetME failed: %v", err)
		}

		data, err := json.MarshalIndent(me, "", "    ")
		if err != nil {
			t.Errorf("json.Marshal failed: %v", err)
		}
		fmt.Printf("me:\n%s\n", data)

	})
	t.Run("TestCreateConfigSet", func(t *testing.T) {
		payload := make(map[string]string)
		var err error

		payload["configSetName"] = configsetName
		payload["swVersion"] = "1.15"
		payload["description"] = "This is a test configset"

		_, err = service.CreateConfigSet(testME, payload)
		if err != nil {
			t.Errorf("CreateConfigSet failed: %v", err)
		}
	})

	t.Run("TestUploadConfigSet", func(t *testing.T) {
		var err error
		exampleYamlContent := []byte(`---
		contents:
		- name: ccd_env.yaml
		  type: yaml
		  content: |
			# This is YAML content as a plain text string.
			apiserverfqdn: api.crlab-vdu014-cnis.deac.gic.ericsson.se
		- name: cluster_config
		  type: directory
		  contents:
		  - name: input
			type: directory
			contents:
			- name: params.yaml
			  type: yaml
			  content:
				cluster_size: 3
				enable_feature_x: true
		- name: single-server-configuration.yaml
		  type: yaml
		  content: |
			equipment:
			  sdi_name: cniscrlab-rem
			  v_pod:
				vpod_id: edge-vpod-dl110-02
			  relay:
				relay_id: Aachen-dl110-02
				relay_configuration:
				  location_info: Remote DRAN Site Aachen - DL110-002
				  next_hop_address_ipv4: 10.87.87.209
				  prefix_length_ipv4: 28
				  ntp_servers:
					- 164.48.10.70
					- 164.48.10.90
				  relay_address_ipv4: 10.87.87.209
				  user_label: DL110-002
				  postal_address:
					house_number: 20
					postal_code: 97231
					room: Hall E
				  address_ranges:
					- address_range_ipv4_id: DL110-002-ilo
					  address_from: 10.87.87.218
					  address_to: 10.87.87.218
					  binding_id: dl110-002-ilo
					  user_label: DL110-002-ilo
			  compute:
				service_id: dl110-002-ilo
				user_label: dl110-02
				server_profile_name: cnis1.15_dl110_vdu_fw_bios.yaml
			cluster:
			  ccd_config:
				name: dl110-02-ccds
				software_version: 2.31.0
				cluster_template_name: cnis1.15_dl110_vdu_midband
		
		- name: user-secrets.yaml
		  type: yaml
		  content: |
			bmc_username:
				type: file
				files:
				  - secrets/bmc_username
				path: cmc-secret/computes/dl110-002/bmc_username
			bmc_password:
				type: file
				files:
				  - secrets/bmc_password
				path: cmc-secret/computes/dl110-002/bmc_password
		
		`)

		//t.Skip("Skipping")
		//testME := "testme123789"
		//configsetName := "test-config-set"

		service := NewOMCRemoteService(url22, username, password, skipTLSVerify)

		err = service.UploadConfigSetFile(testME,
			configsetName,
			"here is a description",
			exampleYamlContent)

		_ = err

	})

	t.Run("TestDeleteConfigSet", func(t *testing.T) {
		_, err := service.DeleteConfigSet(testME, configsetName)
		if err != nil {
			t.Errorf("DeleteConfigSet failed: %v", err)
		}
	})

	t.Run("TestDeleteME", func(t *testing.T) {

		_, err := service.DeleteME(testME)

		if err != nil {
			t.Errorf("DeleteME failed: %v", err)
		}

	})

}

func ConvertToStringMap(m map[interface{}]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		strKey := fmt.Sprint(k)
		result[strKey] = convertValue(v)
	}
	return result
}

// convertStringMap handles map[string]interface{} that might contain nested interface maps
func convertStringMap(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		result[k] = convertValue(v)
	}
	return result
}

// convertValue handles the conversion of different value types
func convertValue(v interface{}) interface{} {
	switch x := v.(type) {
	case map[interface{}]interface{}:
		// Recursively convert nested maps
		return ConvertToStringMap(x)
	case []interface{}:
		// Handle arrays/slices
		arr := make([]interface{}, len(x))
		for i, item := range x {
			arr[i] = convertValue(item)
		}
		return arr
	case map[string]interface{}:
		// Already a string map, but might contain nested interface maps
		return convertStringMap(x)
	default:
		// Return unchanged for other types
		return v
	}
}

func TestO2imsTemplates(t *testing.T) {

	service := NewOMCRemoteService(url22, username, password, skipTLSVerify)
	testDir := "/tmp/TestO2imsTemplates"

	// Create the directory if not present
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		if err := os.Mkdir(testDir, os.ModePerm); err != nil {
			t.Errorf("Unable to create directory %s: %v", testDir, err)
		}
	}

	// Remove all files in the directory
	files, err := os.ReadDir(testDir)
	if err != nil {
		t.Errorf("Unable to read directory %s: %v", testDir, err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		err = os.Remove(filepath.Join(testDir, file.Name()))
		if err != nil {
			t.Errorf("Unable to remove file %s: %v", file.Name(), err)
		}
	}

	_ = service

	t.Run("TestGenConfigsetFromO2imsTemplate", func(t *testing.T) {
		t.Skip()

		// Check that GenConfigsetFromO2imsTemplate returns an error for invalid parameters
		templateName := "single-node-lpg2"
		_ = templateName
		templateVersion := "cnis-1.15_v1"
		_ = templateVersion
		templateParamContent := []byte(`---
resourceParams:
  managed_element:                                                  # RO
    product: CNIS
    type: SingleServer
    software_version: 1.15
  single-server-configuration:                                      # This section goes into single-server-configuration.yaml
    equipment:
      sdi_name: dc284-sdi3
      v_pod:
        vpod_id: edge-vpod-srv06
        user_label: "My First cluster!"                             # Opt - not sure this is supported
      relay:                                                        # The whole section should be Opt, though it's required, currently
        relay_id: Lulea-dc284srv06
        relay_configuration:
          location_info: Remote DRAN Site Lulea - dc284srv06
          next_hop_address_ipv4: 10.33.17.9
          prefix_length_ipv4: 30
          ntp_servers:
            - 10.152.4.198
            - 10.152.4.215
          relay_address_ipv4: 10.33.17.9
          user_label: Remote DRAN Site Lulea - dc284srv06
          postal_address:
            house_number: 20
            postal_code: 97231
            room: Hall E
          address_ranges:
            - address_range_ipv4_id: dc284srv06-iLO
              address_from: 10.33.17.10
              address_to: 10.33.17.10
              binding_id: ILOCZ22040JZ8
              user_label: dc284srv06-iLO
      compute:
        service_id: ILOCZ22040JZ8
        user_label: dc484srv06-dl110gen10plus
        server_profile_name: cnis1.13_dl110_vdu_fw_bios.yaml        # RO
    cluster:
      ccd_config:
        name: dc284srv06-ccds                                       # Opt?
        software_version: 2.29.0                                    # RO
        cluster_template_name: cnis1.13_dl110_vdu_midband_Vault     # RO
clusterParams:
  ccd_env:
    apiserverfqdn: dc284srv06kubeapi.pcl.seki.gic.ericsson.se       # This parameter goes into ccd_env.yaml
  params:
    control_plane_external_ip: 10.33.17.115
    bootstrap_node_ip: 10.33.17.115                         # Parameter will be removed by CCD 
    apiserver_extra_sans:
      - dc284srv06kubeapi.pcl.seki.gic.ericsson.se
      - api.eccd.local
    cr-registry_hostname: dc284srv06registry.pcl.seki.gic.ericsson.se
    ingress_ip: 10.33.35.152
    alertmanager_hostname: "dc284srv06alertmanager.pcl.seki.gic.ericsson.se"
    victoria_metrics_hostname: "dc284srv06prometheus.pcl.seki.gic.ericsson.se"
    nameservers:
      - "10.221.16.10"
      - "10.221.16.11"
    ntp_servers:
      - "10.152.4.198"
      - "10.152.4.215"
      - "10.152.4.232"
      - "10.152.4.245"
    timezone: "UTC"
    nels_host_ip: 10.155.142.69
    nels_host_name: nelsaas-vnf2-thrift.sero.gic.ericsson.se
    nels_port: 9099
    nels_customer_id: 800141
    nels_swlt_id: STB-CCD-1
    sftp_url:  "ftps://10.33.151.130:22"
    remote_image_server_url: https://10.33.151.130:6182
    value_packs:
      - value_packs/CXP9043234-2.29.0-f55e166ce369dba3eb8c866a2cde27ec.tar.gz
    networks:
      - name: edgeccdintsp
        vlan: 3320
      - name: edgeccdomsp
        vlan: 1101
        gateway_ipv4: 10.33.17.113
        ip_pools:
        - start: 10.33.17.115
          end: 10.33.17.115
          prefix: 29
      - name: edgeranoamsp
        vlan: 1111
        gateway_ipv4: 10.33.17.57
        ip_pools:
        - start: 10.33.17.58
          end: 10.33.17.58
          prefix: 29
    routes:
      config:
        - destination:
          next-hop-address:
            gateway-from-net:
              net-name: edgeccdomsp                        
          next-hop-interface: edge_ccd_om
    ecfe:
      address-pools:
        - name: edge-ecfe-om-pool
          addresses:
            - 10.33.17.136-10.33.17.139
        - name: ran-om-pool
          addresses:
            - 10.33.17.148-10.33.17.151
      static-bfd-peers:
        - peer-address: 10.33.17.113          #peer-address for edge-ecfe-om network
        - peer-address: 10.33.17.57           #peer-address for ran-om network
    mcm_fqdn: dc173centralfileserver4.pcl.seki.gic.ericsson.se
    omc_fqdn: dc173omc1.pcl.seki.gic.ericsson.se
  user_secrets:                                                    # Opt (the whole section)
    bmc_username:
      type: file
      files: 
        - secrets/bmc_username
      path: cmc-secret/computes/dc284srv06/bmc_username
    bmc_password:
      type: file
      files:
        - secrets/bmc_password
      path: cmc-secret/computes/dc284srv06/bmc_password		
`)

		var templateParamContentMap map[string]interface{}
		if err := yaml.Unmarshal([]byte(templateParamContent), &templateParamContentMap); err != nil {
			t.Errorf("Unable to unmarshal yaml string: %v", err)
		}

		// Pretty print yamlMap
		yamlBytes, err := yaml.Marshal(templateParamContentMap)
		if err != nil {
			t.Errorf("Unable to marshal yamlMap: %v", err)
		}
		fmt.Println(string(yamlBytes))

		configSetMap, err := service.GenConfigsetFromO2imsTemplate(templateName, templateVersion, templateParamContentMap)
		if err != nil {
			t.Errorf("GenConfigsetFromO2imsTemplate returne an error ")
		}

		yamlBytes, err = yaml.Marshal(configSetMap)
		if err != nil {
			t.Errorf("Unable to marshal configSetMap: %v", err)
		}
		fmt.Println(string(yamlBytes))

		err = CreateConfigSetFromJSON(testDir, yamlBytes)
		if err != nil {
			t.Errorf("Unable to create configset: %v", err)
		}

		// Print tree structure of directory
		err = filepath.Walk("/tmp/abcdefg", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				fmt.Println("[DIR] ", path)
			} else {
				fmt.Println("[FILE] ", path)
			}
			return nil
		})

	})

}

/*
// TestFetchAuthTokenWithMockServer is a test case that uses a mock server
func TestFetchAuthTokenWithMockServer(t *testing.T) {
	// Create a new instance of OMCRemoteServiceImpl
	service := NewOMCRemoteService(username, password, skipTLSVerify)

	// Mock the HTTP request and response
	req, err := http.NewRequest("POST", "https://example.com/token",
		bytes.NewBufferString("client_id=test&client_secret=test&grant_type=password&username=test&password=test"))
	if err != nil {
		t.Fatalf("failed to create token request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Create a mock HTTP client that returns a mock response
	client := &http.Client{
		Transport: &mockRoundTripper{
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       http.NoBody,
			},
		},
	}
	service.HTTPClient = client

	// Call the FetchAuthToken function
	err = service.FetchAuthToken()
	if err != nil {
		t.Errorf("FetchAuthToken failed: %v", err)
	}

}

// mockRoundTripper is a mock implementation of http.RoundTripper
type mockRoundTripper struct {
	response *http.Response
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.response, nil
}
*/
