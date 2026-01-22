package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/zjrosen/perles/internal/orchestration/controlplane"
	"github.com/zjrosen/perles/internal/orchestration/controlplane/mocks"
)

// === Tests ===

func TestHandler_Create(t *testing.T) {
	mockCP := mocks.NewMockControlPlane(t)
	mockCP.EXPECT().
		Create(mock.Anything, mock.MatchedBy(func(spec controlplane.WorkflowSpec) bool {
			// Without WorkflowCreator, goal is passed directly as InitialGoal with "# Goal\n\n" prefix
			return spec.TemplateID == "cook" && spec.InitialGoal == "# Goal\n\nBuild feature X"
		})).
		Return(controlplane.WorkflowID("wf-123"), nil).
		Once()

	h := NewHandler(mockCP)

	body := `{"template_id": "cook", "goal": "Build feature X"}`
	req := httptest.NewRequest(http.MethodPost, "/workflows", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Routes().ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var resp CreateWorkflowResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "wf-123", resp.ID)
}

func TestHandler_Create_InvalidJSON(t *testing.T) {
	mockCP := mocks.NewMockControlPlane(t)

	h := NewHandler(mockCP)

	req := httptest.NewRequest(http.MethodPost, "/workflows", bytes.NewBufferString("not json"))
	w := httptest.NewRecorder()

	h.Routes().ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)

	var resp ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "invalid_json", resp.Code)
}

func TestHandler_Get(t *testing.T) {
	mockCP := mocks.NewMockControlPlane(t)
	mockCP.EXPECT().
		Get(mock.Anything, controlplane.WorkflowID("wf-123")).
		Return(&controlplane.WorkflowInstance{
			ID:          "wf-123",
			TemplateID:  "cook",
			State:       controlplane.WorkflowRunning,
			InitialGoal: "Build feature",
			MCPPort:     19001,
		}, nil).
		Once()
	mockCP.EXPECT().
		GetHealthStatus(controlplane.WorkflowID("wf-123")).
		Return(controlplane.HealthStatus{
			WorkflowID: "wf-123",
			IsHealthy:  true,
		}, true).
		Once()

	h := NewHandler(mockCP)

	req := httptest.NewRequest(http.MethodGet, "/workflows/wf-123", nil)
	w := httptest.NewRecorder()

	h.Routes().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp WorkflowResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "wf-123", resp.ID)
	assert.Equal(t, "running", resp.State)
	assert.Equal(t, 19001, resp.Port)
	assert.True(t, resp.IsHealthy)
}

func TestHandler_Get_NotFound(t *testing.T) {
	mockCP := mocks.NewMockControlPlane(t)
	mockCP.EXPECT().
		Get(mock.Anything, controlplane.WorkflowID("unknown")).
		Return(nil, controlplane.ErrWorkflowNotFound).
		Once()

	h := NewHandler(mockCP)

	req := httptest.NewRequest(http.MethodGet, "/workflows/unknown", nil)
	w := httptest.NewRecorder()

	h.Routes().ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandler_Start(t *testing.T) {
	mockCP := mocks.NewMockControlPlane(t)
	mockCP.EXPECT().
		Start(mock.Anything, controlplane.WorkflowID("wf-123")).
		Return(nil).
		Once()

	h := NewHandler(mockCP)

	req := httptest.NewRequest(http.MethodPost, "/workflows/wf-123/start", nil)
	w := httptest.NewRecorder()

	h.Routes().ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
}

func TestHandler_Stop(t *testing.T) {
	mockCP := mocks.NewMockControlPlane(t)
	mockCP.EXPECT().
		Stop(mock.Anything, controlplane.WorkflowID("wf-123"), mock.MatchedBy(func(opts controlplane.StopOptions) bool {
			return opts.Reason == "user requested" && opts.Force == true
		})).
		Return(nil).
		Once()

	h := NewHandler(mockCP)

	body := `{"reason": "user requested", "force": true}`
	req := httptest.NewRequest(http.MethodPost, "/workflows/wf-123/stop", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Routes().ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
}

func TestHandler_List(t *testing.T) {
	mockCP := mocks.NewMockControlPlane(t)
	mockCP.EXPECT().
		List(mock.Anything, mock.Anything).
		Return([]*controlplane.WorkflowInstance{
			{ID: "wf-1", TemplateID: "cook", State: controlplane.WorkflowRunning},
			{ID: "wf-2", TemplateID: "plan", State: controlplane.WorkflowPending},
		}, nil).
		Once()
	mockCP.EXPECT().
		GetHealthStatus(controlplane.WorkflowID("wf-1")).
		Return(controlplane.HealthStatus{WorkflowID: "wf-1", IsHealthy: true}, true).
		Once()
	mockCP.EXPECT().
		GetHealthStatus(controlplane.WorkflowID("wf-2")).
		Return(controlplane.HealthStatus{WorkflowID: "wf-2", IsHealthy: false}, true).
		Once()

	h := NewHandler(mockCP)

	req := httptest.NewRequest(http.MethodGet, "/workflows", nil)
	w := httptest.NewRecorder()

	h.Routes().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp ListWorkflowsResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 2, resp.Total)
	assert.Len(t, resp.Workflows, 2)
}

func TestHandler_Health(t *testing.T) {
	mockCP := mocks.NewMockControlPlane(t)
	mockCP.EXPECT().
		List(mock.Anything, mock.Anything).
		Return([]*controlplane.WorkflowInstance{
			{ID: "wf-1", Name: "Test Workflow", State: controlplane.WorkflowRunning},
		}, nil).
		Once()
	mockCP.EXPECT().
		GetHealthStatus(controlplane.WorkflowID("wf-1")).
		Return(controlplane.HealthStatus{
			WorkflowID: "wf-1",
			IsHealthy:  true,
		}, true).
		Once()

	h := NewHandler(mockCP)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	h.Routes().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.Status)
	require.Len(t, resp.Workflows, 1)
	assert.Equal(t, "wf-1", resp.Workflows[0].ID)
	assert.True(t, resp.Workflows[0].IsHealthy)
}
