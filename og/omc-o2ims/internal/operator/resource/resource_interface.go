package resource

type Resource interface {
	GetID() string
	SetInitFields(string, map[string]interface{}) error
	SetDeleteFlag() error
	GetDeleteFlag() bool
	Compare(string, map[string]interface{}, bool) (bool, error)
	GetNew() interface{}
	GetStatus() (map[string]interface{}, error)
	Reconcile() error //TBD
}

type NotifierEvent struct {
	Type    string
	Message string
	Payload map[string]interface{}
}

const (
	Progressing string = "progressing"
	Deleting    string = "deleting"
	Fulfilled   string = "fulfilled"
	Failed      string = "failed"
)

const (
	ReconciliationState   = "reconciliationState"
	SubState              = "subState"
	ApiFailure            = "apiFailure"
	ApiRetryCount         = "apiRetryCount"
	ApiBackOffTime        = "backOffTime"
	MarkedForDeletion     = "markedForDeletion"
	TemplateName          = "templateName"
	TemplateVersion       = "templateVersion"
	TemplateParamsCRC     = "templateParamsCRC"
	TemplateParamsApplied = "templateParamsApplied"
	WorkflowId            = "workflowId"
	ConfigSetName         = "configSetName"
	ConfigSetCrc          = "configSetCrc"
	MeName                = "meName"
	MeDescription         = "meDescription"
	MeProductType         = "meProductType"
	MeFlavorType          = "meFlavorType"
	MeSwVer               = "meSwVer"
	OmcOperation          = "omcOperation"
	StartTime             = "startTime"
	EndTime               = "endTime"
	LastUpdateTime        = "lastUpdateTime"
	TransitionTime        = "transitionTime"
	ReconciliationTimeout = "reconciliationTimeout"
)

const (
	Unknown string = "unknown"
	//The resource has just been created but is not yet processed.
	Init string = "init"
	//Init  -> PendingForPrevious
	//      -> Progressing(Updating, Installing)
	//      -> Error

	//The resource is waiting for previous operations
	PendingForPrevious string = "pendingForPrevious"
	//PendingForPrevious -> TBD Aborting
	//      			 -> Progressing(Updating, Installing)
	//      			 -> Error

	//The resource is actively working towards aligning with the desired
	// state specified by the user.
	// Updating   string = "updating"
	// Installing string = "installing"
	Provisioning string = "provisioning"
	//Progressing  => (Updating, Installing)
	//      → Completed: If the desired state is reached successfully.
	//      → Error: If an issue arises during the process.
	//      → PendingForPrevious: if desired state is changed

	//The resource is aligned with the desired state specified by the user.
	Completed string = "completed"
	//      → Updating : Only if params are changed & ME as expcted is READY
	//      → Installing: RARLY if params are changed & ME as expcted is NEW
	//                     or DEFINED
	//      → Completed: Mostly if there are not param changes
	//      → Deleting: If the resource is deleted
	//		→ Error: If an issue arises during the process.
	//				 OMC api failure
	//               Template issue etc
	//The resource is in the process of being deleted.

	Removing string = "removing"
	// → Deleted: There is not deleted
	// → Error: If the deletion process encounters an issue.
	Deleted string = "deleted"
	Error   string = "error"
)

const (
	ProvisioningCreateConfig     string = "CreateConfig"
	ProvisioningPushConfig       string = "PushConfig"
	ProvisioningOperationStart   string = "OperationStart"
	ProvisioningOperationMonitor string = "OperationMonitor"
	ProvisioningError            string = "Error"
)
