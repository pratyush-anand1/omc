package omc_rest

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

const (
	// URL of the OMC REST API.
	//OMCURL = "https://gui-omc.cniscrlab-cl10.deac.gic.ericsson.se/"
	// Username for the OMC REST API.
	//Username = "admin"
	// Password for the OMC REST API.
	//Password = "admin"

	//OMCBaseURL              = "https://gui-omc.cniscrlab-cl10.deac.gic.ericsson.se"
	TokenURL                = "%s/auth/realms/omc/protocol/openid-connect/token"
	CreateMEEndpoint        = "%s/cmc/api/managed-elements/v1"
	GetMEStatusEndpoint     = "%s/cmc/api/managed-elements/v1/%s"
	CreateConfigSetEndpoint = "%s/cmc/api/config-mgmt/v1/%s/config-sets"
	UploadConfigSetEndpoint = "%s/cmc/api/config-mgmt/v1/%s/config-sets/%s"
	DeleteConfigSetEndpoint = "%s/cmc/api/config-mgmt/v1/%s/config-sets/%s"
	RunOperationLCMEndpoint = "%scmc/api/lcm/v1/run_operations"
	ClientID                = "OMC_API_Server"
	ClientSecret            = "6976dcae-6166-485b-91d1-c529ce402f96"
	AuthExpiryThreshold     = 280 * time.Second // keep just less than 5 minutes
)

const (
	state = "state"
)

var (
	InfoLogger  = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarnLogger  = log.New(os.Stdout, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
)

// OMCConfig is the configuration for the OMC REST API.
// It can be initialized only once.
type OMCConfig struct {
	//once     sync.Once
	Url      string
	Username string
	Password string
}

// OMC is the global instance of OMCConfig.
//var OMC = &OMCConfig{
//	Url:      OMCURL,
//	Username: Username,
//	Password: Password,
//}

// Configure initializes the OMC config.
// // It can be called only once.
// func (c *OMCConfig) Configure(url, username, password string) {
// 	c.once.Do(func() {
// 		c.Url = url
// 		c.Username = username
// 		c.Password = password
// 	})
// }

type OMCRemoteServiceImpl struct {
	AuthToken  string
	AuthExpiry time.Time
	HTTPClient *http.Client
	Username   string
	Password   string
	OmcBaseUrl string
}

func NewOMCRemoteService(omcBaseURL, username, password string, InsecureSkipVerify bool) *OMCRemoteServiceImpl {

	return &OMCRemoteServiceImpl{
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: InsecureSkipVerify, // Disables TLS certificate verification
				},
			},
		},
		Username:   username,
		Password:   password,
		OmcBaseUrl: omcBaseURL,
	}
}

func (s *OMCRemoteServiceImpl) FetchAuthToken() error {
	payload := fmt.Sprintf("client_id=%s&username=%s&password=%s&grant_type=password&client_secret=%s",
		ClientID, s.Username, s.Password, ClientSecret)

	tokenURL := fmt.Sprintf(TokenURL, s.OmcBaseUrl)

	req, err := http.NewRequest("POST",
		tokenURL, bytes.NewBufferString(payload))

	if err != nil {
		return fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch auth token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token request failed, status code: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}

	if token, ok := result["access_token"].(string); ok {
		s.AuthToken = token
		s.AuthExpiry = time.Now().Add(AuthExpiryThreshold)

		log.Println("Token refreshed successfully.")
		return nil
	}
	return errors.New("access_token not found in response")
}

func (s *OMCRemoteServiceImpl) GetME(meName string) (map[string]interface{}, error) {

	// Refresh the token if it has expired
	if s.AuthExpiry.Before(time.Now()) || s.AuthToken == "" {
		if err := s.FetchAuthToken(); err != nil {
			return nil, err
		}
	}

	endpoint := fmt.Sprintf(GetMEStatusEndpoint, s.OmcBaseUrl, meName)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create GET request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.AuthToken)
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute managed element get status request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("managed element status request failed response: (%w)", err)
		}
		message, ok := result["message"].(string)
		if !ok {
			message = "unknown"
		}
		return nil, fmt.Errorf("managed element status request failed: (%s), httpstatus: %d", message, resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return result, nil
}

func (s *OMCRemoteServiceImpl) DeleteME(meName string) (map[string]string, error) {
	// Refresh the token if it has expired
	if s.AuthExpiry.Before(time.Now()) || s.AuthToken == "" {
		if err := s.FetchAuthToken(); err != nil {
			return nil, err
		}
	}

	//GET and Delete Endpoints are same
	endpoint := fmt.Sprintf(GetMEStatusEndpoint, s.OmcBaseUrl, meName)
	req, err := http.NewRequest("DELETE", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to DELETE: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.AuthToken)
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute DELETE request: %w", err)
	}
	defer resp.Body.Close()

	//FIXME check if me is not preset then aslo it should be error
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch ME status, status code: %d", resp.StatusCode)
	}

	return nil, nil
}

func (s *OMCRemoteServiceImpl) CreateME(payload map[string]string) error {

	// Refresh the token if it has expired
	if s.AuthExpiry.Before(time.Now()) || s.AuthToken == "" {
		if err := s.FetchAuthToken(); err != nil {
			return err
		}
	}
	var err error
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request payload: %w", err)
	}
	if err != nil {
		return fmt.Errorf("failed to encode request payload: %w", err)
	}
	reqBodyBytes := []byte(reqBody)
	endpoint := fmt.Sprintf(CreateMEEndpoint, s.OmcBaseUrl)
	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(reqBodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.AuthToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request to OMC: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var err error
		var response map[string]interface{}
		if resp.ContentLength != 0 {
			if err = json.NewDecoder(resp.Body).Decode(&response); err == nil {
				if message, ok := response["message"].(string); ok {
					return fmt.Errorf("failed to create managed element (%s), httpstatus: %d", message, resp.StatusCode)
				}
			}
		}
		return fmt.Errorf("failed to create managed element, httpstatus: %d", resp.StatusCode)
	}
	return nil
}

func (s *OMCRemoteServiceImpl) CreateConfigSet(meName string,
	payload map[string]string) (map[string]string, error) {
	var err error
	if s.AuthExpiry.Before(time.Now()) || s.AuthToken == "" {
		if err = s.FetchAuthToken(); err != nil {
			return nil, err
		}
	}

	if meName == "" {
		return nil, errors.New("meName is mandatory field")
	}
	endpoint := fmt.Sprintf(CreateConfigSetEndpoint, s.OmcBaseUrl, meName)
	configSetName, ok := payload["configSetName"]
	if !ok {
		return nil, errors.New("configSetName is mandatory field")
	}
	meSWVersion, ok := payload["swVersion"]
	if !ok {
		return nil, errors.New("swVersion is mandatory field")
	}
	description, ok := payload["description"]
	if !ok {
		return nil, errors.New("description is mandatory field")
	}

	params := url.Values{}
	params.Set("configSetName", configSetName)
	params.Set("swVersion", meSWVersion)
	params.Set("description", description)
	endpoint += "?" + params.Encode()

	req, err := http.NewRequest("POST", endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.AuthToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to OMC: %w", err)
	}
	defer resp.Body.Close()

	message := "unknown"
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err == nil {
		message, _ = response["message"].(string)
	}
	// ConfigSet already exists, we are not concerned about it
	// its okay for us

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusConflict {
		return nil, fmt.Errorf("failed to create (%s) for %s, (%s) httpstatus: %d", configSetName, meName, message, resp.StatusCode)
	}
	//Fixme me we canrespond to the message later
	return nil, nil
}

func (s *OMCRemoteServiceImpl) DeleteConfigSet(meName, configSetName string) (map[string]string, error) {
	var err error
	if s.AuthExpiry.Before(time.Now()) || s.AuthToken == "" {
		if err = s.FetchAuthToken(); err != nil {
			return nil, err
		}
	}
	if meName == "" || configSetName == "" {
		var message string
		if meName == "" {
			message = "meName is mandatory field"
		}
		if configSetName == "" {
			message = "configSetName is mandatory field"
		}
		return nil, errors.New(message)
	}
	endpoint := fmt.Sprintf(DeleteConfigSetEndpoint, s.OmcBaseUrl, meName, configSetName)
	req, err := http.NewRequest("DELETE", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to DELETE: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.AuthToken)
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute DELETE request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to execute DELETE reques: %d", resp.StatusCode)
	}
	return nil, nil
}

type Content struct {
	Name     string    `yaml:"name"`               // Name of the file/directory
	Type     string    `yaml:"type"`               // "directory", "yaml", "text"
	Content  any       `yaml:"content,omitempty"`  // File content (text, parsed YAML, etc)
	Contents []Content `yaml:"contents,omitempty"` // Child nodes for directories
}

// RootConfigSet contains the list of contents in the configset.
type RootContent struct {
	Contents []Content `yaml:"contents"`
}

func CreateConfigSetFromJSON(path string, yamlContent []byte) error {
	var contents RootContent
	fmt.Printf("yamlContent: %s\n", string(yamlContent))

	if err := yaml.Unmarshal(yamlContent, &contents); err != nil {
		return err
	}

	for _, subContent := range contents.Contents {
		fmt.Printf("subContent: %v\n", subContent.Name)
		if err := createContent(path, subContent); err != nil {
			fmt.Printf("subContent: %v  %v\n", subContent.Name, err)
			return err
		}
	}
	return nil
}

func createContent(path string, c Content) error {

	currentPath := filepath.Join(path, c.Name)
	fmt.Printf("currentPath: %s\n", currentPath)
	if c.Type == "directory" {
		// Create the directory.

		err := os.MkdirAll(currentPath, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %v", currentPath, err)
		}
		// Recursively create any children.
		for _, c := range c.Contents {
			if err := createContent(currentPath, c); err != nil {
				return err
			}
		}
	} else if c.Type == "yaml" {
		var data []byte

		// If content is provided as a string, use it as is.
		// Otherwise, marshal the content into YAML.
		switch content := c.Content.(type) {
		case string:
			data = []byte(content)
		default:
			var err error
			data, err = yaml.Marshal(content)
			if err != nil {
				return fmt.Errorf("failed to marshal YAML for file %s: %v", currentPath, err)
			}
		}
		// Write the file.
		err := os.WriteFile(currentPath, data, 0644)
		if err != nil {
			return fmt.Errorf("failed to write file %s: %v", currentPath, err)
		}
	}
	return nil
}

func CreateTarGz(sourceDir string, outputPath string) error {
	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer outputFile.Close()

	// Create gzip writer
	gzipWriter := gzip.NewWriter(outputFile)
	defer gzipWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// Walk through the source directory
	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %s: %v", path, err)
		}

		// Get the relative path for the tar header
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return fmt.Errorf("error getting relative path: %v", err)
		}

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("error creating tar header: %v", err)
		}

		// Update header name with relative path
		header.Name = relPath

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("error writing tar header: %v", err)
		}

		// If it's a regular file, write its contents
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("error opening file %s: %v", path, err)
			}
			defer file.Close()

			if _, err := io.Copy(tarWriter, file); err != nil {
				return fmt.Errorf("error copying file contents: %v", err)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *OMCRemoteServiceImpl) UploadConfigSetFile(meName,
	configSetName,
	commitMessage string,
	yamlContent []byte) error {

	if meName == "" || configSetName == "" {
		var message string
		if meName == "" {
			message = "meName is mandatory field"
		}
		if configSetName == "" {
			message = "configSetName is mandatory field"
		}
		return errors.New(message)
	}
	//Create the directory in /temp/mename/configsetname/version
	tempDir := filepath.Join("/tmp", meName+"-"+configSetName)
	//trunc := strings.Replace(configSetName, "-", "", -1)
	config_tgz := filepath.Join("/tmp", meName+"-"+configSetName+".tar.gz")

	if _, err := os.Stat(tempDir); err == nil {
		// Remove the existing directory
		if err := os.RemoveAll(tempDir); err != nil {
			return fmt.Errorf("failed to remove the existing directory %s: %v", tempDir, err)
		}
	}
	if _, err := os.Stat(config_tgz); err == nil {
		if err := os.Remove(config_tgz); err != nil {
			return fmt.Errorf("failed to remove the existing file %s: %v", config_tgz, err)
		}
	}

	// Create the directory
	err := os.MkdirAll(tempDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create the directory %s: %v", tempDir, err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			log.Printf("failed to delete the directory %s: %v", tempDir, err)
		}
		if err := os.Remove(config_tgz); err != nil {
			log.Printf("failed to remove the existing file %s: %v", config_tgz, err)
		}
	}()

	err = CreateConfigSetFromJSON(tempDir, yamlContent)
	if err != nil {
		return fmt.Errorf("failed to create configset.yaml file: %v", err)
	}
	err = CreateTarGz(tempDir, filepath.Join("", config_tgz))
	if err != nil {
		return fmt.Errorf("failed to create tar.gz file: %v", err)
	}

	//HTTP part begins
	if s.AuthExpiry.Before(time.Now()) || s.AuthToken == "" {
		if err := s.FetchAuthToken(); err != nil {
			return err
		}
	}

	endpoint := fmt.Sprintf(UploadConfigSetEndpoint, s.OmcBaseUrl, meName, configSetName)

	req, err := http.NewRequest("PUT", endpoint, bytes.NewBuffer(yamlContent))
	if err != nil {
		return fmt.Errorf("failed to create the HTTP request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.AuthToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")

	params := req.URL.Query()
	params.Set("commitMessage", commitMessage)
	req.URL.RawQuery = params.Encode()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	defer writer.Close()
	req.Header.Set("Content-Type", writer.FormDataContentType())

	partHeader := make(textproto.MIMEHeader)
	partHeader.Set("Content-Disposition", `form-data; name="file"; filename="`+filepath.Base(config_tgz)+`"`)
	partHeader.Set("Content-Type", "application/gzip")
	part, err := writer.CreatePart(partHeader)

	if err != nil {
		return fmt.Errorf("failed to create the multipart writer: %v", err)
	}
	file, err := os.Open(config_tgz)
	if err != nil {
		return fmt.Errorf("failed to open the file: %v", err)
	}
	defer file.Close()
	_, err = io.Copy(part, file)
	if err != nil {
		return fmt.Errorf("failed to copy the file: %v", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close the multipart writer: %v", err)
	}

	req.Body = io.NopCloser(body)
	req.ContentLength = int64(body.Len())
	req.Header.Set("Cache-Control", "no-cache")

	fmt.Printf("Multipart Body:\n")
	resp, err := s.HTTPClient.Do(req)

	if err != nil {
		fmt.Printf("failed to make request to OMC: %v\n", err)
		return fmt.Errorf("failed to make request to OMC: %w", err)
	}
	fmt.Printf("response status code: %d\n", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Response:\n")
		dump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			fmt.Printf("error dumping response: %v\n", err)
		} else {
			fmt.Printf("Dump:\n%s\n", dump)
		}

		var message string
		var response map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&response); err == nil {
			message, _ = response["message"].(string)
		}
		if message == "" {
			message = resp.Status
		}
		fmt.Printf("failed to upload the configset: %s\n", message)
		return fmt.Errorf("failed to upload the configset: %s", message)
	}
	return nil
}

func (s *OMCRemoteServiceImpl) GetO2imsTemplateList() (map[string]interface{}, error) {
	// implementation
	response := make(map[string]interface{})
	response["dummy"] = "dummy"
	return response, nil
}

func (s *OMCRemoteServiceImpl) CheckO2imsTemplateSupport(templateName, templateVersion string) (map[string]string, error) {
	response := make(map[string]string)
	response["dummy"] = "dummy"
	return response, nil
}

func (s *OMCRemoteServiceImpl) VerifyO2imsTemplateParams(templateName, templateVersion string, params map[string]interface{}) (map[string]string, error) {
	response := make(map[string]string)
	response["dummy"] = "dummy"
	return response, nil
}

// GetMEDetaislFromO2imsTemplate retrieves the details of a managed element.
//
// templateName: the name of the template.
// templateVersion: the version of the template.
// params: the parameters of the template.
// Returns the details of the managed element and an error if the retrieval fails.
func (s *OMCRemoteServiceImpl) GetMEDetaislFromO2imsTemplate(
	templateName,
	templateVersion string,
	params map[string]interface{}) (map[string]interface{}, error) {

	if templateName != "single-node-lpg2" || templateVersion != "cnis-1.15_v1" {
		return nil, fmt.Errorf("template not supported: %s:%s", templateName, templateVersion)
	}

	// check params
	if params == nil {
		return nil, fmt.Errorf("params is required")
	}

	resourceParams, ok := params["resourceParams"].(map[string]interface{})
	if !ok || resourceParams == nil {
		return nil, fmt.Errorf("resourceParams is required")
	}

	managedElement, ok := resourceParams["managed_element"].(map[string]interface{})
	if !ok || managedElement == nil {
		return nil, fmt.Errorf("managed_element is required")
	}

	mandatoryFields := map[string]interface{}{
		"product":          "",
		"type":             "",
		"software_version": "",
	}

	for key := range mandatoryFields {
		if _, ok := managedElement[key]; !ok {
			return nil, fmt.Errorf("managed_element is missing required key: %s", key)
		}
		if value, ok := managedElement[key].(string); !ok || value == "" {
			return nil, fmt.Errorf("managed_element is missing non empty string value for key: %s", key)
		}
	}
	response := managedElement
	return response, nil
}

// For now we do
// MandatoryFields represents the required fields and their values
//type MandatoryFields struct {
//	TemplateName    string
//	TemplateVersion string
//	TemplateParams  interface{}
//}

func (s *OMCRemoteServiceImpl) GenConfigsetFromO2imsTemplate(
	templateName string,
	templateVersion string,
	tparams map[string]interface{}) (
	map[string]interface{},
	error) {
	if templateName != "single-node-lpg2" || templateVersion != "cnis-1.15_v1" {
		return nil, fmt.Errorf("template not supported: %s:%s", templateName, templateVersion)
	}

	resourceParams, ok := tparams["resourceParams"].(map[string]interface{}) //map[interface {}]interface {}

	if !ok || resourceParams == nil {
		ErrorLogger.Printf("resourceParams is required: %v", tparams)
		return nil, fmt.Errorf("resourceParams is required")
	}

	clusterParams, ok := tparams["clusterParams"].(map[string]interface{})
	if !ok || clusterParams == nil {
		ErrorLogger.Printf("clusterParams is required: %v", tparams)
		return nil, fmt.Errorf("clusterParams is required")
	}

	singleServerConfiguration, ok := resourceParams["single-server-configuration"].(map[string]interface{})

	fmt.Printf("resourceParams type: %T\n", resourceParams)
	if !ok || singleServerConfiguration == nil {
		ErrorLogger.Printf("singleServerConfiguration is required: %v", tparams)
		return nil, fmt.Errorf("single-server-configuration is required")
	}

	_ = singleServerConfiguration //file

	ccd_env, ok := clusterParams["ccd_env"].(map[string]interface{})
	if !ok || ccd_env == nil {
		ErrorLogger.Printf("ccd_env is required: %v", tparams)
		return nil, fmt.Errorf("ccd_env is required")
	}
	_ = ccd_env //file

	p, ok := clusterParams["params"].(map[string]interface{})
	if !ok || p == nil {
		ErrorLogger.Printf("params is required: %v", tparams)
		return nil, fmt.Errorf("params is required")
	}

	params := map[string]interface{}{
		"params": p,
	}
	_ = params //file

	user_secrets, ok := clusterParams["user_secrets"].(map[string]interface{})

	if !ok || user_secrets == nil {
		return nil, fmt.Errorf("user_secrets is required")
	}

	_ = user_secrets //file

	responseYaml := map[string]interface{}{
		"contents": []interface{}{
			map[interface{}]interface{}{
				"name":    "ccd_env.yaml",
				"type":    "yaml",
				"content": ccd_env,
			},
			map[interface{}]interface{}{
				"name": "cluster_config",
				"type": "directory",
				"contents": []interface{}{
					map[string]interface{}{
						"name": "input",
						"type": "directory",
						"contents": []interface{}{
							map[string]interface{}{
								"name":    "params.yaml",
								"type":    "yaml",
								"content": params,
							},
						},
					},
				},
			},
			map[interface{}]interface{}{
				"name":    "single-server-configuration.yaml",
				"type":    "yaml",
				"content": singleServerConfiguration,
			},
			map[interface{}]interface{}{
				"name":    "user-secrets.yaml",
				"type":    "yaml",
				"content": user_secrets,
			},
		},
	}
	return responseYaml, nil
}

func (s *OMCRemoteServiceImpl) GetAllME() ([]map[string]interface{}, error) {
	response := []map[string]interface{}{}
	return response, nil
}

func (s *OMCRemoteServiceImpl) GetActiveWFListOfMe(meName string) (string, error) {
	response := "dummy"
	return response, nil
	// implementation
}

func (s *OMCRemoteServiceImpl) UpdateME(meName, state, softwareVersion, activeOperations string) error {
	dummy := meName + state + softwareVersion + activeOperations
	if dummy != "dummy" {
		return errors.New("invalid parameters")
	}
	return nil
	// implementation
}

func (s *OMCRemoteServiceImpl) ListConfigSets(meName string) (map[string]interface{}, error) {
	response := make(map[string]interface{})
	response["dummy"] = "dummy"
	return response, nil
	// implementation
	// implementation
}

func (s *OMCRemoteServiceImpl) GetLCMOperList() (map[string]interface{}, error) {
	response := make(map[string]interface{})
	response["dummy"] = "dummy"
	return response, nil
}

func (s *OMCRemoteServiceImpl) RunLCMOper(payload map[string]interface{}) (map[string]string, error) {
	response := make(map[string]string)

	// additionalParamsDefaultValues := map[string]interface{}{
	// 	"unattended":                         true,
	// 	"ignoreEquipmentHealthcheckWarnings": true,
	// 	"promptForError":                     true,
	// 	"deleteVpod":                         false,
	// 	"deleteRelay":                        false,
	// }

	// check for mandatory fields
	if payload["operation"] == "" || payload["managedElement"] == "" || payload["configSet"] == "" {
		return nil, fmt.Errorf("mandatory fields operation, managedElement, and configSet are required")
	}

	// Add optionalLCMParams to payload if not present
	//
	// optionalLCMParams is a string and contains key-value pairs.
	// optionalLCMParams can have the following key-value pairs:
	//
	// Key-Value Pairs:
	// "unattended":                         true
	// "ignoreEquipmentHealthcheckWarnings": true
	// "promptForError":                     true
	// "excludeClusterHealthcheckAlarms":    "9895953, 9895955, 9895994, 9895956, 9895941 "
	// "deleteVpod":                         false
	// "deleteRelay":                        false
	//
	// If optionalLCMParams is not present, add it as empty string.
	if _, ok := payload["optionalLCMParams"]; !ok {
		payload["optionalLCMParams"] = ""
	}

	if _, ok := payload["additionalParams"]; !ok {
		payload["additionalParams"] = ""
	}

	// if _, ok := payload["additionalParams"]; !ok {
	// 	payload["additionalParams"] = additionalParamsDefaultValues
	// } else {
	// 	// add the missing keys and their values from the default values.
	// 	addnParam := payload["additionalParams"].(map[string]interface{})
	// 	for k, v := range additionalParamsDefaultValues {
	// 		if _, ok := addnParam[k]; !ok {
	// 			addnParam[k] = v
	// 		}
	// 	}
	// 	payload["additionalParams"] = addnParam
	// }

	if s.AuthExpiry.Before(time.Now()) || s.AuthToken == "" {
		if err := s.FetchAuthToken(); err != nil {
			return nil, err
		}
	}

	// Dump the payload for debugging.

	//c := payload["configSet"].(string)
	//c = c + "junk"
	//payload["configSet"] = c
	payloadStr, _ := json.MarshalIndent(payload, "", "    ")
	log.Printf("payload: %s", payloadStr)

	endpoint := fmt.Sprintf(RunOperationLCMEndpoint, s.OmcBaseUrl)
	var err error
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to encode request payload: %w", err)
	}
	reqBodyBytes := []byte(reqBody)
	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(reqBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to run lcm request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.AuthToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make run lcm request to OMC: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		//Add pending and other  statu stoo
		var err error
		var response map[string]interface{}
		if resp.ContentLength != 0 {
			if err = json.NewDecoder(resp.Body).Decode(&response); err == nil {
				if message, ok := response["message"].(string); ok {
					return nil, fmt.Errorf("failed to create run lcm workflow (%s), httpstatus: %d", message, resp.StatusCode)
				}
			}
		}
		return nil, fmt.Errorf("failed to create run lcm workflow, httpstatus: %d", resp.StatusCode)
	}
	return response, nil
}

func (s *OMCRemoteServiceImpl) GetWorkflow(workflowId string) (map[string]string, error) {
	response := map[string]string{"dummy": "dummy"}
	return response, nil
}

func (s *OMCRemoteServiceImpl) UpdateWorkflow(workflowId string, workflow map[string]string) error {
	dummy := workflowId + strings.Join([]string{state}, " ")
	if dummy != "dummy" {
		return errors.New("invalid parameters")
	}

	return nil
}

func (s *OMCRemoteServiceImpl) UpdateWorkflowStatus(workflowId, state string) error {
	dummy := workflowId + strings.Join([]string{state}, " ")
	if dummy != "dummy" {
		return errors.New("invalid parameters")
	}
	return nil
}
