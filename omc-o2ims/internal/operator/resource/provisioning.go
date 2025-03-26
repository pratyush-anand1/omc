package resource

import (
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"log"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/enrayga/omc-o2ims/internal/service/omc_rest"
	"gopkg.in/yaml.v2"
)

var (
	InfoLogger  = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarnLogger  = log.New(os.Stdout, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
)

func NewProvisioningRequest(resourceName,
	APIVersion,
	kind string,
	metaData map[string]interface{},
	spec map[string]interface{},
	status map[string]interface{}) *ProvisioningRequest {
	return &ProvisioningRequest{
		ResourceName:  resourceName,
		APIVersion:    APIVersion,
		Kind:          kind,
		MetaData:      metaData,
		Spec:          spec,
		Status:        status,
		remoteService: nil,
	}
}

type ProvisioningRequest struct {
	ResourceName  string                 `json:"resourceName"`
	APIVersion    string                 `json:"apiVersion"`
	Kind          string                 `json:"kind"`
	MetaData      map[string]interface{} `json:"metadata"`
	Spec          map[string]interface{} `json:"spec"`
	Status        map[string]interface{} `json:"status"`
	remoteService omc_rest.OMCRemoteService
	mu            sync.RWMutex
}

type Status struct {
	Extensions           map[string]interface{} `json:"extensions"`
	ProvisionedResources ProvisionedResources   `json:"provisionedResources"`
	ProvisioningStatus   ProvisioningStatus     `json:"provisioningStatus"`
}

type ProvisioningStatus struct {
	ProvisioningMessage    string `json:"provisioningMessage"`
	ProvisioningState      string `json:"provisioningState"`
	ProvisioningUpdateTime string `json:"provisioningUpdateTime"`
}

type ProvisionedResources struct {
	OCloudInfrastructureResourceIds []string `json:"oCloudInfrastructureResourceIds"`
	OCloudNodeClusterId             string   `json:"oCloudNodeClusterId"`
}

func (r *ProvisioningRequest) GetID() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.ResourceName
}

// update the initial reconciliation info
// Do not rcall if already populated or somethign
func (r *ProvisioningRequest) InitReconciliationInfo(resourceName string) {

	startTime := getCurrentTime()
	lastUpdateTime := startTime
	transitionTime := startTime

	reconciliationInfo := map[string]interface{}{
		ReconciliationState:   Init,
		SubState:              "",
		ApiFailure:            "",
		ApiRetryCount:         "0",
		ApiBackOffTime:        startTime,
		MarkedForDeletion:     false,
		TemplateName:          "",
		TemplateVersion:       "",
		TemplateParamsCRC:     "",
		TemplateParamsApplied: false,
		WorkflowId:            "",
		ConfigSetName:         "",
		ConfigSetCrc:          "",
		MeName:                resourceName,    //FIXME: we get from metadata and not spec
		MeDescription:         resourceName,    //FIXME: we get from spec
		MeProductType:         "CNIS",          //FIXME: Hardcoded we get from Params anyways  will get updated
		MeFlavorType:          "single-server", //FIXME: Hardcoded we get from Params anyways  will get updated
		MeSwVer:               "1.15",          //FIXME: Hardcoded we get from Params anyways  will get updated
		OmcOperation:          "",
		StartTime:             startTime,
		EndTime:               "",
		TransitionTime:        transitionTime,
		ReconciliationTimeout: "",
		LastUpdateTime:        lastUpdateTime,
	}
	if _, ok := r.Status["extensions"]; !ok {
		r.Status["extensions"] = make(map[string]interface{})
	}

	ext := r.Status["extensions"].(map[string]interface{})

	//check if reconciliationInfo is not present or not a map
	if _, ok := ext["reconciliationInfo"]; !ok || reflect.TypeOf(ext["reconciliationInfo"]).Kind() != reflect.Map {
		ext["reconciliationInfo"] = reconciliationInfo
	} else {
		//iterate thoru each key and add the missing key with value
		for k, v := range reconciliationInfo {
			if _, ok := ext["reconciliationInfo"].(map[string]interface{})[k]; !ok {
				ext["reconciliationInfo"].(map[string]interface{})[k] = v
			}
		}
	}

	//if _, ok := ext["reconciliationInfo"]; !ok {
	//ext["reconciliationInfo"] = reconciliationInfo
}

func (r *ProvisioningRequest) SetInitFields(
	name string,
	fields map[string]interface{}) error {

	specFound := false
	statusFound := false
	reconcileInfoFound := false

	if name == "" {
		return errors.New("Name cannot be empty")
	}
	r.ResourceName = name
	for field, value := range fields {
		switch field {
		case "apiVersion":
			r.APIVersion = value.(string)
		case "kind":
			r.Kind = value.(string)
		case "metadata":
			r.MetaData = value.(map[string]interface{})
			if metadata, ok := r.MetaData["name"]; ok {
				r.ResourceName = metadata.(string)
			}
			//We do not need this to be mantained int the struct
			// so strip it for now
			delete(r.MetaData, "managedFields")
		case "spec":
			r.Spec = value.(map[string]interface{})
			if _, ok := r.Spec["templateName"]; !ok {
				return errors.New("missing templateName in spec")
			}

			if _, ok := r.Spec["templateVersion"]; !ok {
				return errors.New("missing templateVersion in spec")
			}

			if _, ok := r.Spec["templateParameters"]; !ok {
				return errors.New("missing templateParameters in spec")
			}
			specFound = true
			if specDescription, ok := r.Spec["description"]; ok {
				r.Spec["description"] = specDescription.(string)
			} else {
				r.Spec["description"] = "manged element with unknown description"
			}
		case "status":
			r.Status = value.(map[string]interface{})
			statusFound = true
			if _, ok := r.Status["extensions"]; ok {
				if extensions, ok := r.Status["extensions"].(map[string]interface{}); ok {
					if _, ok := extensions["reconciliationInfo"]; ok {
						reconcileInfoFound = true
						// we have reconciliation info, so we are good
						InfoLogger.Printf("reconcileInfo Found in the status (restarted case): %v\n", reconcileInfoFound)
					}
				}
			}
		}
	}

	if !specFound {
		return errors.New("missing Spec in ProvisioningRequest")
	}

	if !statusFound {
		r.Status = map[string]interface{}{}
	}

	if r.Status["provisioningStatus"] == nil {
		r.Status["provisioningStatus"] = make(map[string]interface{})
	}
	provisioningStatus := r.Status["provisioningStatus"].(map[string]interface{})

	if provisioningStatus["provisioningMessage"] == nil {
		provisioningStatus["provisioningMessage"] = ""
	}
	if provisioningStatus["provisioningState"] == nil {
		provisioningStatus["provisioningState"] = ""
	}
	if provisioningStatus["provisioningUpdateTime"] == nil {
		provisioningStatus["provisioningUpdateTime"] = ""
	}
	r.InitReconciliationInfo(name)
	return nil
}

func (r *ProvisioningRequest) Compare(
	name string,
	fields map[string]interface{},
	apply bool) (bool, error) {

	r.mu.Lock()
	defer r.mu.Unlock()

	changed := false

	//FIXME we ignore anything else

	spec := fields["spec"].(map[string]interface{})
	res_spec := r.Spec

	oldTemplateName := res_spec["templateName"].(string)
	newTemplateName := spec["templateName"].(string)

	if oldTemplateName != newTemplateName {
		changed = true
		InfoLogger.Printf("provisioning request templateName changed from %s to %s\n", oldTemplateName, newTemplateName)
		if apply {
			r.Spec["templateName"] = newTemplateName
		}
	}

	oldTemplateVersion := res_spec["templateVersion"].(string)
	newTemplateVersion := spec["templateVersion"].(string)

	if oldTemplateVersion != newTemplateVersion {
		changed = true
		InfoLogger.Printf("provisioning request templateVersion changed from %s to %s\n", oldTemplateVersion, newTemplateVersion)
		if apply {
			r.Spec["templateVersion"] = newTemplateVersion
		}
	}

	res_spec_params := res_spec["templateParameters"].(map[string]interface{})
	spec_params := spec["templateParameters"].(map[string]interface{})

	// safety check
	if !reflect.DeepEqual(spec_params, res_spec_params) {

		changed = true
		InfoLogger.Println("provisioning request templateParameters changed")

		if apply {
			r.Spec["templateParameters"] = spec_params
		}
	}

	return changed, nil
}

func (r *ProvisioningRequest) GetStatus() (map[string]interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	k := r.Status
	return k, nil
}

// GetNew returns an empty provisioning request.
func (r *ProvisioningRequest) GetNew() interface{} {
	return &ProvisioningRequest{}
}

// UpdateProvisioningStatus updates the provisioning status of the request.
// It only updates the selected fields.
func (r *ProvisioningRequest) UpdateProvisioningStatus(
	newProvisioningStatus map[string]interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.Status["provisioningStatus"]; !ok {
		r.Status["provisioningStatus"] = make(map[string]interface{})
	}

	//"provisioningMessage": "Cluster provisioning completed",
	//"provisioningState": "fulfilled",
	//"provisioningUpdateTime": "2025-01-05T17:19:11Z"

	for k, v := range newProvisioningStatus {
		r.Status["provisioningStatus"].(map[string]interface{})[k] = v
	}
	return nil
}

// updateReconcileInfo updates the reconcile info of the request.
// It only updates the selected fields.
func (r *ProvisioningRequest) updateReconcileInfo(
	newReconcileInfo map[string]interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.Status["extensions"]; !ok {
		r.Status["extensions"] = make(map[string]interface{})
	}

	ext := r.Status["extensions"].(map[string]interface{})

	if _, ok := ext["reconciliationInfo"]; !ok {
		ext["reconciliationInfo"] = make(map[string]interface{})
	}

	//fmt.Printf("gyan newReconcileInfo:  %v %v\n", OmcOperation, newReconcileInfo)

	for k, v := range newReconcileInfo {
		switch k {
		case ReconciliationState,
			SubState,
			ApiFailure,
			ApiRetryCount,
			ApiBackOffTime,
			TemplateName,
			MarkedForDeletion,
			TemplateVersion,
			TemplateParamsCRC,
			TemplateParamsApplied,
			WorkflowId,
			ConfigSetName,
			ConfigSetCrc,
			MeName,
			MeDescription,
			MeProductType,
			MeFlavorType,
			MeSwVer,
			OmcOperation,
			StartTime,
			EndTime,
			TransitionTime,
			ReconciliationTimeout,
			LastUpdateTime:

			r.Status["extensions"].(map[string]interface{})["reconciliationInfo"].(map[string]interface{})[k] = v
		}
	}
	return nil
}

func getParamCRC(params map[string]interface{}) string {
	table := crc32.MakeTable(crc32.IEEE)
	crc := crc32.Checksum([]byte(fmt.Sprintf("%v", params)), table)
	return strconv.FormatUint(uint64(crc), 10)
}

func getCurrentTime() string {
	return time.Now().Format(time.RFC3339)
}

func (r *ProvisioningRequest) Dump() {
	req, _ := json.MarshalIndent(r, "", "    ")
	fmt.Println(string(req))
}

func (r *ProvisioningRequest) specChanged(specTmplName,
	specTmplVer,
	reconTemplName,
	reconTemplVersion string,
	specTemplParam map[string]interface{},
	reconTemplCrc string) (bool, error) {

	if specTmplName != reconTemplName {
		//fmt.Printf("templateName has changed\n")
		return true, nil
	}

	if specTmplVer != reconTemplVersion {
		//fmt.Printf("templateVersion has changed\n")
		return true, nil
	}

	if specTemplParam == nil {
		//fmt.Printf("templateParams is missing\n")
		return true, nil
	}
	specTemplateCRC := getParamCRC(specTemplParam)

	if reconTemplCrc != specTemplateCRC {
		//fmt.Printf("templateParamsCRC has changed\n")
		return true, nil
	}

	//fmt.Printf("No changes in provisioning request\n")
	return false, nil
}

// GetReconciliationInfo returns the current reconciliation info
func (r *ProvisioningRequest) GetReconciliationInfo() (
	map[string]interface{},
	error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if _, ok := r.Status["extensions"]; !ok {
		r.Status["extensions"] = make(map[string]interface{})
	}

	ext := r.Status["extensions"].(map[string]interface{})

	if _, ok := ext["reconciliationInfo"]; !ok {
		ext["reconciliationInfo"] = make(map[string]interface{})
	}

	//FIXME error check
	reconciliationInfo, _ := ext["reconciliationInfo"].(map[string]interface{})

	return reconciliationInfo, nil
}

func (s *ProvisioningRequest) checkRemoteService() {
	//TBD add logic to determine which service to use
	if s.remoteService != nil {
		return
	}

	backendType := os.Getenv("OMC_BACKEND")
	if backendType == "omc_rest_v1" {

		backendURL, ok := os.LookupEnv("OMC_BACKEND_URL")
		if !ok {
			ErrorLogger.Println("Failed to read OMC_BACKEND_URL environment variable")
		}

		backendUsername, ok := os.LookupEnv("OMC_BACKEND_USERNAME")
		if !ok {
			ErrorLogger.Println("Failed to read OMC_BACKEND_USERNAME environment variable")
		}

		backendPassword, ok := os.LookupEnv("OMC_BACKEND_PASSWORD")
		if !ok {
			ErrorLogger.Println("Failed to read OMC_BACKEND_PASSWORD environment variable")
		}

		if backendURL == "" || backendUsername == "" || backendPassword == "" {
			ErrorLogger.Println("Failed to read OMC_BACKEND environment variables")
		}

		s.remoteService = omc_rest.NewOMCRemoteService(backendURL, backendUsername, backendPassword, true)
		InfoLogger.Println("Using OMC REST V1 omc_rest_v1")

	} else {
		s.remoteService = omc_rest.NewMockOMCRemoteService()
		InfoLogger.Println("Using OMC REST Simulator")
	}
}

// Validates the template and the template params
func (s *ProvisioningRequest) validateTemplates(tempateName,
	templateVersion string) error {
	// Check if the template is supported
	if _, err := s.remoteService.CheckO2imsTemplateSupport(tempateName,
		templateVersion); err != nil {
		return err
	}
	return nil
}

// validateTemplateParams validates the template parameters.
//
// specTmplName: the name of the template.
// specTmplVer: the version of the template.
// specTemplParam: the parameters of the template.
// Returns an error if the validation fails.
func (s *ProvisioningRequest) validateTemplateParams(specTmplName,
	specTmplVer string,
	specTemplParam map[string]interface{}) error {

	_ = specTemplParam
	_ = specTmplName
	_ = specTmplVer

	// Implementation goes here
	return nil
}

// GetTemplateNameVersionAndParams returns the template name, version, and template params.
// Returns an error if the parameters are not found.
func (s *ProvisioningRequest) GetTemplateNameVersionAndParams(root map[string]interface{}) (
	string,
	string,
	map[string]interface{},
	error) {

	templateName, ok := root["templateName"].(string)
	if !ok {
		return "", "", nil, fmt.Errorf("templateName not found")
	}

	templateVersion, ok := root["templateVersion"].(string)
	if !ok {
		return "", "", nil, fmt.Errorf("templateVersion not found")
	}

	templateParameters, ok := root["templateParameters"].(map[string]interface{})
	if !ok {
		return "", "", nil, fmt.Errorf("templateParameters not found")
	}

	return templateName, templateVersion, templateParameters, nil
}

// Generates a configset from the template
func (s *ProvisioningRequest) genConfigset(templateName,
	templateVersion string,
	templateParameters map[string]interface{}) (
	map[string]interface{},
	error) {

	yaml, err := s.remoteService.GenConfigsetFromO2imsTemplate(
		templateName,
		templateVersion,
		templateParameters,
	)

	return yaml, err
}

// Create the configset and upload the configset
func (s *ProvisioningRequest) createConfigSetAndUpload(
	meName,
	meSWVersion,
	configSetName string,
	configSet map[string]interface{}, commitMessage string) error {

	//var templateParamContentMap map[string]interface{}
	//if err := yaml.Unmarshal([]byte(templateParamContent), &templateParamContentMap); err != nil {
	//	t.Errorf("Unable to unmarshal yaml string: %v", err)
	//}

	// Pretty print yamlMap
	//yamlBytes, err := yaml.Marshal(templateParamContentMap)
	//if err != nil {
	//		t.Errorf("Unable to marshal yamlMap: %v", err)
	//	}
	//	fmt.Println(string(yamlBytes))

	//	configSetMap, err := service.GenConfigsetFromO2imsTemplate(templateName, templateVersion, templateParamContentMap)
	//	if err != nil {
	//		t.Errorf("GenConfigsetFromO2imsTemplate returne an error ")
	//	}

	payload := make(map[string]string)
	payload["configSetName"] = configSetName
	payload["swVersion"] = meSWVersion
	payload["description"] = commitMessage
	_ = configSet
	// Create the configset
	_, err := s.remoteService.CreateConfigSet(meName, payload)
	if err != nil {
		return err
	}

	yamlBytes, err := yaml.Marshal(configSet)
	if err != nil {
		return fmt.Errorf("failed to marshal configSet: %v", err)
	}

	err = s.remoteService.UploadConfigSetFile(meName,
		configSetName,
		commitMessage,
		yamlBytes)
	if err != nil {
		return err
	}
	return nil
}

func (s *ProvisioningRequest) createMEWithParams(mename,
	description,
	productType,
	flavor string) error {

	//Depends on the real impmentation
	//Here used for mock testing
	payload := map[string]string{
		"name":        mename,
		"description": description,
		"product":     productType,
		"flavor":      flavor,
	}
	// var err error

	// payload["name"] = mename
	// payload["description"] = description
	// payload["product"] = productType
	// payload["flavor"] = flavor

	// check if the managed element exist
	me, err := s.remoteService.GetME(mename)
	if err == nil {
		//FIXME we may remove description check for now
		if me["description"] != payload["description"] ||
			me["product"] != payload["product"] ||
			me["flavor"] != payload["flavor"] {
			// return err if the managed element fields do not match
			return fmt.Errorf("me exist! existing/expected %s/%s %s/%s %s/%s",
				me["description"], payload["description"],
				me["product"], payload["product"],
				me["flavor"], payload["flavor"])
		}
		return nil
	}
	//FIXME: More checks on error for now just assuming managed element does not exist
	err = s.remoteService.CreateME(payload)
	if err != nil {
		ErrorLogger.Printf("CreateME failed for %s:%v\n", mename, err)
	} else {
		InfoLogger.Printf("CreateME successful for %s\n", mename)
	}
	return err
}

// LCMRunOperation runs an LCM operation
func (s *ProvisioningRequest) LCMRunOperation(me,
	configset,
	operation,
	optionalLCMParams string,
	additionalParams map[string]interface{}) (string, error) {

	payload := make(map[string]interface{})

	payload["configSet"] = configset
	payload["managedElements"] = me
	payload["operationName"] = operation
	payload["optionalLCMParams"] = ""
	payload["additionalParams"] = additionalParams

	//fmt.Printf("ProvisioningRequest LCMRunOperation payload %v\n", payload)

	wf, err := s.remoteService.RunLCMOper(payload)
	if err != nil {
		return "", err
	}
	wfID := fmt.Sprintf("%v", wf["workflowId"])
	return wfID, nil
}

// GetRemoteService returns the OMCRemoteService.
func (s *ProvisioningRequest) GetRemoteService() omc_rest.OMCRemoteService {
	return s.remoteService
}

// SetRemoteService sets the OMCRemoteService.
func (s *ProvisioningRequest) SetRemoteService(service omc_rest.OMCRemoteService) {
	s.remoteService = service
}

func (s *ProvisioningRequest) IsWorkflowFinishedAndMEReady(
	workflowID,
	meName,
	meState4 string) (bool, error) {
	wf, err := s.remoteService.GetWorkflow(workflowID)

	if err != nil {
		return false, err
	}

	// Check if the workflow is done

	wfState := wf["state"]
	wfState = strings.ToLower(wfState)

	if wfState != "succeeded" {
		return false, nil
	}

	// Check the managed element status
	me, err := s.remoteService.GetME(meName)
	if err != nil {
		return false, err
	}

	// Check if the managed element is in ready state
	// Check if the managed element is in a stable state
	//FIXME Remove New
	meState, ok := me["state"].(string)
	if !ok {
		return false, fmt.Errorf("invalid state in managed element")
	}
	meState = strings.ToLower(meState)

	switch meState {
	case "new", "defined", "ready", "error":
		fmt.Printf("managed element %v reached stable state me[state]: %v\n", me["name"], me["state"])
		return true, nil
		// Proceed to check if the workflow is done
	default:
		//fmt.Printf("checking for managed element %v to non transient me[state]: %v\n", me["name"], me["state"])
		return false, nil
	}
}

// getMEStatus gets the managed element status
func (s *ProvisioningRequest) getMEStatus(meName string) (string,
	string,
	error) {
	me, err := s.remoteService.GetME(meName)
	if err != nil || me == nil {
		return "", "", err
	}

	adminState, ok := me["state"].(map[string]interface{})["administrative"].(string)
	if !ok {
		adminState = "unknown"
	}

	operState, ok := me["state"].(map[string]interface{})["operational"].(string)
	if !ok {
		operState = "unknown"
	}
	return adminState, operState, nil
}

// SetDeleteFlag sets the flag to delete the resource.
func (s *ProvisioningRequest) SetDeleteFlag() error {

	reconInfo := make(map[string]interface{})
	reconInfo["markedForDeletion"] = true
	_ = s.updateReconcileInfo(reconInfo)
	return nil
}

// GetDeleteFlag gets the flag to delete the resource.
func (s *ProvisioningRequest) GetDeleteFlag() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if ext, ok := s.Status["extensions"].(map[string]interface{}); ok {
		if recon, ok := ext["reconciliationInfo"]; ok {
			if ext, ok := recon.(map[string]interface{}); ok {
				if markedForDeletion, ok := ext["markedForDeletion"]; ok {
					return markedForDeletion.(bool)
				}
			}
		}
	}
	return false
}

// checkReconcileState is an internal function which checks for all mandatory fields
// before reconciliation starts.
func (s *ProvisioningRequest) checkManadatoryFields() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	spec := s.Spec
	status := s.Status

	// check if mandatory fields are present
	if spec["templateName"] == nil || spec["templateVersion"] == nil {
		return errors.New("missing Mandatory fields: templateName, templateVersion")
	}
	// check if templateParams is present and of type map interface
	if spec["templateParameters"] == nil || reflect.TypeOf(spec["templateParameters"]).Kind() != reflect.Map {
		return errors.New("missing mandatory fields or wrong type: templateParameters")
	}

	// check if extensions is present and of type map interface
	if status["extensions"] == nil || reflect.TypeOf(status["extensions"]).Kind() != reflect.Map {
		return errors.New("missing mandatory fields or wrong type: extensions")
	}
	// check if reconciliationInfo is present and of type map interface
	if status["extensions"].(map[string]interface{})["reconciliationInfo"] == nil ||
		reflect.TypeOf(status["extensions"].(map[string]interface{})["reconciliationInfo"]).Kind() != reflect.Map {
		return errors.New("missing Mandatory fields or wrong type: reconciliationInfo")
	}

	reconciliationInfo := status["extensions"].(map[string]interface{})["reconciliationInfo"].(map[string]interface{})

	if reconciliationInfo[ReconciliationState] == nil {
		return errors.New("missing Mandatory fields: ReconciliationState")
	}
	if reconciliationInfo[MarkedForDeletion] == nil {
		return errors.New("missing Mandatory fields: markedForDeletion")
	}

	if reflect.TypeOf(reconciliationInfo[MarkedForDeletion]).Kind() != reflect.Bool {
		return errors.New("markedForDeletion should be of type bool")
	}

	if reconciliationInfo[TemplateName] == nil {
		return errors.New("missing Mandatory fields: TemplateName")
	}
	if reconciliationInfo[TemplateVersion] == nil {
		return errors.New("missing Mandatory fields: TemplateVersion")
	}
	if reconciliationInfo[TemplateParamsCRC] == nil {
		return errors.New("missing Mandatory fields: TemplateParamsCRC")
	}
	if reconciliationInfo[TemplateParamsApplied] == nil {
		return errors.New("missing Mandatory fields: TemplateParamsApplied")
	}
	if reconciliationInfo[WorkflowId] == nil {
		return errors.New("missing Mandatory fields: WorkflowId")
	}
	if reconciliationInfo[ConfigSetName] == nil {
		return errors.New("missing Mandatory fields: ConfigSetName")
	}
	if reconciliationInfo[ConfigSetCrc] == nil {
		return errors.New("missing Mandatory fields: ConfigSetCrc")
	}
	if reconciliationInfo[MeName] == nil {
		return errors.New("missing Mandatory fields: MeName")
	}
	if reconciliationInfo[MeDescription] == nil {
		return errors.New("missing Mandatory fields: MeDescription")
	}
	if reconciliationInfo[MeProductType] == nil {
		return errors.New("missing Mandatory fields: MeProductType")
	}
	if reconciliationInfo[MeFlavorType] == nil {
		return errors.New("missing Mandatory fields: MeFlavorType")
	}
	if reconciliationInfo[MeSwVer] == nil {
		return errors.New("missing Mandatory fields: MeSwVer")
	}
	if reconciliationInfo[OmcOperation] == nil {
		return errors.New("missing Mandatory fields: OmcOperation")
	}
	if reconciliationInfo[StartTime] == nil {
		return errors.New("missing Mandatory fields: StartTime")
	}
	if reconciliationInfo[EndTime] == nil {
		return errors.New("missing Mandatory fields: EndTime")
	}
	if reconciliationInfo[LastUpdateTime] == nil {
		return errors.New("missing Mandatory fields: LastUpdateTime")
	}
	if reconciliationInfo[TransitionTime] == nil {
		return errors.New("missing Mandatory fields: TransitionTime")
	}
	if reconciliationInfo[ReconciliationTimeout] == nil {
		return errors.New("missing Mandatory fields: ReconciliationTimeout")
	}

	if status["provisioningStatus"] == nil {
		return errors.New("missing Mandatory fields: provisioningStatus")
	}

	provisioningStatus, ok := status["provisioningStatus"].(map[string]interface{})
	if !ok {
		return errors.New("provisioningStatus is not of type map[string]interface{}")
	}

	if provisioningStatus["provisioningMessage"] == nil {
		return errors.New("missing Mandatory fields: provisioningStatus.provisioningMessage")
	}

	if provisioningStatus["provisioningState"] == nil {
		return errors.New("missing Mandatory fields: provisioningStatus.provisioningState")
	}

	if provisioningStatus["provisioningUpdateTime"] == nil {
		return errors.New("missing Mandatory fields: provisioningStatus.ProvisioningUpdateTime")
	}
	return nil
}

// UpdateReconcileInfo is used to update the reconcile info of the request.
// It is updated locally during the reconcile for each provisioning element and is
// subsequently used to update the real reconcile info.

type reconcileData struct {
	genConfigset     bool
	configsetName    string
	configsetVersion string
	configSetCRC     string

	specChanged       bool
	specTmplName      string
	specTmplVer       string
	reconTemplName    string
	reconTemplVersion string
	specTemplParam    map[string]interface{}
	reconTemplCrc     string
	reconTemplApplied bool
	meName            string
	meDescription     string
	meProductType     string
	meFlavorType      string
	meSwVer           string
	workflowId        string
	reconcileState    string
	markedForDeletion bool
	omcOperation      string
	subState          string
	apiFailure        string
	apiRetryCount     string
	apiBackoffTime    time.Time
}

func initReconcileData() reconcileData {
	return reconcileData{
		genConfigset:      false,
		configsetName:     "",
		configsetVersion:  "",
		configSetCRC:      "",
		specChanged:       false,
		specTmplName:      "",
		specTmplVer:       "",
		reconTemplName:    "",
		reconTemplVersion: "",
		specTemplParam:    nil,
		reconTemplCrc:     "",
		reconTemplApplied: false,
		meName:            "",
		meDescription:     "",
		meProductType:     "",
		meFlavorType:      "",
		meSwVer:           "",
		workflowId:        "",
		reconcileState:    "",
		markedForDeletion: false,
		omcOperation:      "",
		subState:          "",
		apiFailure:        "",
		apiRetryCount:     "0",
		apiBackoffTime:    time.Now(),
	}
}

func (r *reconcileData) populateFromReconciliationInfo(rInfo map[string]interface{}) {
	r.reconcileState = rInfo[ReconciliationState].(string)
	r.reconTemplName = rInfo[TemplateName].(string)
	r.reconTemplVersion = rInfo[TemplateVersion].(string)
	r.reconTemplCrc = rInfo[TemplateParamsCRC].(string)
	r.reconTemplApplied = rInfo[TemplateParamsApplied].(bool)
	r.meName = rInfo[MeName].(string)
	r.meDescription = rInfo[MeDescription].(string)
	r.meProductType = rInfo[MeProductType].(string)
	r.meFlavorType = rInfo[MeFlavorType].(string)
	r.meSwVer = rInfo[MeSwVer].(string)

	r.configsetName = rInfo[ConfigSetName].(string)
	r.configsetVersion = "IDK"
	r.configSetCRC = rInfo[ConfigSetCrc].(string)
	r.workflowId = rInfo[WorkflowId].(string)
	r.omcOperation = rInfo[OmcOperation].(string)
	r.subState = rInfo[SubState].(string)
	r.apiFailure = rInfo[ApiFailure].(string)
	r.apiRetryCount = rInfo[ApiRetryCount].(string)
}

const (
	//managed element start with New then goes to Defined
	InstallUpdateCreateConfig     string = "CreateConfig"
	InstallUpdatePushConfig       string = "PushConfig"
	InstallUpdateOperationStart   string = "OperationStart"
	InstallUpdateOperationMonitor string = "OperationMonitor"
	InstallUpdateError            string = "Error"
)

func MakeRFC1123Compliant(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)

	// Replace invalid characters with hyphens
	reg := regexp.MustCompile(`[^a-z0-9-]`)
	name = reg.ReplaceAllString(name, "-")

	// Replace multiple consecutive hyphens with a single hyphen
	reg = regexp.MustCompile(`-+`)
	name = reg.ReplaceAllString(name, "-")

	// Trim hyphens from start and end
	name = strings.Trim(name, "-")

	// Ensure the length is not more than 63 characters
	if len(name) > 63 {
		name = name[:63]
		// If truncation resulted in ending with a hyphen, remove it
		name = strings.TrimSuffix(name, "-")
	}
	return name
}

func retrieveConfigSet(configSetName, meName string) (map[string]interface{}, error) {
	fileName := fmt.Sprintf("/tmp/%s-%s.yaml", meName, configSetName)

	// check if the file exist
	_, err := os.Stat(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("configset file %s not found", fileName)
		}
		return nil, err
	}

	yamlBytes, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	yamlContent := make(map[string]interface{})
	err = yaml.Unmarshal(yamlBytes, &yamlContent)
	if err != nil {
		return nil, err
	}

	return yamlContent, nil
}

// SaveConfigSet saves the yaml content in a file in /tmp with the name configsetName-meName.yaml
// if the file is already present, it gets deleted before
func persisitConfigSet(configSetName, meName string, yamlContent map[string]interface{}) error {
	yamlBytes, err := yaml.Marshal(yamlContent)
	if err != nil {
		return err
	}

	fileName := fmt.Sprintf("/tmp/%s-%s.yaml", meName, configSetName)

	// delete the file if it is present
	if _, err := os.Stat(fileName); !os.IsNotExist(err) {
		err = os.Remove(fileName)
		if err != nil {
			return err
		}
	}

	// create the file
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	// write the content
	_, err = file.Write(yamlBytes)
	if err != nil {
		return err
	}

	return nil
}

func (r *reconcileData) handleProvisioning(s *ProvisioningRequest, isMarkedForDeletion bool, isInstall bool) (string, string) {

	/*
	   State Machine Logic:
	   +───────────────────────────────────+────────────────────────────────────────────+
	   │ managed element State & Conditions             │ Action & Transitions                       │
	   ├───────────────────────────────────┼────────────────────────────────────────────┤
	   │ IF MarkedForDeletion              │ Transition to:                             │
	   │ AND State ∈ {Defined,             │ → Removing                                 │
	   │             Ready, Error}         │ Message: "Removal in Progress"             │
	   ├───────────────────────────────────┼────────────────────────────────────────────┤
	   │ IF State ∈ {Installing,           │ Transition to:                             │
	   │            Importing, Upgrading,  │ → PendingForPrevious                       │
	   │            Uninstalling,          │ Reason: "Wait for current operation        │
	   │            Reinstalling}          │  completion before handling spec change"   │
	   │ AND SpecChanged
	   | or MarkedForDeletion              │                                            │
	   ├───────────────────────────────────┼────────────────────────────────────────────┤
	   │ IF State ∈ {Ready, Error, Defined}│ Begin Install/Update Sequence:             │
	   │ AND NOT MarkedForDeletion         │ 1. InstallUpdateCreateConfig (CreateConfig)│
	   │                                   │ 2. InstallUpdatePushConfig (PushConfig)    │
	   │                                   │ 3. InstallUpdateOperationStart             │
	   │                                   │ 4. InstallUpdateOperationMonitor           │
	   │                                   │ Terminal States:                           │
	   │                                   │ - InstallUpdateError (on failure)          │
	   │                                   │ - Ready (on success)                       │
	   +───────────────────────────────────+────────────────────────────────────────────+
	   - → = transitions to
	   - Terminal states end state progression
	*/
	var retrycount, _ = strconv.Atoi(r.apiRetryCount)
	var nextState, message string
	defer func() {
		r.apiRetryCount = strconv.Itoa(retrycount)
	}()
	if isMarkedForDeletion {
		nextState = Removing
		return nextState, "delete is requested hence hence moving forward to delete"
	}
	nextState = Provisioning
	switch r.subState {
	case ProvisioningCreateConfig, "":
		r.subState = ProvisioningCreateConfig
		var genConfigSet = true
		name := r.reconTemplName + "-" + r.reconTemplVersion
		name = MakeRFC1123Compliant(name)
		if name == r.configsetName && r.configSetCRC == r.reconTemplCrc {
			genConfigSet = false
		}
		if genConfigSet {
			templateParameters := s.Spec["templateParameters"].(map[string]interface{})
			templateName, templateVersion, ok := s.Spec["templateName"].(string), s.Spec["templateVersion"].(string), true
			if !ok {
				nextState = Error
				message = fmt.Sprintf("failed to extract templateName and templateVersion from spec:%v", s.Spec)
				return nextState, message
			}
			configset, err := s.genConfigset(templateName, templateVersion, templateParameters)
			if err != nil {
				retrycount++
				message = fmt.Sprintf("Failed to generate configset %s", err.Error())
				r.apiFailure = err.Error()
				return nextState, message
			} else {
				retrycount = 0
				r.apiFailure = ""
				r.configsetName = name
				crc := getParamCRC(r.specTemplParam)
				r.configSetCRC = crc
				r.genConfigset = true
				r.subState = ProvisioningPushConfig
				_ = persisitConfigSet(r.configsetName, r.meName, configset)
				message = fmt.Sprintf("successfully generated configset %s", r.configsetName)
			}
		}
		fallthrough
	case ProvisioningPushConfig:
		r.subState = ProvisioningPushConfig
		configset, err := retrieveConfigSet(r.configsetName, r.meName)
		if err != nil {
			message = fmt.Sprintf("failed to retrive configset %s (going back to create) ", err.Error())
			r.subState = ProvisioningCreateConfig
			return nextState, message
		}
		commitMessage := "Created and uploaded by O2IMS operator with param CRC" + r.configSetCRC
		err = s.createConfigSetAndUpload(r.meName, r.meSwVer, r.configsetName, configset, commitMessage)

		if err != nil {
			message = fmt.Sprintf("failed to upload configset %s", err.Error())
			retrycount++
			r.apiFailure = err.Error()
			return nextState, message
		} else {
			retrycount = 0
			r.apiFailure = ""
			r.subState = ProvisioningOperationStart
			message = fmt.Sprintf("successfully uploaded configset %s", r.configsetName)
		}
	case ProvisioningOperationStart:
		r.subState = ProvisioningOperationStart
		var operation string
		var optionalLCMParams string
		additionalParams := map[string]interface{}{
			"unattended":                         true,
			"ignoreEquipmentHealthCheckWarnings": true,
			"promptForError":                     true,
			"excludeClusterHealthCheckAlarms":    "9895953,9895955,9895994,9895956,9895941",
			"deleteVpod":                         false,
			"deleteRelay":                        false,
		}
		administrativeState, operationalState, err := s.getMEStatus(r.meName)
		if err != nil {
			retrycount++
			message = fmt.Sprintf("managed element status retrieval failed with (%s),retry count : %d", err.Error(), retrycount)
			r.apiFailure = err.Error()
			return nextState, message
		}
		retrycount = 0
		administrativeState = strings.ToLower(administrativeState)
		operationalState = strings.ToLower(operationalState)
		if administrativeState == "locked" {
			message = fmt.Sprintf("managed element is locked and operational state is %s", operationalState)
			nextState = PendingForPrevious
			return nextState, message
		}
		if operationalState == "defined" {
			operation = "deploy"
			optionalLCMParams = "FIXME Hard coded"
		} else {
			operation = "update"
			optionalLCMParams = "FIXME Hard coded"
		}

		r.workflowId, err = s.LCMRunOperation(r.meName, r.configsetName, operation, optionalLCMParams, additionalParams)
		r.omcOperation = operation
		if err != nil {
			message = fmt.Sprintf("Failed to run install/upgrade operation %s", err.Error())
			retrycount++
			r.apiFailure = err.Error()
			return nextState, message
		} else {
			r.subState = ProvisioningOperationMonitor
			message = fmt.Sprintf("Started upgrade/install operation %s", r.workflowId)
			retrycount = 0
			r.apiFailure = ""
			return nextState, message
		}
	case ProvisioningOperationMonitor:
		administrativeState, operationalState, err := s.getMEStatus(r.meName)
		if err != nil {
			retrycount++
			message = fmt.Sprintf("managed element status retrieval failed with (%s),retry count : %d", err.Error(), retrycount)
			r.apiFailure = err.Error()
			return nextState, message
		}
		retrycount = 0
		administrativeState = strings.ToLower(administrativeState)
		operationalState = strings.ToLower(operationalState)
		if administrativeState == "locked" {
			if r.specChanged {
				message = "spec changed is detected going to pending"
				nextState = PendingForPrevious
			} else {
				message = "waiting for managed element to finish the operation! current state :" + administrativeState + "&" + operationalState

			}
			return nextState, message
		} else {
			r.subState = ""
			if operationalState == "ready" {
				nextState = Completed
				message = "successfully provisioned"
			} else {
				nextState = Error
				message = "Provisioning Failed current state : " + administrativeState + "&" + operationalState
			}
		}

	default:
		nextState = Error
		message = fmt.Sprintf("this is unexpected the current reconciliation state is %s and substate is %s", r.reconcileState, r.subState)
		ErrorLogger.Printf("this is unexpected the current reconciliation state is %s and substate is %s", r.reconcileState, r.subState)

	}

	return nextState, message
}

const (
	//managed element start with New then goes to Defined
	InitSubstateCreateME    string = "CreateME"
	InitSubstateWaitingOnME string = "WaitingOnME"
	InitSubstateMEError     string = "MEError"
)

const (
	// managed element start with New then goes to Defined
	//DeleteSubstatePendingForPreviousOperation string = "PendingForPreviousOperation"
	DeleteSubstateUndeploying           string = "Undeploying"
	DeleteSubstateWaitingForUndeploying string = "WaitingForUndeploying"
	DeleteSubstateRemovingME            string = "RemovingME"
)

func (r *reconcileData) handleDelete(s *ProvisioningRequest) (string, string) {

	var retrycount int
	nextState := Removing
	retrycount, _ = strconv.Atoi(r.apiRetryCount)
	var message string

	defer func() {
		r.apiRetryCount = strconv.Itoa(retrycount)
	}()

	// Check for the status of the managed element
	administrativeState,
		operationalState,
		err := s.getMEStatus(r.meName)

	//FIXME get a better way
	if err != nil && strings.Contains(err.Error(), "httpstatus: 404") {
		retrycount = 0
		r.apiFailure = ""
		r.subState = ""
		message = "successfully removed"
		return Deleted, message
	}

	if err != nil {
		retrycount++
		// Retrying, so retain the substate and state
		// We are still in the same state
		ErrorLogger.Printf("failed to get managed element status: %v", err)
		message = fmt.Sprintf("failed to create managed element status %s, retry count: %d",
			err.Error(), retrycount)
		r.apiFailure = err.Error()
		//mask the err
		err = nil
		return nextState, message
	} else {
		retrycount = 0
		r.apiFailure = ""
		// Successfully created managed element, we will bump of the substate to check for me can be used for
		// next operation until then retain the state
		if administrativeState == "locked" {
			if r.subState != DeleteSubstateWaitingForUndeploying {
				nextState = PendingForPrevious
				message = fmt.Sprintf("managed element is locked (before removing), current state: %s & %s", administrativeState, operationalState)
				return nextState, message
			}
			//stay in DeleteSubstateWaitingForUndeploying
			message := fmt.Sprintf("waiting for undeloy to finish before removing managed element, current state: %s & %s", administrativeState, operationalState)
			return Removing, message
		}
	}

	//Unlocked and other oper states
	if operationalState == "defined" {
		r.subState = DeleteSubstateRemovingME
	}

	switch r.subState {
	case "", DeleteSubstateUndeploying:
		r.subState = DeleteSubstateUndeploying

		InfoLogger.Printf("XXXXXXX add logic to do a  undeploy for %s", r.meName)

		operation := "undeploy"
		optionalLCMParams := ""
		additionalParams := map[string]interface{}{
			"unmanageCompute": true,
			"promptForError":  true,
			"deleteVpod":      true,
			"deleteRelay":     true,
		}
		configset := "master"

		r.workflowId, err = s.LCMRunOperation(r.meName,
			configset,
			operation,
			optionalLCMParams,
			additionalParams)
		r.omcOperation = operation
		if err != nil {
			message = fmt.Sprintf("Failed to run undeploy operation %s",
				err.Error())
			retrycount++
			r.apiFailure = err.Error()
			return Removing, message
		} else {
			r.subState = DeleteSubstateWaitingForUndeploying
			message = fmt.Sprintf("Started undeploy operation %s", r.workflowId)
			retrycount = 0
			r.apiFailure = ""
			return Removing, message
		}
	case DeleteSubstateWaitingForUndeploying:
		//we reach where when ME is unlocked indicating undeploy is done
		fallthrough
	case DeleteSubstateRemovingME:
		_, err := s.remoteService.DeleteME(r.meName)
		if err != nil {
			retrycount++
			// Retrying, so retain the substate and state
			// We are still in the same state
			ErrorLogger.Printf("failed to remove managed element: %v", err)
			message := fmt.Sprintf("failed to remove managed element %s, retry count: %d",
				err.Error(), retrycount)
			r.apiFailure = err.Error()
			//mask the err
			err = nil
			r.subState = DeleteSubstateRemovingME
			return Removing, message
		} else {
			r.subState = ""
			message := "successfully removed"
			return Deleted, message
		}
	}
	return Removing, "Should not have come here"
}

func (r *reconcileData) handleInitState(s *ProvisioningRequest, isMarkedForDeletion bool) (string, string) {
	//   +-----------------+
	//   |   <CreateME>    |
	//   +-----------------+
	//   | Retry CreateME  |
	//   | until success   |
	//   +-----------------+
	//           |
	//           v
	//   +-----------------+
	//   | <WaitingOnME>   | << double  difference between Documentation and code
	//   +-----------------+  << seems there is no new state in the code
	//   |   Retry until   |
	//   |   Define or     |
	//   |   Ready or      |
	//   |   Error 		   |
	//   +-----------------+
	//           |
	//           v
	//   +-----------------+
	//   | <Next State>	   |
	//   +-----------------+
	//
	//   Note: Max Retry falurs moves it error state
	//

	if isMarkedForDeletion {
		return Removing, "Marked For Deletion"
	}
	var retrycount int
	//var err error
	var nextState, message string

	nextState = Init

	defer func() {
		r.apiRetryCount = strconv.Itoa(retrycount)
	}()

	retrycount, _ = strconv.Atoi(r.apiRetryCount)

	switch r.subState {
	case InitSubstateCreateME, "":
		r.subState = InitSubstateCreateME

		InfoLogger.Printf("In Init Substate CreateME")

		templateParameters := s.Spec["templateParameters"].(map[string]interface{})

		//TBC
		templateName, templateVersion, ok := s.Spec["templateName"].(string), s.Spec["templateVersion"].(string), true
		if !ok {
			nextState = Error
			message = "failed to extract template name and version from spec"
			ErrorLogger.Printf("failed to extract template name and version from spec: %v", s.Spec)
			return nextState, message
		}
		r.reconTemplName, r.reconTemplVersion = templateName, templateVersion
		r.meDescription = s.Spec["description"].(string)

		me, err := s.remoteService.GetMEDetaislFromO2imsTemplate(r.reconTemplName, r.reconTemplVersion, templateParameters)
		if err != nil {
			nextState = InstallUpdateError
			message = fmt.Sprintf("failed to get managed element details: %v", err)
			ErrorLogger.Printf("failed to get managed element details: %v", err)
			return nextState, message
		}

		r.meProductType = me["product"].(string)
		r.meFlavorType = me["type"].(string)

		//now where uised now
		r.meSwVer = me["software_version"].(string)

		err = s.createMEWithParams(r.meName,
			r.meDescription,
			r.meProductType,
			r.meFlavorType)

		//hand situation where managed element is already created etc

		if err != nil {
			retrycount++
			// Retrying, so retain the substate and state
			// We are still in the same state
			ErrorLogger.Printf("failed to create managed element: %v", err)
			message = fmt.Sprintf("failed to create managed element %s, retry count: %d",
				err.Error(), retrycount)
			r.apiFailure = err.Error()
			//mask the err
			err = nil
			return nextState, message
		} else {
			retrycount = 0
			r.apiFailure = ""
			// Successfully created managed element, we will bump of the substate to check for me can be used for
			// next operation until then retain the state
			r.subState = InitSubstateWaitingOnME
			message = fmt.Sprintf("successfully created managed element  %s", r.meName)
			InfoLogger.Printf("successfully created managed element %s", r.meName)
		}

		// We are now waiting on managed element to be in defined state
		// so we can use it for next operation
		// We are still in the same state
		// Note: we are checking managed element state in the same reconcile polling
		fallthrough

	//Same as pending wait but we expect thing to be clean
	case InitSubstateWaitingOnME:
		r.subState = InitSubstateWaitingOnME

		administrativeState,
			operationalState,
			err := s.getMEStatus(r.meName)

		if err != nil {
			retrycount++
			message = fmt.Sprintf("failed to get managed element %s, retry count: %d",
				err.Error(), retrycount)
			r.apiFailure = err.Error()
			return nextState, message
		}

		retrycount = 0
		r.apiFailure = ""

		if administrativeState == "unlocked" &&
			(operationalState == "ready" ||
				operationalState == "defined" ||
				operationalState == "error") {
			//next state from here is alway Install
			//if me exist and state is ready then
			//we can call it updating (now provisioning)
			nextState = Provisioning
			message = "Init state the request is succssfull"
			//transition to next state so substate has no meaning
			r.subState = ""
		}

	default:
		message = fmt.Sprintf("Invalid Substate %s", r.subState)
		nextState = Error
	}
	return nextState, message
}

// handlePendingForPrevious is the state handler for PendingForPrevious substate
func (r *reconcileData) handlePendingForPrevious(s *ProvisioningRequest, isMarkedForDeletion bool) (string, string) {
	var retrycount int
	var err error
	var nextState, message string

	nextState = PendingForPrevious

	defer func() {
		r.apiRetryCount = strconv.Itoa(retrycount)
	}()

	retrycount, _ = strconv.Atoi(r.apiRetryCount)

	administrativeState,
		operationalState,
		err := s.getMEStatus(r.meName)

	if err != nil {
		retrycount++
		message = fmt.Sprintf("failure in getting me status %s, retry count: %d",
			err.Error(), retrycount)
		r.apiFailure = err.Error()
		return nextState, message

	}
	retrycount = 0
	administrativeState = strings.ToLower(administrativeState)
	operationalState = strings.ToLower(operationalState)
	//if the managed element is locked then continue monitoring
	if administrativeState == "locked" {
		message = fmt.Sprintf("waiting for managed element to be unlocked! current state: %s %s", administrativeState, operationalState)

		return nextState, message
	} else {
		r.subState = ""
		//
		// we came here only because we wanted to do some install update or delete operation
		if operationalState == "ready" || operationalState == "defined" || operationalState == "error" {
			message = "previous pending operation is completed "
			nextState = Provisioning
			if isMarkedForDeletion {
				return Removing, "Marked For Deletion"
			}
		}
	}
	return nextState, message
}

func (r *reconcileData) handleCompletedForError(s *ProvisioningRequest,
	currentState string,
	isMarkedForDeletion bool) (string, string) {

	// Completed and Error states:
	//
	// +-------------------------------+
	// |   Completed / Error   this is |
	// |    a stable state             |
	// +-------------------------------+
	// |   it can go to InstallUpdate  |
	// |   or Install cause is spec    |
	// |   change                      |
	// |                               |
	// |   it can also go to Removing  |
	// |   case is marked for deletion |
	// |                               |
	// |   it can just stay here       |
	// +-------------------------------+

	_ = s
	nextState := currentState
	message := fmt.Sprintf(" Request in  %s state", currentState)

	if r.specChanged {
		nextState = Provisioning // we can differntiate between install and update later
		message = fmt.Sprintf("spec changed detected in %s state", currentState)
	}

	// Deletion overrides any other state but we wait for previous
	if isMarkedForDeletion {
		nextState = Removing
		message = "Marked For Deletion"
	}
	return nextState, message
}

func (s *ProvisioningRequest) Reconcile() error {

	r := initReconcileData()

	err := s.checkManadatoryFields()
	if err != nil {
		ErrorLogger.Printf("error in checkManadatoryFields %s\n", err.Error())
		return err
	}

	var message string
	err = nil
	var nextState string

	defer func() {
		// Update Internal Reconciliation Info
		reconcileUpdate := make(map[string]interface{})
		updateTime := getCurrentTime()
		//fmt.Printf("Reconcile %s %s\n", r.reconcileState, nextState)

		if r.reconcileState != nextState {
			reconcileUpdate[ReconciliationState] = nextState
			reconcileUpdate[TransitionTime] = updateTime
			r.apiFailure = ""
			if nextState != PendingForPrevious {
				r.workflowId = ""
				r.omcOperation = ""
			}
			r.apiRetryCount = "0"
			r.apiBackoffTime = time.Now()
			r.subState = ""

			if nextState == Completed || nextState == Error || nextState == Deleted {
				//fmt.Printf("Skipping Terminated %v state : %v", r.meName, nextState)
				reconcileUpdate[EndTime] = updateTime
			}
		}

		if r.genConfigset {
			reconcileUpdate[ConfigSetName] = r.configsetName
			reconcileUpdate[ConfigSetCrc] = r.configSetCRC
			reconcileUpdate[TemplateParamsApplied] = true
			r.genConfigset = false
		}

		// we have detected spec changed and taken necessary action
		// update the spec info in recon
		if r.specChanged {
			reconcileUpdate[TemplateName] = r.specTmplName
			reconcileUpdate[TemplateVersion] = r.specTmplVer
			reconcileUpdate[TemplateParamsApplied] = false
			crc := getParamCRC(r.specTemplParam)
			reconcileUpdate[TemplateParamsCRC] = crc

		}

		//Updated for every reconcile
		reconcileUpdate[LastUpdateTime] = updateTime
		reconcileUpdate[ApiRetryCount] = r.apiRetryCount
		reconcileUpdate[SubState] = r.subState
		reconcileUpdate[ApiFailure] = r.apiFailure
		reconcileUpdate[WorkflowId] = r.workflowId
		reconcileUpdate[OmcOperation] = r.omcOperation

		_ = s.updateReconcileInfo(reconcileUpdate)
		var provisioningState string

		//Determine Provisioning Status
		switch nextState {
		case Init, PendingForPrevious, Provisioning:
			provisioningState = "progressing"
		case Completed:
			provisioningState = "fulfilled"
		case Removing:
			provisioningState = "deleting"
		case Error:
			provisioningState = "failed"
		default:
			provisioningState = "unknown"
		}

		newProvStatus := make(map[string]interface{})
		newProvStatus["provisioningMessage"] = message
		newProvStatus["provisioningState"] = provisioningState
		newProvStatus["provisioningUpdateTime"] = updateTime

		_ = s.UpdateProvisioningStatus(newProvStatus)
		//	}
		fmt.Printf("status: %v\n", s.Status)
	}()

	s.checkRemoteService()
	rInfo, _ := s.GetReconciliationInfo()
	if rInfo[MarkedForDeletion].(bool) {
		r.markedForDeletion = true
	}
	r.populateFromReconciliationInfo(rInfo)

	// the spec may cange any time for any state
	// so keep lookig for the spec change
	// however if we are in waiting for pending
	// we do not care we just ignore  the changes

	if (r.reconcileState != PendingForPrevious) &&
		(r.reconcileState != Removing) {
		r.specTmplName,
			r.specTmplVer,
			r.specTemplParam, _ = s.GetTemplateNameVersionAndParams(s.Spec)

		r.specChanged, _ = s.specChanged(r.specTmplName,
			r.specTmplVer,
			r.reconTemplName,
			r.reconTemplVersion,
			r.specTemplParam,
			r.reconTemplCrc)

		if r.specChanged {
			err = s.validateTemplates(r.specTmplName, r.specTmplVer)
			if err != nil {
				nextState = Error
				message = "template validation failed! Unsupported valid Template Name or Version"
				return nil
			}

			err = s.validateTemplateParams(r.specTmplName,
				r.specTmplVer,
				r.specTemplParam)
			if err != nil {
				nextState = Error
				message = fmt.Sprintf("unsupported Params in template %s",
					err.Error())
				return nil
			}
		}
	}
	InfoLogger.Printf("ReconcileState: %s, SubState: %s, markedForDeletion: %v, specChanged: %v",
		r.reconcileState,
		r.subState,
		r.markedForDeletion,
		r.specChanged)

	switch r.reconcileState {
	case Init:
		// handle the init state
		// think about restartability
		nextState, message = r.handleInitState(s, r.markedForDeletion)
		return nil

		// Entry :
		// Data (used)
		// Substate
		// Exit	:
		// Data (altered)
		// Next States
	case Provisioning:
		// handle the Installing and Updating state
		isInstall := r.reconcileState == Provisioning
		nextState, message = r.handleProvisioning(s, r.markedForDeletion, isInstall)
		return nil

		// Entry:
		// - Prior to the run operation if the managed element is found in a non running satete
		// (Install/Update or Remove).
		// Substate:
		// Data (used):
		// - Managed element's admin and operational state (e.g., Ready, Defined, Error).
		// - Flags
		// Exit:
		// - Ensures the managed element is left in a valid state (Ready, Defined, or Error).
		// Data (altered):
		// Next States:
		// - If the managed element is marked for deletion, transition to the "Removing" state.
		// - Otherwise, determine the next state based on the operation requirements:
		//   - Transition to "Install/Update" state

	case PendingForPrevious:
		nextState = r.reconcileState
		nextState, message = r.handlePendingForPrevious(s, r.markedForDeletion)

		//Can we fall through ??
		return nil

	case Completed, Error:
		nextState = r.reconcileState
		if r.specChanged {
			nextState, message = r.handleCompletedForError(s, nextState, r.markedForDeletion)
			return nil
		}
		if r.markedForDeletion {
			nextState = Removing
			message = "Marked for deletion"
		}

	case Removing:
		//var operation string
		//var optionalLCMParams string
		//r.workflowId
		//additionalParams := make(map[string]string)
		nextState = r.reconcileState
		nextState, message = r.handleDelete(s)

		//initiate removal
	case Deleted:
		nextState = Deleted
	default:
		if r.markedForDeletion {
			nextState = Removing
			message = "Marked for deletion"
		} else {
			nextState = Error
			message = "Unknown state for provisioning we can do more recovery tricks here"
		}
	}
	return nil
}
