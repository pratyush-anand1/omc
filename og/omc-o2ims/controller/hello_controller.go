package controller

// don't delete the swagger annotations

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/enrayga/omc-o2ims/internal/models"
	"github.com/gin-gonic/gin"
)

var (
	file_lock         sync.Mutex
	dataStoreFileName = "provisioning_requests.json"
	provisionRequests = make(map[string]models.ProvisioningRequestInfo)
)

func loadProvisioningRequests() {
	file_lock.Lock()
	defer file_lock.Unlock()

	file, err := os.ReadFile(dataStoreFileName)
	if err == nil {
		json.Unmarshal(file, &provisionRequests)
	}
}

func saveProvisioningRequests() error {
	file_lock.Lock()
	defer file_lock.Unlock()

	data, err := json.MarshalIndent(provisionRequests, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(dataStoreFileName, data, 0644)
}

// @Summary Returns a list of provisioning requests
// @Tags O2ims_infrastructureProvisioning
// @Produce json
// @Success 200 {object} models.ProvisioningRequestList
// @Router /o2ims-infrastructureprovisioning/v1/provisioningrequests [get]
func GetProvisioningRequests(c *gin.Context) {
	items := make([]models.ProvisioningRequestInfo, 0, len(provisionRequests))
	for _, request := range provisionRequests {
		items = append(items, request)
	}
	c.JSON(http.StatusOK, models.ProvisioningRequestList{Items: items})
}

// @Summary Creates a new provisioning request
// @Tags O2ims_infrastructureProvisioning
// @Accept json
// @Produce json
// @Param request body models.ProvisioningRequest true "ProvisioningRequest object"
// @Success 201 {object} models.ProvisioningRequestInfo
// @Router /o2ims-infrastructureprovisioning/v1/provisioningrequests [post]
func CreateProvisioningRequest(c *gin.Context) {
	var request models.ProvisioningRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	provisioningRequestId := fmt.Sprintf("req-%d", time.Now().Unix())
	provisioningRequestInfo := models.ProvisioningRequestInfo{
		ProvisioningRequestId: provisioningRequestId,
		Name:                  request.Name,
		Description:           request.Description,
		TemplateName:          request.TemplateName,
		TemplateVersion:       request.TemplateVersion,
		TemplateParameters:    request.TemplateParameters,
		Status: models.ProvisioningStatus{
			UpdateTime:        time.Now().Format(time.RFC3339),
			Message:           "Provisioning request created",
			ProvisioningPhase: "PENDING",
		},
		ProvisionedResources: models.ProvisionedResourceSet{},
	}

	provisionRequests[provisioningRequestId] = provisioningRequestInfo
	if err := saveProvisioningRequests(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save provisioning request"})
		return
	}
	c.JSON(http.StatusCreated, provisioningRequestInfo)
}

// @Summary Retrieves a single provisioning request
// @Tags O2ims_infrastructureProvisioning
// @Produce json
// @Param provisioningRequestId path string true "ProvisioningRequest ID"
// @Success 200 {object} models.ProvisioningRequestInfo
// @Failure 404 {object} models.ProblemDetails
// @Router /o2ims-infrastructureprovisioning/v1/provisioningrequests/{provisioningRequestId} [get]
func GetProvisioningRequest(c *gin.Context) {
	id := c.Param("provisioningRequestId")
	request, exists := provisionRequests[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Provisioning request not found"})
		return
	}
	c.JSON(http.StatusOK, request)
}

// @Summary Deletes a provisioning request
// @Tags O2ims_infrastructureProvisioning
// @Produce json
// @Param provisioningRequestId path string true "ProvisioningRequest ID"
// @Success 204
// @Failure 404 {object} models.ProblemDetails
// @Router /o2ims-infrastructureprovisioning/v1/provisioningrequests/{provisioningRequestId} [delete]
func DeleteProvisioningRequest(c *gin.Context) {
	id := c.Param("provisioningRequestId")
	if _, exists := provisionRequests[id]; !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Provisioning request not found"})
		return
	}
	delete(provisionRequests, id)
	_ = saveProvisioningRequests()
	c.Status(http.StatusNoContent)
}

func RegisterO2imsRoutes(router *gin.Engine) {
	loadProvisioningRequests()
	o2imsGroup := router.Group("/o2ims-infrastructureprovisioning/v1")
	{
		o2imsGroup.GET("/provisioningrequests", GetProvisioningRequests)
		o2imsGroup.POST("/provisioningrequests", CreateProvisioningRequest)
		o2imsGroup.GET("/provisioningrequests/:provisioningRequestId", GetProvisioningRequest)
		o2imsGroup.DELETE("/provisioningrequests/:provisioningRequestId", DeleteProvisioningRequest)
	}
}
