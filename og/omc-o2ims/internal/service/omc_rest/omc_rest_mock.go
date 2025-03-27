package omc_rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const defaultInstallTime = time.Minute * 2
const defaultUninstallTime = time.Minute * 2

type MockOMCRemoteService interface {
	OMCRemoteService

	SetAPIError(string, error, int, int)
	SetMeStatus(string, map[string]interface{})
}

type mockMe struct {
	self      map[string]interface{}
	configset map[string]interface{}
}
type mockOMCRemoteService struct {
	ManagedElements []mockMe
	workflows       []map[string]string
	APIError        error
	APIErrorCount   int
	APIErrorSkip    int
	APIName         string
}

func NewMockOMCRemoteService() MockOMCRemoteService {
	return &mockOMCRemoteService{}
}

func (s *mockOMCRemoteService) GetAllME() (
	[]map[string]interface{},
	error) {
	MEs := []map[string]interface{}{}
	for _, me := range s.ManagedElements {
		MEs = append(MEs, me.self)
	}
	return MEs, nil
}

func (s *mockOMCRemoteService) DeleteME(name string) (
	map[string]string,
	error) {
	found := false
	// Delete the managed element if name matches
	for _, me := range s.ManagedElements {
		if me.self["name"] == name {
			fmt.Printf("Deleting managed element: %s\n", name)
			s.ManagedElements = append(s.ManagedElements[:1],
				s.ManagedElements[2:]...)
			found = true
			break
		}
	}
	if found {
		response := map[string]string{
			"message": "Managed element deleted successfully",
		}
		return response, nil
	} else {
		response := map[string]string{
			"message": "Managed element not found",
		}
		return response, apierrors.NewNotFound(schema.GroupResource{}, name)
	}
}

// SetAPIError is used to set the API error for a particular API name
func (s *mockOMCRemoteService) SetAPIError(apiName string,
	err error,
	errCount int,
	skip int) {
	s.APIName = apiName
	s.APIError = err
	s.APIErrorCount = errCount
	s.APIErrorSkip = skip
}
func (s *mockOMCRemoteService) GetME(name string) (
	map[string]interface{},
	error) {

	if s.APIName == "GetME" {
		if s.APIErrorCount > 0 && s.APIErrorSkip == 0 {
			s.APIErrorCount--
			return nil, s.APIError
		}
		if s.APIErrorSkip > 0 {
			s.APIErrorSkip--
		}
	}

	found := false
	for i := range s.ManagedElements {
		me := &s.ManagedElements[i]
		if me.self["name"] == name {
			found = true
			_ = found
			resp := s.ManagedElements[i].self
			return resp, nil
		}
	}
	return nil, errors.New("managed element not found")
}

// SetMeStatus sets the status of a managed element.
func (s *mockOMCRemoteService) SetMeStatus(
	meName string,
	state map[string]interface{}) {
	for i, me := range s.ManagedElements {
		if me.self["name"] == meName {
			if state != nil {
				s.ManagedElements[i].self["state"] = state
			}
			//s.ManagedElements[i].getMeError = err
		}
	}
}

func (s *mockOMCRemoteService) CreateME(
	payload map[string]string) error {
	var err error
	if err != nil {
		return err
	}

	if s.APIName == "CreateME" {
		if s.APIErrorCount > 0 && s.APIErrorSkip == 0 {
			s.APIErrorCount--
			return s.APIError
		}
		if s.APIErrorSkip > 0 {
			s.APIErrorSkip--
		}
	}

	for _, me := range s.ManagedElements {
		if me.self["name"] == payload["name"] {
			return apierrors.NewAlreadyExists(schema.GroupResource{
				Group:    "omc.oran.org.omc.v1alpha1",
				Resource: "ManagedElement",
			}, payload["name"])
		}
	}

	// Create the managed element and add to the list
	me := mockMe{}
	createTime := time.Now().Format(time.RFC3339)

	me.self = map[string]interface{}{
		"activeOperations": []string{},
		"annotations":      map[string]interface{}{},
		"description":      payload["description"],
		"flavor":           payload["flavor"],
		"name":             payload["name"],
		"owner":            map[string]interface{}{},
		"product":          payload["product"],
		"softwareVersion":  "",
		"state": map[string]interface{}{
			"administrative": "unlocked",
			"operational":    "defined",
		},
		"operation":      "",
		"creationTime":   createTime,
		"lastUpdateTime": createTime,
	}

	me.configset = map[string]interface{}{}
	s.ManagedElements = append(s.ManagedElements, me)
	/* These are possible states of Managed Element */
	//"defined", "error", "import"  "install", "maintenance","ready","reinstall","uninstall","upgrade","validation"

	return nil
}

func (s *mockOMCRemoteService) GetActiveWFListOfMe(
	meName string) (string, error) {
	wfStr := ""

	for _, me := range s.ManagedElements {
		if me.self["name"] == meName {
			wf := me.self["activeOperations"]
			if wf != nil {
				if wfs, ok := wf.([]interface{}); ok {
					if len(wfs) > 0 {
						wfStr = fmt.Sprintf("%v", wfs[0])
					}
				}
			}
			if wf == nil {
				wfStr = ""
			}

			return wfStr, nil
		}
	}
	return "", errors.New("managed element not found")
}

func (s *mockOMCRemoteService) UpdateME(name,
	op,
	swVer,
	activeOperations string) error {

	op = strings.ToLower(op)

	if op != "deploy" && op != "undeploy" && op != "update" && op != "" {
		return errors.New("invalid operation")
	}

	for i, _ := range s.ManagedElements {
		me := &s.ManagedElements[i]
		if me.self["name"] == name {
			me.self["operation"] = op
			// ["install", "uninstall", "reinstall", "software upgrade", "configuration change", "import"],
			if op != "" {
				if op == "deploy" {
					me.self["operation"] = "install"
				} else if op == "Undeploy" {
					me.self["operation"] = "uninstall"
				} else if op == "Update" {
					me.self["operation"] = "upgrading"
				}
			}

			if swVer != "" {
				me.self["softwareVersion"] = swVer
			}
			if activeOperations != "" {
				me.self["activeOperations"] = activeOperations
			}
			me.self["lastUpdateTime"] = time.Now().Format(time.RFC3339)
			return nil
		}
	}
	return errors.New("managed element not found")
}

func (s *mockOMCRemoteService) ListConfigSets(managedElementName string) (map[string]interface{}, error) {
	configSets := []map[string]interface{}{
		{
			"createdAt":          1683631812746,
			"description":        "config1 for ccd cluster",
			"managedElementName": "dc112",
			"name":               "user-branch1",
			"updatedAt":          1683632115046,
			"version":            "1.0.0",
		},
		{
			"createdAt":          1683631812746,
			"description":        "config2 for ccd cluster",
			"managedElementName": "dc113",
			"name":               "user-branch2",
			"updatedAt":          1683632115046,
			"version":            "2.0.0",
		},
	}

	response := map[string]interface{}{
		"configSets": configSets,
		"offset":     0,
		"size":       2,
		"total":      2,
	}

	return response, nil
}

func (s *mockOMCRemoteService) CreateConfigSet(
	managedElementName string,
	payload map[string]string,
) (map[string]string, error) {

	response := map[string]string{
		"message": "Config set created successfully",
	}

	return response, nil
}

func (s *mockOMCRemoteService) DeleteConfigSet(
	managedElementName string,
	configSetName string,
) (map[string]string, error) {

	response := map[string]string{
		"code":    "0",
		"message": "Config set deleted successfully",
	}

	return response, nil
}

// Mock function for uploading a specified config tar.gz file to a specified config branch
func (s *mockOMCRemoteService) UploadConfigSetFile(
	managedElementName string,
	configSetName string,
	commitMessage string,
	yaml []byte) error {
	return nil
}

// Mock function for getting a list of O2IMA template names and versions
func (s *mockOMCRemoteService) GetO2imsTemplateList() (map[string]interface{}, error) {
	templateList := []map[string]interface{}{
		{
			"name":    "template1",
			"version": "1.0.0",
		},
		{
			"name":    "template2",
			"version": "2.0.0",
		},
	}

	response := map[string]interface{}{
		"templateList": templateList,
	}

	return response, nil
}

func (s *mockOMCRemoteService) CheckO2imsTemplateSupport(
	templateName string,
	templateVersion string,
) (map[string]string, error) {
	if templateName == "template1" && templateVersion == "1.0.0" {
		response := map[string]string{
			"message": "Template is supported",
		}
		return response, nil
	}

	response := map[string]string{
		"message": "Template is not supported",
	}

	return response, nil
}

// VerifyO2imsTemplateParams verifies if the template params are valid.
func (s *mockOMCRemoteService) VerifyO2imsTemplateParams(
	templateName string,
	templateVersion string,
	params map[string]interface{},
) (map[string]string, error) {
	if templateName == "template1" && templateVersion == "1.0.0" {
		response := map[string]string{
			"message": "Template params are valid",
		}
		return response, nil
	}
	response := map[string]string{
		"message": "Template params are not valid",
	}
	return response, nil
}

// Mock function for generating a configset for the template
func (s *mockOMCRemoteService) GenConfigsetFromO2imsTemplate(templateName string,
	templateVersion string,
	params map[string]interface{}) (map[string]interface{},
	error) {
	response := map[string]interface{}{
		"message": "Template params are valid",
		"yaml": `
	base:
	  ccdadmconfig:
	    # Content of ccdadmconfig.yaml
	  userSecrets:
	    # Content of user-secrets.yaml
	  systemSecrets:
	    # Content of system-secrets.yaml
	  rootDeviceHints:
	    # Content of rootdevicehints.yaml
	`,
	}
	return response, nil
}

func (s *mockOMCRemoteService) GetLCMOperList() (map[string]interface{}, error) {

	response := map[string]interface{}{
		"apiVersion": "plugins/v1",
		"kind":       "OperationsList",
		"metadata": map[string]interface{}{
			"productType": map[string]interface{}{
				"flavor":          "bm-sdi2",
				"product":         "CCD",
				"softwareVersion": "2.24",
			},
		},
		"operations": []map[string]interface{}{
			{
				"additionalParams": []map[string]interface{}{
					{
						"description":  "Run complete upgrade without additional user input",
						"isMandatory":  true,
						"paramName":    "unattended",
						"valueDefault": false,
						"valueType":    "boolean",
					},
				},
				"categories": []string{
					"configuration_change",
					"software_upgrade",
				},
				"description": "ccd upgrade",
				"lockType":    "exclusive",
				"name":        "upgrade",
			},
			{
				"categories": []string{
					"import",
				},
				"lockType": "exclusive",
				"name":     "import",
			},
			{
				"categories": []string{
					"install",
				},
				"lockType": "exclusive",
				"name":     "install",
			},
		},
	}

	return response, nil
}

// Mock method for triggering and executing an LCM operation
func (s *mockOMCRemoteService) RunLCMOper(
	payload map[string]interface{}) (map[string]string, error) {

	payloadStr, err := json.MarshalIndent(payload, "", "    ")
	if err != nil {
		return nil, err
	}
	_ = payloadStr
	//fmt.Printf("mockOMCRemoteService RunLCMOper payload:\n%s\n", payloadStr)

	id, _ := uuid.NewRandom()
	resp := map[string]string{
		"workflowId": id.String(),
	}

	op, ok := payload["operationName"].(string)
	op = strings.ToLower(op)
	if !ok || (op != "undeploy" && op != "deploy" && op != "update") {
		return nil, errors.New("invalid operation")
	}
	createTime := time.Now().Format(time.RFC3339)
	wf := map[string]string{
		"id":             id.String(),
		"creationTime":   createTime,
		"lastUpdateTime": createTime,
		"state":          "running",
		"operation":      payload["operationName"].(string),
	}
	meName, ok := payload["managedElements"].(string)
	if !ok {
		return nil, errors.New("invalid managedElements in payload")
	}
	s.UpdateME(meName, op, "", id.String())

	s.workflows = append(s.workflows, wf)

	return resp, nil
}

// GetWorkflow retrieves a workflow by its ID.
func (s *mockOMCRemoteService) GetWorkflow(
	workflowId string) (map[string]string, error) {
	// Find the workflow in the list
	for i, wf := range s.workflows {
		if wf["id"] == workflowId {
			updateTime, err := time.Parse(time.RFC3339, wf["lastUpdateTime"])
			if err != nil {
				return nil, err
			}

			//Undeploy" && op != "Deploy" && op != "Update")
			op, ok := wf["operation"]
			op = strings.ToLower(op)
			if !ok {
				return nil, errors.New("invalid operation in payload")
			}
			if op == "deploy" {
				if time.Since(updateTime) > defaultInstallTime {
					s.workflows[i]["state"] = "succeeded"
				}
			} else if op == "undeploy" {
				if time.Since(updateTime) > defaultUninstallTime {
					s.workflows[i]["state"] = "succeeded"
				}
			} else if op == "update" {
				if time.Since(updateTime) > defaultInstallTime {
					s.workflows[i]["state"] = "succeeded"
				}
			}
			return wf, nil
		}
	}
	// Return an error if the workflow is not found
	return nil, fmt.Errorf("workflow with ID %s not found", workflowId)
}

func (s *mockOMCRemoteService) UpdateWorkflowStatus(workflowId,
	state string) error {

	validStates := map[string]bool{
		"Aborting":   true,
		"Aborted":    true,
		"New":        true,
		"Paused":     true,
		"Running":    true,
		"Waiting":    true,
		"Recovering": true,
		"Recovered":  true,
		"Failed":     true,
		"Succeeded":  true,
	}

	if !validStates[state] {
		return fmt.Errorf("invalid state %s", state)
	}
	// Find the workflow in the list
	updateTime := time.Now().Format(time.RFC3339)

	for i, wf := range s.workflows {
		if wf["id"] == workflowId {
			s.workflows[i]["state"] = state
			s.workflows[i]["lastUpdateTime"] = updateTime
			return nil
		}
	}
	// Return an error if the workflow is not found
	return fmt.Errorf(
		"workflow with ID %s not found", workflowId)
}

// UpdateWorkflow updates the specified workflow.
func (s *mockOMCRemoteService) UpdateWorkflow(
	workflowId string,
	workflow map[string]string,
) error {
	return nil
}

func (s *mockOMCRemoteService) GetMEDetaislFromO2imsTemplate(
	templateName,
	templateVersion string,
	params map[string]interface{}) (map[string]interface{}, error) {
	managedElement := map[string]interface{}{
		"product":          "CNIS",
		"type":             "SingleServer",
		"software_version": "1.15",
	}

	return managedElement, nil
}
