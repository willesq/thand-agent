package daemon

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/common"
	"github.com/thand-io/agent/internal/config"
	"github.com/thand-io/agent/internal/models"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
)

type ExecutionsPageData struct {
	config.TemplateData
	Executions []*models.WorkflowExecutionInfo `json:"executions"`
}

// listRunningWorkflows lists all running workflow executions
//
//	@Summary		List workflow executions
//	@Description	Get a list of all running workflow executions for the authenticated user
//	@Tags			executions
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	ExecutionsPageData		"List of workflow executions"
//	@Failure		400	{object}	map[string]any	"Bad request"
//	@Failure		401	{object}	map[string]any	"Unauthorized"
//	@Failure		500	{object}	map[string]any	"Internal server error"
//	@Router			/executions [get]
//	@Security		BearerAuth
func (s *Server) listRunningWorkflows(c *gin.Context) {

	ctx := context.Background()

	temporalService := s.Config.GetServices().GetTemporal()

	if temporalService == nil || !temporalService.HasClient() {
		s.getErrorPage(c, http.StatusBadRequest, "Temporal service is not configured")
		return
	}

	if !s.Config.IsServer() {
		// In non-server mode we can assume a default user
		// TODO: Proxy request to server
		s.getErrorPage(c, http.StatusBadRequest, "Workflow listing is only available in server mode")
		return
	}

	_, foundUser, err := s.getUser(c)

	if err != nil {
		s.getErrorPage(c, http.StatusUnauthorized, "Unauthorized: unable to get user for list of running workflows", err)
		return
	}

	if foundUser == nil || foundUser.User == nil || len(foundUser.User.Email) == 0 {
		s.getErrorPage(c, http.StatusUnauthorized, "Unauthorized: user information is incomplete", nil)
		return
	}

	temporalClient := temporalService.GetClient()

	resp, err := temporalClient.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
		Namespace: temporalService.GetNamespace(),
		PageSize:  100,
		Query:     fmt.Sprintf("TaskQueue='%s' AND user='%s'", temporalService.GetTaskQueue(), foundUser.User.Email),
		//NextPageToken: nextPageToken,
	})

	if err != nil {
		s.getErrorPage(c, http.StatusInternalServerError, "Failed to list workflows", err)
		return
	}

	runningWorkflows := []*models.WorkflowExecutionInfo{}

	for _, exec := range resp.Executions {
		runningWorkflows = append(
			runningWorkflows, workflowExecutionInfo(exec))
	}

	response := ExecutionsPageData{
		TemplateData: s.GetTemplateData(c),
		Executions:   runningWorkflows,
	}

	if s.canAcceptHtml(c) {

		s.renderHtml(c, "executions.html", response)

	} else {

		c.JSON(http.StatusOK, response)
	}

}

// createWorkflow creates a new workflow execution
//
//	@Summary		Create workflow execution
//	@Description	Create and start a new workflow execution
//	@Tags			executions
//	@Accept			json
//	@Produce		json
//	@Param			workflow	body		map[string]any	true	"Workflow creation request"
//	@Success		201			{object}	map[string]any	"Workflow created"
//	@Failure		400			{object}	map[string]any	"Bad request"
//	@Router			/execution [post]
//	@Security		BearerAuth
func (s *Server) createWorkflow(c *gin.Context) {
	// TODO: Implement workflow creation logic
}

// getRunningWorkflow retrieves details of a specific workflow execution
//
//	@Summary		Get workflow execution
//	@Description	Retrieve detailed information about a specific workflow execution
//	@Tags			executions
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string					true	"Workflow execution ID"
//	@Success		200	{object}	ExecutionStatePageData	"Workflow execution details"
//	@Failure		400	{object}	map[string]any	"Bad request"
//	@Failure		401	{object}	map[string]any	"Unauthorized"
//	@Failure		500	{object}	map[string]any	"Internal server error"
//	@Router			/execution/{id} [get]
//	@Security		BearerAuth
func (s *Server) getRunningWorkflow(c *gin.Context) {
	workflowId := c.Param("id")

	if len(workflowId) == 0 {
		s.getErrorPage(c, http.StatusBadRequest, "Workflow ID is required")
		return
	}

	temporal := s.Config.GetServices().GetTemporal()

	if temporal == nil || !temporal.HasClient() {
		s.getErrorPage(c, http.StatusNotImplemented, "Temporal service is not configured")
		return
	}

	if !s.Config.IsServer() {
		// In non-server mode we can assume a default user
		// TODO: Proxy request to server
		s.getErrorPage(c, http.StatusNotImplemented, "Workflow listing is only available in server mode")
		return
	}

	_, _, err := s.getUser(c)

	if err != nil {
		s.getErrorPage(c, http.StatusUnauthorized, "Unauthorized: unable to get user to get workflow information", err)
		return
	}

	data, err := s.getWorkflowExecutionState(c, workflowId)

	if err != nil {
		s.getErrorPage(c, http.StatusInternalServerError, err.Error(), err)
		return
	}

	if s.canAcceptHtml(c) {

		s.renderHtml(c, "execution.html", data)

	} else {

		c.JSON(http.StatusOK, data)
	}
}

func (s *Server) getExecutionsPage(c *gin.Context) {
	s.listRunningWorkflows(c)
}

// getWorkflowExecutionState retrieves the current state of a workflow execution
// and returns it as ExecutionStatePageData ready for rendering or JSON response
func (s *Server) getWorkflowExecutionState(c *gin.Context, workflowID string) (*ExecutionStatePageData, error) {
	ctx := context.Background()

	temporal := s.Config.GetServices().GetTemporal()

	if temporal == nil || !temporal.HasClient() {
		return nil, fmt.Errorf("temporal service is not configured")
	}

	temporalClient := temporal.GetClient()

	// Get the workflow execution information
	wkflw, err := temporalClient.DescribeWorkflowExecution(ctx, workflowID, models.TemporalEmptyRunId)

	if err != nil {
		return nil, fmt.Errorf("failed to get workflow state: %w", err)
	}

	wklwInfo := wkflw.GetWorkflowExecutionInfo()

	if wklwInfo == nil {
		return nil, fmt.Errorf("workflow execution not found")
	}

	workflowExecInfo := workflowExecutionInfo(wklwInfo)

	var workflowTask models.WorkflowTask

	// If workflow hasn't completed, query for the current state
	if workflowExecInfo.CloseTime == nil {

		// Create a timeout context for the query to avoid hanging requests
		timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		queryResponse, err := temporalClient.QueryWorkflowWithOptions(timeoutCtx, &client.QueryWorkflowWithOptionsRequest{
			WorkflowID: workflowID,
			RunID:      models.TemporalEmptyRunId,
			QueryType:  models.TemporalGetWorkflowTaskQueryName,
			Args:       nil,
		})

		if err == nil {

			err = queryResponse.QueryResult.Get(&workflowTask)

			if err != nil {
				return nil, fmt.Errorf("failed to get workflow state: %w", err)
			}

			workflowName := workflowTask.WorkflowName

			if len(workflowName) == 0 {

				elevationReq, err := workflowTask.GetContextAsElevationRequest()

				if err == nil && elevationReq != nil {
					workflowName = elevationReq.Workflow
				}

			}

			// Get the workflow template name if available
			foundWorkflow, err := s.GetConfig().GetWorkflowByName(workflowName)

			if err != nil {
				logrus.WithError(err).Warn("Failed to get workflow definition")
			} else {
				workflowTask.Workflow = foundWorkflow.GetWorkflow()
			}

			// Copy over task status phases to the response
			phases := []string{}
			for phase := range workflowTask.TasksStatusPhase {
				phases = append(phases, phase)
			}
			workflowExecInfo.History = phases

			workflowExecInfo.Input = workflowTask.Input
			workflowExecInfo.Output = workflowTask.Output
			workflowExecInfo.Context = workflowTask.Context

		} else {

			logrus.WithError(err).Warnln("Failed to query workflow for current state")
			workflowExecInfo.Output = err.Error()

		}

	} else if wklwInfo.GetStatus() != enums.WORKFLOW_EXECUTION_STATUS_TERMINATED {

		// Otherwise if the workflow has completed then get the last output

		fut := temporalClient.GetWorkflow(
			ctx, workflowID, models.TemporalEmptyRunId)

		err := fut.Get(ctx, &workflowTask)

		if err != nil {
			logrus.WithError(err).Warnln("Failed to get workflow output")
			workflowExecInfo.Output = err.Error()
		}

	}

	data := &ExecutionStatePageData{
		TemplateData: s.GetTemplateData(c),
		Execution:    workflowExecInfo,
		Workflow:     workflowTask.Workflow,
	}

	return data, nil
}

func workflowExecutionInfo(
	workflowInfo *workflow.WorkflowExecutionInfo,
) *models.WorkflowExecutionInfo {

	exec := workflowInfo.GetExecution()

	searchAttributes := workflowInfo.GetSearchAttributes().GetIndexedFields()

	response := models.WorkflowExecutionInfo{
		WorkflowID: exec.GetWorkflowId(),
		RunID:      exec.GetRunId(),
		StartTime:  workflowInfo.GetStartTime().AsTime(),
		Status:     strings.ToUpper(workflowInfo.GetStatus().String()),
		Identities: []string{},
		History:    []string{},
	}

	if workflowInfo.GetCloseTime() != nil {
		closeTime := workflowInfo.GetCloseTime().AsTime()
		response.CloseTime = &closeTime
	}

	// Safely extract search attributes with proper type conversion
	dataConverter := converter.GetDefaultDataConverter()

	if userAttr, exists := searchAttributes[models.VarsContextUser]; exists && userAttr != nil {
		var userValue string
		if err := dataConverter.FromPayload(userAttr, &userValue); err == nil {
			response.User = userValue
		}
	}

	if roleAttr, exists := searchAttributes[models.VarsContextRole]; exists && roleAttr != nil {
		var roleValue string
		if err := dataConverter.FromPayload(roleAttr, &roleValue); err == nil {
			response.Role = roleValue
		}
	}

	if workflowInfo.GetStatus() == enums.WORKFLOW_EXECUTION_STATUS_RUNNING {
		if workflowStatusAttr, exists := searchAttributes["status"]; exists && workflowStatusAttr != nil {
			var statusValue string
			if err := dataConverter.FromPayload(workflowStatusAttr, &statusValue); err == nil {
				response.Status = strings.ToUpper(statusValue)
			}
		}
	}

	if approvedAttr, exists := searchAttributes[models.VarsContextApproved]; exists && approvedAttr != nil {
		var approvedValue bool
		if err := dataConverter.FromPayload(approvedAttr, &approvedValue); err == nil {
			response.Approved = &approvedValue
		}
	}

	if workflowAttr, exists := searchAttributes[models.VarsContextWorkflow]; exists && workflowAttr != nil {
		var workflowValue string
		if err := dataConverter.FromPayload(workflowAttr, &workflowValue); err == nil {
			response.Workflow = workflowValue
		}
	}

	if taskAttr, exists := searchAttributes["task"]; exists && taskAttr != nil {
		var taskValue string
		if err := dataConverter.FromPayload(taskAttr, &taskValue); err == nil {
			response.Task = taskValue
		}
	}

	if reasonAttr, exists := searchAttributes["reason"]; exists && reasonAttr != nil {
		var reasonValue string
		if err := dataConverter.FromPayload(reasonAttr, &reasonValue); err == nil {
			response.Reason = reasonValue
		}
	}

	if durationAttr, exists := searchAttributes["duration"]; exists && durationAttr != nil {
		var durationValue int64
		if err := dataConverter.FromPayload(durationAttr, &durationValue); err == nil {
			response.Duration = durationValue
		}
	}

	if identitiesAttr, exists := searchAttributes["identities"]; exists && identitiesAttr != nil {
		var identitiesValue []string
		if err := dataConverter.FromPayload(identitiesAttr, &identitiesValue); err == nil {
			response.Identities = identitiesValue
		}
	}

	return &response

}

// signalRunningWorkflow sends a signal to a running workflow
//
//	@Summary		Signal workflow execution
//	@Description	Send a signal event to a running workflow execution
//	@Tags			executions
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"Workflow execution ID"
//	@Param			input	query		string					true	"Encoded signal data"
//	@Success		200		{object}	map[string]any	"Signal sent successfully"
//	@Failure		400		{object}	map[string]any	"Bad request"
//	@Failure		401		{object}	map[string]any	"Unauthorized"
//	@Failure		403		{object}	map[string]any	"Forbidden"
//	@Failure		500		{object}	map[string]any	"Internal server error"
//	@Router			/execution/{id}/signal [get]
//	@Security		BearerAuth
func (s *Server) signalRunningWorkflow(c *gin.Context) {

	workflowId := c.Param("id")
	// get input from the query parameters
	input := c.Query("input")

	if len(input) == 0 {
		s.getErrorPage(c, http.StatusBadRequest, "Input parameter is required")
		return
	}

	if !s.Config.IsServer() {
		s.getErrorPage(c, http.StatusForbidden, "Forbidden: unable to signal workflow in non-server mode", nil)
		return
	}

	if !s.Config.GetServices().HasTemporal() {
		s.getErrorPage(c, http.StatusInternalServerError, "Temporal service is not configured", nil)
		return
	}

	_, foundUser, err := s.getUser(c)

	if err != nil {
		s.getErrorPage(c, http.StatusUnauthorized, "Unauthorized: unable to get user for signaling workflow", err)
		return
	}

	// Convert state to cloudevent Signal
	// Tasks may contain sensitive information, ensure encryption is used
	decodedTask, err := models.EncodingWrapper{}.DecodeAndDecrypt(input, s.Config.GetServices().GetEncryption())

	if err != nil {
		s.getErrorPage(c, http.StatusBadRequest, "Failed to decode workflow state", err)
		return
	}

	if decodedTask.Type != models.ENCODED_WORKFLOW_SIGNAL {
		s.getErrorPage(c, http.StatusBadRequest, fmt.Sprintf("invalid workflow state type: %s", decodedTask.Type), nil)
		return
	}

	var signal cloudevents.Event
	dataMap, ok := decodedTask.Data.(map[string]any)
	if !ok {
		s.getErrorPage(c, http.StatusBadRequest, "Failed to parse workflow state: invalid data type", nil)
		return
	}
	err = common.ConvertMapToInterface(dataMap, &signal)

	if err != nil {
		s.getErrorPage(c, http.StatusBadRequest, "Failed to parse workflow state", err)
		return
	}

	// Extensions only support basic types so we need to set the user identity as a string
	signal.SetExtension(models.VarsContextUser, foundUser.User.GetIdentity())

	if len(signal.FieldErrors) > 0 {
		logrus.WithField("errors", signal.FieldErrors).
			Error("failed to set user extension on cloudevent")
		s.getErrorPage(c, http.StatusBadRequest, fmt.Sprintf("Failed to set user extension on cloudevent: %v", signal.FieldErrors))
		return
	}

	ctx := context.Background()

	serviceClient := s.Config.GetServices()

	temporalService := serviceClient.GetTemporal()
	temporalClient := temporalService.GetClient()

	// Lets signal the workflow to continue
	err = temporalClient.SignalWorkflow(
		ctx, workflowId, models.TemporalEmptyRunId,
		models.TemporalEventSignalName, signal)

	if err != nil {
		s.getErrorPage(c, http.StatusInternalServerError, "Failed to signal workflow", err)
		return
	}

	data, err := s.getWorkflowExecutionState(c, workflowId)

	if err != nil {
		s.getErrorPage(c, http.StatusInternalServerError, err.Error(), err)
		return
	}

	if s.canAcceptHtml(c) {

		s.renderHtml(c, "execution.html", data)

	} else {

		c.JSON(http.StatusOK, data)
	}

}
