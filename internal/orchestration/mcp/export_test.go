package mcp

import (
	"github.com/zjrosen/perles/internal/orchestration/message"
	"github.com/zjrosen/perles/internal/orchestration/pool"
)

// GetPool returns the worker pool for test access.
func (cs *CoordinatorServer) GetPool() *pool.WorkerPool {
	return cs.pool
}

// GetMessageIssue returns the message issue for test access.
func (cs *CoordinatorServer) GetMessageIssue() *message.Issue {
	return cs.msgIssue
}

// IsValidTaskID exposes the validation function for testing.
func IsValidTaskID(taskID string) bool {
	return isValidTaskID(taskID)
}

// ValidateTaskAssignment exposes the validation function for testing.
func (cs *CoordinatorServer) ValidateTaskAssignment(workerID, taskID string) error {
	return cs.validateTaskAssignment(workerID, taskID)
}

// ValidateReviewAssignment exposes the validation function for testing.
func (cs *CoordinatorServer) ValidateReviewAssignment(reviewerID, taskID, implementerID string) error {
	return cs.validateReviewAssignment(reviewerID, taskID, implementerID)
}

// DetectOrphanedTasks exposes the detection function for testing.
func (cs *CoordinatorServer) DetectOrphanedTasks() []string {
	return cs.detectOrphanedTasks()
}

// CheckStuckWorkers exposes the detection function for testing.
func (cs *CoordinatorServer) CheckStuckWorkers() []string {
	return cs.checkStuckWorkers()
}

// SetWorkerAssignment allows tests to set worker assignments directly.
func (cs *CoordinatorServer) SetWorkerAssignment(workerID string, assignment *WorkerAssignment) {
	cs.assignmentsMu.Lock()
	defer cs.assignmentsMu.Unlock()
	cs.workerAssignments[workerID] = assignment
}

// SetTaskAssignment allows tests to set task assignments directly.
func (cs *CoordinatorServer) SetTaskAssignment(taskID string, assignment *TaskAssignment) {
	cs.assignmentsMu.Lock()
	defer cs.assignmentsMu.Unlock()
	cs.taskAssignments[taskID] = assignment
}
