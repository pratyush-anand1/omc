package omc_rest

// RemoteClient is an interface for interacting with a remote service.
type OMCRemoteService interface {
	GetO2imsTemplateList() (map[string]interface{}, error)
	CheckO2imsTemplateSupport(templateName,
		templateVersion string) (map[string]string,
		error)
	VerifyO2imsTemplateParams(templateName,
		templateVersion string,
		params map[string]interface{}) (map[string]string, error)
	GenConfigsetFromO2imsTemplate(templateName,
		templateVersion string,
		params map[string]interface{}) (map[string]interface{},
		error)
	GetMEDetaislFromO2imsTemplate(templateName,
		templateVersion string,
		params map[string]interface{}) (map[string]interface{}, error)

	GetAllME() ([]map[string]interface{}, error)
	DeleteME(meName string) (map[string]string, error)
	GetME(meName string) (map[string]interface{}, error)

	CreateME(payload map[string]string) error
	GetActiveWFListOfMe(meName string) (string, error)
	UpdateME(meName, state, softwareVersion, activeOperations string) error

	ListConfigSets(meName string) (map[string]interface{}, error)
	CreateConfigSet(meName string,
		payload map[string]string) (map[string]string,
		error)
	DeleteConfigSet(meName,
		configSetName string) (map[string]string,
		error)
	UploadConfigSetFile(meName,
		configSetName,
		commitMessage string,
		yaml []byte) error

	GetLCMOperList() (map[string]interface{}, error)
	RunLCMOper(payload map[string]interface{}) (map[string]string, error)

	GetWorkflow(workflowId string) (map[string]string, error)
	UpdateWorkflow(workflowId string, workflow map[string]string) error
	UpdateWorkflowStatus(workflowId, state string) error
}
