package mcp

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/zjrosen/perles/internal/mocks"
	"github.com/zjrosen/perles/internal/orchestration/claude"
	"github.com/zjrosen/perles/internal/orchestration/events"
	"github.com/zjrosen/perles/internal/orchestration/message"
	"github.com/zjrosen/perles/internal/orchestration/pool"
)

// ============================================================================
// Integration Tests with Mock BD
// ============================================================================

// TestIntegration_FullTaskLifecycle tests the complete lifecycle of a task from assignment to completion.
func TestIntegration_FullTaskLifecycle(t *testing.T) {
	workerPool := pool.NewWorkerPool(pool.Config{})
	defer workerPool.Close()

	msgIssue := message.New()
	cs := NewCoordinatorServer(claude.NewClient(), workerPool, msgIssue, "/tmp/test", 8765, nil, mocks.NewMockBeadsExecutor(t))

	implementerID := "worker-1"
	reviewerID := "worker-2"
	taskID := "perles-abc.1"

	// Create workers
	_ = workerPool.AddTestWorker(implementerID, pool.WorkerReady)
	_ = workerPool.AddTestWorker(reviewerID, pool.WorkerReady)

	// Step 1: Assign task to implementer
	cs.assignmentsMu.Lock()
	cs.workerAssignments[implementerID] = &WorkerAssignment{
		TaskID:     taskID,
		Role:       RoleImplementer,
		Phase:      events.PhaseImplementing,
		AssignedAt: time.Now(),
	}
	cs.taskAssignments[taskID] = &TaskAssignment{
		TaskID:      taskID,
		Implementer: implementerID,
		Status:      TaskImplementing,
		StartedAt:   time.Now(),
	}
	cs.assignmentsMu.Unlock()

	// Verify state
	cs.assignmentsMu.RLock()
	require.Equal(t, events.PhaseImplementing, cs.workerAssignments[implementerID].Phase, "Phase mismatch")
	cs.assignmentsMu.RUnlock()

	// Step 2: Implementer completes implementation
	workerStore := newMockMessageStore()
	ws := NewWorkerServer(implementerID, workerStore)
	callback := &coordinatorStateCallback{cs: cs}
	ws.SetStateCallback(callback)

	handler := ws.handlers["report_implementation_complete"]
	_, err := handler(context.Background(), json.RawMessage(`{"summary": "Implemented feature X"}`))
	require.NoError(t, err, "report_implementation_complete failed")

	// Verify transition to awaiting review
	cs.assignmentsMu.RLock()
	require.Equal(t, events.PhaseAwaitingReview, cs.workerAssignments[implementerID].Phase, "Phase mismatch")
	cs.assignmentsMu.RUnlock()

	// Step 3: Assign reviewer
	cs.assignmentsMu.Lock()
	cs.workerAssignments[reviewerID] = &WorkerAssignment{
		TaskID:        taskID,
		Role:          RoleReviewer,
		Phase:         events.PhaseReviewing,
		AssignedAt:    time.Now(),
		ImplementerID: implementerID,
	}
	cs.taskAssignments[taskID].Reviewer = reviewerID
	cs.taskAssignments[taskID].Status = TaskInReview
	cs.taskAssignments[taskID].ReviewStartedAt = time.Now()
	cs.assignmentsMu.Unlock()

	// Step 4: Reviewer approves
	reviewerStore := newMockMessageStore()
	reviewerWs := NewWorkerServer(reviewerID, reviewerStore)
	reviewerWs.SetStateCallback(callback)

	reviewHandler := reviewerWs.handlers["report_review_verdict"]
	_, err = reviewHandler(context.Background(), json.RawMessage(`{"verdict": "APPROVED", "comments": "LGTM"}`))
	require.NoError(t, err, "report_review_verdict failed")

	// Verify reviewer back to idle
	cs.assignmentsMu.RLock()
	require.Equal(t, events.PhaseIdle, cs.workerAssignments[reviewerID].Phase, "Reviewer phase mismatch")
	cs.assignmentsMu.RUnlock()

	// Step 5: Approve commit (simulated coordinator action)
	cs.assignmentsMu.Lock()
	cs.taskAssignments[taskID].Status = TaskApproved
	cs.assignmentsMu.Unlock()

	// Step 6: Implementer commits
	cs.assignmentsMu.Lock()
	cs.workerAssignments[implementerID].Phase = events.PhaseCommitting
	cs.taskAssignments[taskID].Status = TaskCommitting
	cs.assignmentsMu.Unlock()

	// Step 7: Commit complete -> task done
	cs.assignmentsMu.Lock()
	cs.workerAssignments[implementerID].Phase = events.PhaseIdle
	cs.workerAssignments[implementerID].TaskID = ""
	cs.taskAssignments[taskID].Status = TaskCompleted
	cs.assignmentsMu.Unlock()

	// Final verification
	cs.assignmentsMu.RLock()
	require.Equal(t, TaskCompleted, cs.taskAssignments[taskID].Status, "Task status mismatch")
	require.Equal(t, events.PhaseIdle, cs.workerAssignments[implementerID].Phase, "Implementer phase mismatch")
	cs.assignmentsMu.RUnlock()
}

// TestIntegration_DenialCycle tests the denial -> feedback -> re-review cycle.
func TestIntegration_DenialCycle(t *testing.T) {
	workerPool := pool.NewWorkerPool(pool.Config{})
	defer workerPool.Close()

	msgIssue := message.New()
	cs := NewCoordinatorServer(claude.NewClient(), workerPool, msgIssue, "/tmp/test", 8765, nil, mocks.NewMockBeadsExecutor(t))

	implementerID := "worker-1"
	reviewerID := "worker-2"
	taskID := "perles-abc.1"

	// Create workers
	_ = workerPool.AddTestWorker(implementerID, pool.WorkerReady)
	_ = workerPool.AddTestWorker(reviewerID, pool.WorkerReady)

	// Setup: implementer has completed, reviewer is reviewing
	cs.assignmentsMu.Lock()
	cs.workerAssignments[implementerID] = &WorkerAssignment{
		TaskID:     taskID,
		Role:       RoleImplementer,
		Phase:      events.PhaseAwaitingReview,
		AssignedAt: time.Now(),
	}
	cs.workerAssignments[reviewerID] = &WorkerAssignment{
		TaskID:        taskID,
		Role:          RoleReviewer,
		Phase:         events.PhaseReviewing,
		AssignedAt:    time.Now(),
		ImplementerID: implementerID,
	}
	cs.taskAssignments[taskID] = &TaskAssignment{
		TaskID:          taskID,
		Implementer:     implementerID,
		Reviewer:        reviewerID,
		Status:          TaskInReview,
		StartedAt:       time.Now(),
		ReviewStartedAt: time.Now(),
	}
	cs.assignmentsMu.Unlock()

	// Reviewer denies
	reviewerStore := newMockMessageStore()
	reviewerWs := NewWorkerServer(reviewerID, reviewerStore)
	callback := &denialCycleCallback{cs: cs}
	reviewerWs.SetStateCallback(callback)

	reviewHandler := reviewerWs.handlers["report_review_verdict"]
	_, err := reviewHandler(context.Background(), json.RawMessage(`{"verdict": "DENIED", "comments": "Missing error handling"}`))
	require.NoError(t, err, "report_review_verdict failed")

	// Verify task status is denied and implementer gets feedback assignment
	cs.assignmentsMu.RLock()
	require.Equal(t, TaskDenied, cs.taskAssignments[taskID].Status, "Task status mismatch")
	cs.assignmentsMu.RUnlock()

	// Coordinator assigns feedback to implementer
	cs.assignmentsMu.Lock()
	cs.workerAssignments[implementerID].Phase = events.PhaseAddressingFeedback
	cs.assignmentsMu.Unlock()

	// Implementer addresses feedback and re-submits
	implementerStore := newMockMessageStore()
	implementerWs := NewWorkerServer(implementerID, implementerStore)
	feedbackCallback := &coordinatorStateCallback{cs: cs}
	implementerWs.SetStateCallback(feedbackCallback)

	implHandler := implementerWs.handlers["report_implementation_complete"]
	_, err = implHandler(context.Background(), json.RawMessage(`{"summary": "Fixed error handling"}`))
	require.NoError(t, err, "report_implementation_complete after feedback failed")

	// Verify back to awaiting review
	cs.assignmentsMu.RLock()
	require.Equal(t, events.PhaseAwaitingReview, cs.workerAssignments[implementerID].Phase, "Phase mismatch")
	cs.assignmentsMu.RUnlock()
}

// denialCycleCallback implements WorkerStateCallback for denial testing.
type denialCycleCallback struct {
	cs *CoordinatorServer
}

func (c *denialCycleCallback) GetWorkerPhase(workerID string) (events.WorkerPhase, error) {
	c.cs.assignmentsMu.RLock()
	defer c.cs.assignmentsMu.RUnlock()

	wa, ok := c.cs.workerAssignments[workerID]
	if !ok {
		return events.PhaseIdle, nil
	}
	return wa.Phase, nil
}

func (c *denialCycleCallback) OnImplementationComplete(workerID, summary string) error {
	c.cs.assignmentsMu.Lock()
	defer c.cs.assignmentsMu.Unlock()

	wa := c.cs.workerAssignments[workerID]
	wa.Phase = events.PhaseAwaitingReview

	if ta, ok := c.cs.taskAssignments[wa.TaskID]; ok {
		ta.Status = TaskInReview
	}

	return nil
}

func (c *denialCycleCallback) OnReviewVerdict(workerID, verdict, comments string) error {
	c.cs.assignmentsMu.Lock()
	defer c.cs.assignmentsMu.Unlock()

	wa := c.cs.workerAssignments[workerID]
	wa.Phase = events.PhaseIdle

	if verdict == "DENIED" {
		if ta, ok := c.cs.taskAssignments[wa.TaskID]; ok {
			ta.Status = TaskDenied
		}
	}

	return nil
}

// TestIntegration_MultipleWorkersMultipleTasks tests concurrent task management.
func TestIntegration_MultipleWorkersMultipleTasks(t *testing.T) {
	workerPool := pool.NewWorkerPool(pool.Config{})
	defer workerPool.Close()

	msgIssue := message.New()
	cs := NewCoordinatorServer(claude.NewClient(), workerPool, msgIssue, "/tmp/test", 8765, nil, mocks.NewMockBeadsExecutor(t))

	// Create 4 workers
	for i := 1; i <= 4; i++ {
		workerID := "worker-" + string(rune('0'+i))
		_ = workerPool.AddTestWorker(workerID, pool.WorkerReady)
	}

	// Assign 2 tasks to different workers
	tasks := []string{"perles-abc.1", "perles-abc.2"}

	cs.assignmentsMu.Lock()
	cs.workerAssignments["worker-1"] = &WorkerAssignment{
		TaskID:     tasks[0],
		Role:       RoleImplementer,
		Phase:      events.PhaseImplementing,
		AssignedAt: time.Now(),
	}
	cs.workerAssignments["worker-2"] = &WorkerAssignment{
		TaskID:     tasks[1],
		Role:       RoleImplementer,
		Phase:      events.PhaseImplementing,
		AssignedAt: time.Now(),
	}
	cs.taskAssignments[tasks[0]] = &TaskAssignment{
		TaskID:      tasks[0],
		Implementer: "worker-1",
		Status:      TaskImplementing,
		StartedAt:   time.Now(),
	}
	cs.taskAssignments[tasks[1]] = &TaskAssignment{
		TaskID:      tasks[1],
		Implementer: "worker-2",
		Status:      TaskImplementing,
		StartedAt:   time.Now(),
	}
	cs.assignmentsMu.Unlock()

	// Query state - should show both workers with tasks
	handler := cs.handlers["query_worker_state"]
	result, err := handler(context.Background(), json.RawMessage(`{}`))
	require.NoError(t, err, "query_worker_state failed")

	var response workerStateResponse
	err = json.Unmarshal([]byte(result.Content[0].Text), &response)
	require.NoError(t, err, "Failed to parse response")

	// Count workers in implementing phase
	implementingCount := 0
	for _, w := range response.Workers {
		if w.Phase == "implementing" {
			implementingCount++
		}
	}
	require.Equal(t, 2, implementingCount, "Expected 2 workers implementing")

	// Check task assignments
	require.Len(t, response.TaskAssignments, 2, "Expected 2 task assignments")
}

// TestIntegration_OrphanRecovery tests detection and recovery of orphaned tasks.
func TestIntegration_OrphanRecovery(t *testing.T) {
	workerPool := pool.NewWorkerPool(pool.Config{})
	defer workerPool.Close()

	cs := NewCoordinatorServer(claude.NewClient(), workerPool, nil, "/tmp/test", 8765, nil, mocks.NewMockBeadsExecutor(t))

	taskID := "perles-abc.1"

	// Create worker and assign task
	_ = workerPool.AddTestWorker("worker-1", pool.WorkerWorking)
	cs.assignmentsMu.Lock()
	cs.taskAssignments[taskID] = &TaskAssignment{
		TaskID:      taskID,
		Implementer: "worker-1",
		Status:      TaskImplementing,
	}
	cs.assignmentsMu.Unlock()

	// No orphans yet
	orphans := cs.detectOrphanedTasks()
	require.Empty(t, orphans, "Expected no orphans")

	// Retire the worker (simulating crash/timeout)
	workerPool.GetWorker("worker-1").Retire()

	// Now task should be orphaned
	orphans = cs.detectOrphanedTasks()
	require.Len(t, orphans, 1, "Expected 1 orphan")
	require.Equal(t, taskID, orphans[0], "Expected orphan to be the task")
}

// TestIntegration_ConcurrentToolCalls tests handling of concurrent tool calls.
func TestIntegration_ConcurrentToolCalls(t *testing.T) {
	workerPool := pool.NewWorkerPool(pool.Config{})
	defer workerPool.Close()

	msgIssue := message.New()
	cs := NewCoordinatorServer(claude.NewClient(), workerPool, msgIssue, "/tmp/test", 8765, nil, mocks.NewMockBeadsExecutor(t))

	// Create workers
	for i := 1; i <= 5; i++ {
		workerID := "worker-" + string(rune('0'+i))
		_ = workerPool.AddTestWorker(workerID, pool.WorkerReady)
	}

	var wg sync.WaitGroup
	ctx := context.Background()

	// Concurrent list_workers calls
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			handler := cs.handlers["list_workers"]
			_, _ = handler(ctx, nil)
		}()
	}

	// Concurrent query_worker_state calls
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			handler := cs.handlers["query_worker_state"]
			_, _ = handler(ctx, json.RawMessage(`{}`))
		}()
	}

	// Concurrent state modifications
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(num int) {
			defer wg.Done()
			taskID := "perles-abc." + string(rune('1'+num))
			workerID := "worker-" + string(rune('1'+num))

			cs.assignmentsMu.Lock()
			cs.workerAssignments[workerID] = &WorkerAssignment{
				TaskID:     taskID,
				Role:       RoleImplementer,
				Phase:      events.PhaseImplementing,
				AssignedAt: time.Now(),
			}
			cs.taskAssignments[taskID] = &TaskAssignment{
				TaskID:      taskID,
				Implementer: workerID,
				Status:      TaskImplementing,
				StartedAt:   time.Now(),
			}
			cs.assignmentsMu.Unlock()
		}(i)
	}

	wg.Wait()

	// Verify state is consistent
	cs.assignmentsMu.RLock()
	for workerID, wa := range cs.workerAssignments {
		if wa.TaskID != "" {
			if ta, ok := cs.taskAssignments[wa.TaskID]; ok {
				require.Equal(t, workerID, ta.Implementer,
					"Inconsistent state: worker %s has task %s but task implementer is %s",
					workerID, wa.TaskID, ta.Implementer)
			}
		}
	}
	cs.assignmentsMu.RUnlock()
}

// TestIntegration_MessageFlow tests message flow between workers and coordinator.
func TestIntegration_MessageFlow(t *testing.T) {
	workerPool := pool.NewWorkerPool(pool.Config{})
	defer workerPool.Close()

	msgIssue := message.New()
	_ = NewCoordinatorServer(claude.NewClient(), workerPool, msgIssue, "/tmp/test", 8765, nil, mocks.NewMockBeadsExecutor(t))

	workerID := "worker-1"
	_ = workerPool.AddTestWorker(workerID, pool.WorkerReady)

	// Worker signals ready
	workerStore := newMockMessageStore()
	ws := NewWorkerServer(workerID, workerStore)

	readyHandler := ws.handlers["signal_ready"]
	_, err := readyHandler(context.Background(), json.RawMessage(`{}`))
	require.NoError(t, err, "signal_ready failed")

	// Verify ready message was posted
	require.Len(t, workerStore.appendCalls, 1)
	require.Equal(t, message.MessageWorkerReady, workerStore.appendCalls[0].Type, "Message type mismatch")

	// Worker posts status update
	postHandler := ws.handlers["post_message"]
	_, err = postHandler(context.Background(), json.RawMessage(`{"to": "COORDINATOR", "content": "Task 50% complete"}`))
	require.NoError(t, err, "post_message failed")

	require.Len(t, workerStore.appendCalls, 2)
}

// TestIntegration_ValidateAssignmentConstraints tests that all assignment constraints are enforced.
func TestIntegration_ValidateAssignmentConstraints(t *testing.T) {
	workerPool := pool.NewWorkerPool(pool.Config{})
	defer workerPool.Close()

	cs := NewCoordinatorServer(claude.NewClient(), workerPool, nil, "/tmp/test", 8765, nil, mocks.NewMockBeadsExecutor(t))

	// Create workers
	_ = workerPool.AddTestWorker("worker-1", pool.WorkerReady)
	_ = workerPool.AddTestWorker("worker-2", pool.WorkerWorking) // Not ready
	_ = workerPool.AddTestWorker("worker-3", pool.WorkerRetired) // Retired

	taskID := "perles-abc.1"

	// Test 1: Assigning to ready worker should pass validation
	err := cs.validateTaskAssignment("worker-1", taskID)
	require.NoError(t, err, "Expected no error for ready worker")

	// Test 2: Assigning to working worker should fail
	err = cs.validateTaskAssignment("worker-2", taskID)
	require.Error(t, err, "Expected error for working worker")

	// Test 3: Assigning to non-existent worker should fail
	err = cs.validateTaskAssignment("nonexistent", taskID)
	require.Error(t, err, "Expected error for non-existent worker")

	// Test 4: Assigning when worker already has task should fail
	cs.workerAssignments["worker-1"] = &WorkerAssignment{
		TaskID: "perles-xyz.1",
		Role:   RoleImplementer,
		Phase:  events.PhaseImplementing,
	}
	err = cs.validateTaskAssignment("worker-1", taskID)
	require.Error(t, err, "Expected error when worker already has task")
}

// TestIntegration_StuckWorkerDetection tests stuck worker detection integration.
func TestIntegration_StuckWorkerDetection(t *testing.T) {
	workerPool := pool.NewWorkerPool(pool.Config{})
	defer workerPool.Close()

	cs := NewCoordinatorServer(claude.NewClient(), workerPool, nil, "/tmp/test", 8765, nil, mocks.NewMockBeadsExecutor(t))

	// Create workers with different assignment times
	cs.workerAssignments["worker-1"] = &WorkerAssignment{
		TaskID:     "perles-abc.1",
		AssignedAt: time.Now().Add(-MaxTaskDuration - 5*time.Minute), // Stuck
	}
	cs.workerAssignments["worker-2"] = &WorkerAssignment{
		TaskID:     "perles-abc.2",
		AssignedAt: time.Now().Add(-10 * time.Minute), // Not stuck
	}
	cs.workerAssignments["worker-3"] = &WorkerAssignment{
		TaskID:     "", // Idle
		AssignedAt: time.Now().Add(-MaxTaskDuration - time.Hour),
	}

	stuck := cs.checkStuckWorkers()
	require.Len(t, stuck, 1, "Expected 1 stuck worker")
	require.Equal(t, "worker-1", stuck[0], "Expected stuck worker worker-1")
}

// ============================================================================
// Integration Tests for Reflection Workflow
// ============================================================================
//
// Manual Verification Procedure for Reflection Workflow
// ======================================================
//
// To manually verify the reflection workflow in a real orchestration session:
//
// 1. Start an orchestration session:
//    $ perles orchestrate --task "Test task for reflection verification"
//
// 2. Assign a task to a worker and wait for implementation to complete.
//
// 3. When the coordinator prompts the worker to commit (after review approval),
//    the worker should include reflections in their final response using:
//    - Tool: post_reflections
//    - Arguments:
//      {
//        "task_id": "perles-xxx.N",
//        "summary": "What was accomplished",
//        "insights": "Discoveries during implementation (optional)",
//        "mistakes": "Any errors and lessons (optional)",
//        "learnings": "General learnings (optional)"
//      }
//
// 4. Verify the reflection file was created:
//    $ cat .perles/sessions/<session-id>/workers/<worker-id>/reflection.md
//
// 5. The file should contain:
//    - Header: # Worker Reflection
//    - Metadata: Worker ID, Task ID, Date
//    - Sections: Summary (always), Insights/Mistakes/Learnings (if provided)
//
// 6. After session completion, check all worker reflections:
//    $ find .perles/sessions/<session-id>/workers -name reflection.md
//
// Automated Test Coverage
// -----------------------
// The following tests verify the reflection workflow programmatically:
// - TestReflectionWorkflow_FullCycle: End-to-end MCP tool to file storage
// - TestMultipleWorkersReflecting: Isolation between workers
// - TestReflectionAfterSessionClose: Graceful error handling
// - TestReflectionMarkdownStructure: Correct markdown format
//

// TestReflectionWorkflow_FullCycle tests the complete reflection workflow from
// MCP tool call through session file storage. This verifies that:
// 1. A WorkerServer can be created and wired with a ReflectionWriter (Session)
// 2. The post_reflections tool call writes to the correct session file
// 3. The file content matches the expected markdown format
func TestReflectionWorkflow_FullCycle(t *testing.T) {
	// Import session package for this test
	// We need to use a real Session as ReflectionWriter

	baseDir := t.TempDir()
	sessionID := "test-reflection-workflow"
	sessionDir := baseDir + "/session"

	// Create a real session to act as ReflectionWriter
	sess, err := newTestSession(sessionID, sessionDir)
	require.NoError(t, err, "Failed to create test session")
	defer sess.Close()

	// Create WorkerServer and wire the session as ReflectionWriter
	workerID := "worker-1"
	msgStore := newMockMessageStore()
	ws := NewWorkerServer(workerID, msgStore)
	ws.SetReflectionWriter(sess)

	// Call post_reflections via the MCP tool handler
	handler := ws.handlers["post_reflections"]
	args := `{
		"task_id": "perles-abc.1",
		"summary": "Implemented user validation with comprehensive test coverage.",
		"insights": "Pre-compiled regex patterns are significantly faster for repeated validations.",
		"mistakes": "Initially forgot to handle empty string edge case.",
		"learnings": "Always validate at system boundaries, not deep in the call stack."
	}`

	result, err := handler(context.Background(), json.RawMessage(args))
	require.NoError(t, err, "post_reflections tool call failed")

	// Verify success response
	require.NotNil(t, result)
	require.NotEmpty(t, result.Content)
	responseText := result.Content[0].Text
	require.Contains(t, responseText, `"status"`)
	require.Contains(t, responseText, `"success"`)
	require.Contains(t, responseText, `"file_path"`)

	// Verify file was created in the correct location
	expectedPath := sessionDir + "/workers/worker-1/reflection.md"
	require.Contains(t, responseText, expectedPath, "Response should contain file path")

	// Read and verify file content
	content, err := readTestFile(expectedPath)
	require.NoError(t, err, "Failed to read reflection file")

	// Verify markdown content structure
	require.Contains(t, content, "# Worker Reflection", "Should have header")
	require.Contains(t, content, "**Worker:** worker-1", "Should have worker ID")
	require.Contains(t, content, "**Task:** perles-abc.1", "Should have task ID")
	require.Contains(t, content, "**Date:**", "Should have date")
	require.Contains(t, content, "## Summary", "Should have Summary section")
	require.Contains(t, content, "Implemented user validation", "Should have summary content")
	require.Contains(t, content, "## Insights", "Should have Insights section")
	require.Contains(t, content, "Pre-compiled regex", "Should have insights content")
	require.Contains(t, content, "## Mistakes & Lessons", "Should have Mistakes section")
	require.Contains(t, content, "Initially forgot", "Should have mistakes content")
	require.Contains(t, content, "## Learnings", "Should have Learnings section")
	require.Contains(t, content, "Always validate", "Should have learnings content")

	// Verify message was posted to coordinator
	require.Len(t, msgStore.appendCalls, 1, "Expected 1 message posted to coordinator")
	require.Contains(t, msgStore.appendCalls[0].Content, "Reflection posted for task perles-abc.1")
}

// TestMultipleWorkersReflecting tests that multiple workers in the same session
// can each post reflections without cross-contamination.
func TestMultipleWorkersReflecting(t *testing.T) {
	baseDir := t.TempDir()
	sessionID := "test-multiple-workers-reflecting"
	sessionDir := baseDir + "/session"

	// Create a real session shared by multiple workers
	sess, err := newTestSession(sessionID, sessionDir)
	require.NoError(t, err, "Failed to create test session")
	defer sess.Close()

	// Create multiple WorkerServers, all sharing the same session as ReflectionWriter
	workers := []struct {
		id      string
		taskID  string
		summary string
	}{
		{"worker-1", "perles-abc.1", "Implemented feature X with comprehensive tests and documentation."},
		{"worker-2", "perles-abc.2", "Fixed critical bug in authentication flow and added regression tests."},
		{"worker-3", "perles-abc.3", "Refactored database layer for improved performance and maintainability."},
	}

	for _, w := range workers {
		msgStore := newMockMessageStore()
		ws := NewWorkerServer(w.id, msgStore)
		ws.SetReflectionWriter(sess)

		handler := ws.handlers["post_reflections"]
		args := json.RawMessage(`{
			"task_id": "` + w.taskID + `",
			"summary": "` + w.summary + `"
		}`)

		result, err := handler(context.Background(), args)
		require.NoError(t, err, "post_reflections failed for %s", w.id)
		require.Contains(t, result.Content[0].Text, "success", "Expected success for %s", w.id)
	}

	// Verify each worker has their own reflection.md file
	for _, w := range workers {
		filePath := sessionDir + "/workers/" + w.id + "/reflection.md"
		content, err := readTestFile(filePath)
		require.NoError(t, err, "Failed to read reflection file for %s", w.id)

		// Verify file contains only this worker's content (no cross-contamination)
		require.Contains(t, content, "**Worker:** "+w.id, "Should have correct worker ID")
		require.Contains(t, content, "**Task:** "+w.taskID, "Should have correct task ID")
		require.Contains(t, content, w.summary, "Should have worker's summary")

		// Verify no other worker's content is present
		for _, other := range workers {
			if other.id != w.id {
				require.NotContains(t, content, "**Worker:** "+other.id,
					"Worker %s's file should not contain worker %s's ID", w.id, other.id)
			}
		}
	}
}

// TestReflectionAfterSessionClose tests that attempting to post a reflection
// after the session is closed returns a graceful error (not a panic).
func TestReflectionAfterSessionClose(t *testing.T) {
	baseDir := t.TempDir()
	sessionID := "test-reflection-session-closed"
	sessionDir := baseDir + "/session"

	// Create session
	sess, err := newTestSession(sessionID, sessionDir)
	require.NoError(t, err, "Failed to create test session")

	// Create WorkerServer with session as ReflectionWriter
	workerID := "worker-1"
	msgStore := newMockMessageStore()
	ws := NewWorkerServer(workerID, msgStore)
	ws.SetReflectionWriter(sess)

	// Close the session BEFORE attempting to post reflection
	err = sess.Close()
	require.NoError(t, err, "Failed to close session")

	// Attempt to post reflection - should return error, not panic
	handler := ws.handlers["post_reflections"]
	args := `{
		"task_id": "perles-abc.1",
		"summary": "This should fail gracefully because session is closed."
	}`

	_, err = handler(context.Background(), json.RawMessage(args))
	require.Error(t, err, "Expected error when writing to closed session")
	require.Contains(t, err.Error(), "failed to save reflection", "Error should mention save failure")
}

// TestReflectionMarkdownStructure verifies the exact markdown format generated
// by buildReflectionMarkdown for various field combinations.
func TestReflectionMarkdownStructure(t *testing.T) {
	baseDir := t.TempDir()
	sessionID := "test-reflection-markdown-structure"
	sessionDir := baseDir + "/session"

	sess, err := newTestSession(sessionID, sessionDir)
	require.NoError(t, err, "Failed to create test session")
	defer sess.Close()

	tests := []struct {
		name          string
		args          string
		expectedParts []string
		notExpected   []string
	}{
		{
			name: "all_fields_present",
			args: `{
				"task_id": "perles-test.1",
				"summary": "Full summary with all fields populated.",
				"insights": "Important insight discovered.",
				"mistakes": "Mistake made and lesson learned.",
				"learnings": "General learning for future reference."
			}`,
			expectedParts: []string{
				"# Worker Reflection",
				"**Worker:** worker-test",
				"**Task:** perles-test.1",
				"**Date:**",
				"## Summary",
				"Full summary with all fields",
				"## Insights",
				"Important insight discovered",
				"## Mistakes & Lessons",
				"Mistake made and lesson",
				"## Learnings",
				"General learning for future",
			},
			notExpected: []string{},
		},
		{
			name: "only_required_fields",
			args: `{
				"task_id": "perles-test.2",
				"summary": "Summary only without optional fields populated."
			}`,
			expectedParts: []string{
				"# Worker Reflection",
				"**Worker:** worker-test",
				"**Task:** perles-test.2",
				"## Summary",
				"Summary only without optional",
			},
			notExpected: []string{
				"## Insights",
				"## Mistakes & Lessons",
				"## Learnings",
			},
		},
		{
			name: "partial_optional_fields",
			args: `{
				"task_id": "perles-test.3",
				"summary": "Summary with only insights provided.",
				"insights": "Insight without mistakes or learnings."
			}`,
			expectedParts: []string{
				"## Summary",
				"## Insights",
				"Insight without mistakes",
			},
			notExpected: []string{
				"## Mistakes & Lessons",
				"## Learnings",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msgStore := newMockMessageStore()
			ws := NewWorkerServer("worker-test", msgStore)
			ws.SetReflectionWriter(sess)

			handler := ws.handlers["post_reflections"]
			result, err := handler(context.Background(), json.RawMessage(tt.args))
			require.NoError(t, err, "post_reflections failed")
			require.Contains(t, result.Content[0].Text, "success")

			// Read the file and verify structure
			filePath := sessionDir + "/workers/worker-test/reflection.md"
			content, err := readTestFile(filePath)
			require.NoError(t, err, "Failed to read reflection file")

			for _, expected := range tt.expectedParts {
				require.Contains(t, content, expected, "Content should contain: %s", expected)
			}

			for _, notExpected := range tt.notExpected {
				require.NotContains(t, content, notExpected, "Content should NOT contain: %s", notExpected)
			}
		})
	}
}

// ============================================================================
// Helper types and functions for reflection integration tests
// ============================================================================

// testSession wraps session.Session for testing purposes.
// This provides a minimal implementation of ReflectionWriter interface.
type testSession struct {
	dir    string
	closed bool
	mu     sync.Mutex
}

func newTestSession(id, dir string) (*testSession, error) {
	// Create directory structure
	if err := createTestDir(dir); err != nil {
		return nil, err
	}
	if err := createTestDir(dir + "/workers"); err != nil {
		return nil, err
	}
	return &testSession{dir: dir}, nil
}

func (s *testSession) WriteWorkerReflection(workerID, taskID string, content []byte) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return "", errClosed
	}

	// Create worker directory if needed
	workerPath := s.dir + "/workers/" + workerID
	if err := createTestDir(workerPath); err != nil {
		return "", err
	}

	// Write reflection file
	reflectionPath := workerPath + "/reflection.md"
	if err := writeTestFile(reflectionPath, content); err != nil {
		return "", err
	}

	return reflectionPath, nil
}

func (s *testSession) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	return nil
}

// errClosed is returned when attempting to write to a closed session.
var errClosed = os.ErrClosed

// File operation helpers for tests
func createTestDir(path string) error {
	return os.MkdirAll(path, 0750)
}

func writeTestFile(path string, content []byte) error {
	return os.WriteFile(path, content, 0600)
}

func readTestFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
