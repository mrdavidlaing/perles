package dashboard

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zjrosen/perles/internal/orchestration/controlplane"
	"github.com/zjrosen/perles/internal/orchestration/events"
	"github.com/zjrosen/perles/internal/orchestration/metrics"
	"github.com/zjrosen/perles/internal/ui/shared/chatrender"
)

func TestNewCoordinatorPanel(t *testing.T) {
	panel := NewCoordinatorPanel(false, false)

	require.NotNil(t, panel)
	require.False(t, panel.IsFocused(), "panel should be unfocused by default")
	require.Empty(t, panel.coordinatorMessages)
	require.True(t, panel.coordinatorDirty)
	require.Equal(t, TabCoordinator, panel.activeTab)
}

func TestCoordinatorPanel_SetWorkflow(t *testing.T) {
	panel := NewCoordinatorPanel(false, false)

	state := &WorkflowUIState{
		CoordinatorMessages: []chatrender.Message{
			{Role: "user", Content: "Hello"},
			{Role: "assistant", Content: "Hi there"},
		},
		CoordinatorStatus:     events.ProcessStatusWorking,
		CoordinatorQueueCount: 1,
	}

	panel.SetWorkflow("wf-123", state)

	require.Equal(t, controlplane.WorkflowID("wf-123"), panel.workflowID)
	require.Len(t, panel.coordinatorMessages, 2)
	require.Equal(t, events.ProcessStatusWorking, panel.coordinatorStatus)
	require.Equal(t, 1, panel.coordinatorQueue)
	require.True(t, panel.coordinatorDirty, "should be dirty after setting workflow")
}

func TestCoordinatorPanel_SetWorkflow_SameWorkflowNewMessages(t *testing.T) {
	panel := NewCoordinatorPanel(false, false)

	// Set initial state
	state := &WorkflowUIState{
		CoordinatorMessages: []chatrender.Message{
			{Role: "user", Content: "Hello"},
		},
		CoordinatorStatus: events.ProcessStatusReady,
	}
	panel.SetWorkflow("wf-123", state)
	panel.coordinatorDirty = false // simulate View() was called

	// Add more messages
	state.CoordinatorMessages = append(state.CoordinatorMessages, chatrender.Message{Role: "assistant", Content: "Hi"})
	state.CoordinatorStatus = events.ProcessStatusWorking
	panel.SetWorkflow("wf-123", state)

	require.Len(t, panel.coordinatorMessages, 2)
	require.Equal(t, events.ProcessStatusWorking, panel.coordinatorStatus)
	require.True(t, panel.coordinatorDirty, "should be dirty when message count changes")
}

func TestCoordinatorPanel_Focus(t *testing.T) {
	panel := NewCoordinatorPanel(false, false)
	panel.Blur()

	require.False(t, panel.IsFocused())

	panel.Focus()

	require.True(t, panel.IsFocused())
}

func TestCoordinatorPanel_SetSize(t *testing.T) {
	panel := NewCoordinatorPanel(false, false)

	panel.SetSize(100, 50)

	require.Equal(t, 100, panel.width)
	require.Equal(t, 50, panel.height)
}

func TestCoordinatorPanel_View_EmptyMessages(t *testing.T) {
	panel := NewCoordinatorPanel(false, false)
	panel.SetSize(80, 20)
	panel.SetWorkflow("wf-123", nil)

	view := panel.View()

	require.NotEmpty(t, view)
	require.Contains(t, view, "Coord", "should show Coordinator tab label")
	require.Contains(t, view, "Msgs", "should show Messages tab label")
}

func TestRenderChatContent_EmptyMessages(t *testing.T) {
	cfg := chatrender.RenderConfig{
		AgentLabel: "Coordinator",
		AgentColor: chatrender.CoordinatorColor,
		UserLabel:  "User",
	}
	content := renderChatContent(nil, 80, cfg)

	require.Contains(t, content, "Waiting for the coordinator to initialize.")
}

func TestRenderChatContent_WithMessages(t *testing.T) {
	messages := []chatrender.Message{
		{Role: "user", Content: "Hello world"},
		{Role: "assistant", Content: "Hi there!"},
	}

	cfg := chatrender.RenderConfig{
		AgentLabel: "Coordinator",
		AgentColor: chatrender.CoordinatorColor,
		UserLabel:  "User",
	}
	content := renderChatContent(messages, 80, cfg)

	require.Contains(t, content, "User")
	require.Contains(t, content, "Hello world")
	require.Contains(t, content, "Coordinator") // Uses "Coordinator" label from RenderConfig
	require.Contains(t, content, "Hi there!")
}

func TestRenderChatContent_ToolCall(t *testing.T) {
	messages := []chatrender.Message{
		{Role: "assistant", Content: "Using a tool", IsToolCall: true},
	}

	cfg := chatrender.RenderConfig{
		AgentLabel: "Coordinator",
		AgentColor: chatrender.CoordinatorColor,
		UserLabel:  "User",
	}
	content := renderChatContent(messages, 80, cfg)

	// Tool calls use the "╰╴" prefix in shared chatrender
	require.Contains(t, content, "╰╴")
	require.Contains(t, content, "Using a tool")
}

func TestRenderChatContent_FiltersEmptyMessages(t *testing.T) {
	messages := []chatrender.Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: ""},    // Empty - should be filtered
		{Role: "assistant", Content: "Hi!"}, // Non-empty - should appear
	}

	cfg := chatrender.RenderConfig{
		AgentLabel: "Coordinator",
		AgentColor: chatrender.CoordinatorColor,
		UserLabel:  "User",
	}
	content := renderChatContent(messages, 80, cfg)

	require.Contains(t, content, "Hello")
	require.Contains(t, content, "Hi!")
	// Should not have empty lines from the filtered message
}

func TestNewCoordinatorPanel_InputStartsUnfocused(t *testing.T) {
	panel := NewCoordinatorPanel(false, false)

	// Verify the input starts unfocused (focus is given on explicit Focus() call)
	require.False(t, panel.input.Focused())
	require.False(t, panel.focused)

	// After Focus(), both should be true
	panel.Focus()
	require.True(t, panel.input.Focused())
	require.True(t, panel.focused)
}

func TestCoordinatorPanel_TabNavigation(t *testing.T) {
	panel := NewCoordinatorPanel(false, false)

	// Initially on TabCoordinator
	require.Equal(t, TabCoordinator, panel.ActiveTab())

	// Tab forward
	panel.NextTab()
	require.Equal(t, TabMessages, panel.ActiveTab())

	// Tab backward
	panel.PrevTab()
	require.Equal(t, TabCoordinator, panel.ActiveTab())

	// Tab wraps around
	panel.PrevTab()
	require.Equal(t, TabMessages, panel.ActiveTab(), "should wrap to last tab")
}

func TestCoordinatorPanel_TabNavigationDebugMode(t *testing.T) {
	panel := NewCoordinatorPanel(true, false) // debug mode, no vim

	// Initially on TabCoordinator
	require.Equal(t, TabCoordinator, panel.ActiveTab())

	// Tab forward through: Coordinator -> Messages -> CmdLog
	panel.NextTab()
	require.Equal(t, TabMessages, panel.ActiveTab())
	panel.NextTab()
	require.Equal(t, 2, panel.ActiveTab(), "should be on command log tab")

	// Tab wraps back to coordinator
	panel.NextTab()
	require.Equal(t, TabCoordinator, panel.ActiveTab())

	// Tab backward from coordinator wraps to command log
	panel.PrevTab()
	require.Equal(t, 2, panel.ActiveTab(), "should wrap to command log tab")
}

func TestCoordinatorPanel_TabNavigationWithWorkers(t *testing.T) {
	panel := NewCoordinatorPanel(false, false)

	// Set workflow with workers
	state := &WorkflowUIState{
		WorkerIDs:         []string{"worker-1", "worker-2"},
		WorkerStatus:      make(map[string]events.ProcessStatus),
		WorkerPhases:      make(map[string]events.ProcessPhase),
		WorkerMessages:    make(map[string][]chatrender.Message),
		WorkerQueueCounts: make(map[string]int),
	}
	panel.SetWorkflow("wf-123", state)

	// Should now have 4 tabs: Coord, Msgs, W1, W2
	require.Equal(t, 4, panel.tabCount())

	// Navigate through all tabs
	require.Equal(t, TabCoordinator, panel.ActiveTab())
	panel.NextTab()
	require.Equal(t, TabMessages, panel.ActiveTab())
	panel.NextTab()
	require.Equal(t, TabFirstWorker, panel.ActiveTab()) // worker-1
	panel.NextTab()
	require.Equal(t, TabFirstWorker+1, panel.ActiveTab()) // worker-2
	panel.NextTab()
	require.Equal(t, TabCoordinator, panel.ActiveTab(), "should wrap back to coordinator")
}

func TestCoordinatorPanel_SetWorkflow_SyncsWorkerData(t *testing.T) {
	panel := NewCoordinatorPanel(false, false)

	state := &WorkflowUIState{
		WorkerIDs: []string{"worker-1", "worker-2"},
		WorkerStatus: map[string]events.ProcessStatus{
			"worker-1": events.ProcessStatusWorking,
			"worker-2": events.ProcessStatusReady,
		},
		WorkerPhases: map[string]events.ProcessPhase{
			"worker-1": events.ProcessPhaseImplementing,
		},
		WorkerMessages: map[string][]chatrender.Message{
			"worker-1": {{Role: "assistant", Content: "Hello from worker"}},
		},
		WorkerQueueCounts: map[string]int{
			"worker-1": 2,
		},
	}
	panel.SetWorkflow("wf-123", state)

	require.Len(t, panel.workerIDs, 2)
	require.Equal(t, events.ProcessStatusWorking, panel.workerStatus["worker-1"])
	require.Equal(t, events.ProcessPhaseImplementing, panel.workerPhases["worker-1"])
	require.Len(t, panel.workerMessages["worker-1"], 1)
	require.Equal(t, 2, panel.workerQueues["worker-1"])
}

func TestCoordinatorPanel_SetWorkflow_ResetsTabWhenWorkerRemoved(t *testing.T) {
	panel := NewCoordinatorPanel(false, false)

	// Initial state with workers
	state := &WorkflowUIState{
		WorkerIDs:         []string{"worker-1", "worker-2"},
		WorkerStatus:      make(map[string]events.ProcessStatus),
		WorkerPhases:      make(map[string]events.ProcessPhase),
		WorkerMessages:    make(map[string][]chatrender.Message),
		WorkerQueueCounts: make(map[string]int),
	}
	panel.SetWorkflow("wf-123", state)

	// Navigate to worker-2 tab
	panel.activeTab = TabFirstWorker + 1 // worker-2

	// Remove workers
	state.WorkerIDs = nil
	panel.SetWorkflow("wf-123", state)

	// Should reset to coordinator since worker tab no longer exists
	require.Equal(t, TabCoordinator, panel.activeTab)
}

func TestCoordinatorPanel_FormatWorkerTabLabel(t *testing.T) {
	panel := NewCoordinatorPanel(false, false)

	require.Equal(t, "W1", panel.formatWorkerTabLabel("worker-1"))
	require.Equal(t, "W99", panel.formatWorkerTabLabel("worker-99"))
	require.Equal(t, "custom", panel.formatWorkerTabLabel("custom"))
	require.Equal(t, "longla", panel.formatWorkerTabLabel("longlabel")) // truncates to 6 chars
}

func TestSetWorkflow_SyncsMetrics(t *testing.T) {
	panel := NewCoordinatorPanel(false, false)

	coordinatorMetrics := &metrics.TokenMetrics{
		TokensUsed:  27000,
		TotalTokens: 200000,
	}
	workerMetrics := map[string]*metrics.TokenMetrics{
		"worker-1": {TokensUsed: 15000, TotalTokens: 200000},
		"worker-2": {TokensUsed: 8000, TotalTokens: 200000},
	}

	state := &WorkflowUIState{
		CoordinatorMetrics: coordinatorMetrics,
		WorkerIDs:          []string{"worker-1", "worker-2"},
		WorkerStatus:       make(map[string]events.ProcessStatus),
		WorkerPhases:       make(map[string]events.ProcessPhase),
		WorkerMessages:     make(map[string][]chatrender.Message),
		WorkerMetrics:      workerMetrics,
		WorkerQueueCounts:  make(map[string]int),
	}

	panel.SetWorkflow("wf-123", state)

	// Verify coordinator metrics synced
	require.Equal(t, coordinatorMetrics, panel.coordinatorMetrics)
	require.Equal(t, 27000, panel.coordinatorMetrics.TokensUsed)
	require.Equal(t, 200000, panel.coordinatorMetrics.TotalTokens)

	// Verify worker metrics synced
	require.Len(t, panel.workerMetrics, 2)
	require.Equal(t, 15000, panel.workerMetrics["worker-1"].TokensUsed)
	require.Equal(t, 8000, panel.workerMetrics["worker-2"].TokensUsed)
}

func TestSetWorkflow_ClearsStaleMetrics(t *testing.T) {
	panel := NewCoordinatorPanel(false, false)

	// First workflow with worker-1 and worker-2
	state1 := &WorkflowUIState{
		CoordinatorMetrics: &metrics.TokenMetrics{TokensUsed: 10000, TotalTokens: 200000},
		WorkerIDs:          []string{"worker-1", "worker-2"},
		WorkerStatus:       make(map[string]events.ProcessStatus),
		WorkerPhases:       make(map[string]events.ProcessPhase),
		WorkerMessages:     make(map[string][]chatrender.Message),
		WorkerMetrics: map[string]*metrics.TokenMetrics{
			"worker-1": {TokensUsed: 5000, TotalTokens: 200000},
			"worker-2": {TokensUsed: 3000, TotalTokens: 200000},
		},
		WorkerQueueCounts: make(map[string]int),
	}
	panel.SetWorkflow("wf-1", state1)

	// Verify first workflow metrics
	require.Len(t, panel.workerMetrics, 2)
	require.NotNil(t, panel.workerMetrics["worker-1"])
	require.NotNil(t, panel.workerMetrics["worker-2"])

	// Second workflow with only worker-3 (different set of workers)
	state2 := &WorkflowUIState{
		CoordinatorMetrics: &metrics.TokenMetrics{TokensUsed: 20000, TotalTokens: 200000},
		WorkerIDs:          []string{"worker-3"},
		WorkerStatus:       make(map[string]events.ProcessStatus),
		WorkerPhases:       make(map[string]events.ProcessPhase),
		WorkerMessages:     make(map[string][]chatrender.Message),
		WorkerMetrics: map[string]*metrics.TokenMetrics{
			"worker-3": {TokensUsed: 7000, TotalTokens: 200000},
		},
		WorkerQueueCounts: make(map[string]int),
	}
	panel.SetWorkflow("wf-2", state2)

	// Verify old workers' metrics are cleared and new worker metrics are set
	require.Len(t, panel.workerMetrics, 1, "should only have 1 worker metrics after switching workflows")
	require.Nil(t, panel.workerMetrics["worker-1"], "worker-1 metrics should be cleared")
	require.Nil(t, panel.workerMetrics["worker-2"], "worker-2 metrics should be cleared")
	require.NotNil(t, panel.workerMetrics["worker-3"], "worker-3 metrics should be set")
	require.Equal(t, 7000, panel.workerMetrics["worker-3"].TokensUsed)

	// Verify coordinator metrics updated
	require.Equal(t, 20000, panel.coordinatorMetrics.TokensUsed)
}

func TestGetActiveMetricsDisplay_Coordinator(t *testing.T) {
	panel := NewCoordinatorPanel(false, false)

	// Set up coordinator with metrics
	state := &WorkflowUIState{
		CoordinatorMetrics: &metrics.TokenMetrics{
			TokensUsed:  27000,
			TotalTokens: 200000,
		},
	}
	panel.SetWorkflow("wf-123", state)
	panel.activeTab = TabCoordinator

	result := panel.getActiveMetricsDisplay()

	// FormatMetricsDisplay returns formatted string like "27k/200k"
	require.NotEmpty(t, result)
	require.Contains(t, result, "27k")
	require.Contains(t, result, "200k")
}

func TestGetActiveMetricsDisplay_Worker(t *testing.T) {
	panel := NewCoordinatorPanel(false, false)

	// Set up with workers and metrics
	state := &WorkflowUIState{
		WorkerIDs:    []string{"worker-1", "worker-2"},
		WorkerStatus: make(map[string]events.ProcessStatus),
		WorkerPhases: make(map[string]events.ProcessPhase),
		WorkerMetrics: map[string]*metrics.TokenMetrics{
			"worker-1": {TokensUsed: 15000, TotalTokens: 200000},
			"worker-2": {TokensUsed: 8000, TotalTokens: 200000},
		},
		WorkerMessages:    make(map[string][]chatrender.Message),
		WorkerQueueCounts: make(map[string]int),
	}
	panel.SetWorkflow("wf-123", state)

	// Select worker-1 tab (TabFirstWorker + 0)
	panel.activeTab = TabFirstWorker

	result := panel.getActiveMetricsDisplay()

	// Should show worker-1's metrics (15k/200k)
	require.NotEmpty(t, result)
	require.Contains(t, result, "15k")
	require.Contains(t, result, "200k")

	// Select worker-2 tab (TabFirstWorker + 1)
	panel.activeTab = TabFirstWorker + 1

	result = panel.getActiveMetricsDisplay()

	// Should show worker-2's metrics (8k/200k)
	require.NotEmpty(t, result)
	require.Contains(t, result, "8k")
	require.Contains(t, result, "200k")
}

func TestGetActiveMetricsDisplay_Messages(t *testing.T) {
	panel := NewCoordinatorPanel(false, false)

	// Set up with coordinator metrics
	state := &WorkflowUIState{
		CoordinatorMetrics: &metrics.TokenMetrics{
			TokensUsed:  27000,
			TotalTokens: 200000,
		},
	}
	panel.SetWorkflow("wf-123", state)

	// Select messages tab
	panel.activeTab = TabMessages

	result := panel.getActiveMetricsDisplay()

	// Should return empty string for message log tab
	require.Empty(t, result)
}

func TestGetActiveMetricsDisplay_NilMetrics(t *testing.T) {
	panel := NewCoordinatorPanel(false, false)

	// Set up without any metrics (nil)
	state := &WorkflowUIState{
		CoordinatorMetrics: nil,
		WorkerIDs:          []string{"worker-1"},
		WorkerStatus:       make(map[string]events.ProcessStatus),
		WorkerPhases:       make(map[string]events.ProcessPhase),
		WorkerMessages:     make(map[string][]chatrender.Message),
		WorkerMetrics:      nil, // nil map
		WorkerQueueCounts:  make(map[string]int),
	}
	panel.SetWorkflow("wf-123", state)

	// Coordinator tab with nil metrics
	panel.activeTab = TabCoordinator
	result := panel.getActiveMetricsDisplay()
	require.Empty(t, result, "should return empty string for nil coordinator metrics")

	// Worker tab with nil metrics map
	panel.activeTab = TabFirstWorker
	result = panel.getActiveMetricsDisplay()
	require.Empty(t, result, "should return empty string for nil worker metrics")
}

func TestGetActiveMetricsDisplay_InvalidWorkerTab(t *testing.T) {
	panel := NewCoordinatorPanel(false, false)

	// Set up with only one worker
	state := &WorkflowUIState{
		WorkerIDs:    []string{"worker-1"},
		WorkerStatus: make(map[string]events.ProcessStatus),
		WorkerPhases: make(map[string]events.ProcessPhase),
		WorkerMetrics: map[string]*metrics.TokenMetrics{
			"worker-1": {TokensUsed: 15000, TotalTokens: 200000},
		},
		WorkerMessages:    make(map[string][]chatrender.Message),
		WorkerQueueCounts: make(map[string]int),
	}
	panel.SetWorkflow("wf-123", state)

	// Try to access worker tab index that doesn't exist (worker-2 at index 1)
	panel.activeTab = TabFirstWorker + 5 // Invalid index

	result := panel.getActiveMetricsDisplay()

	// Should return empty string for invalid worker index (no panic)
	require.Empty(t, result)
}

func TestView_ShowsMetricsInBottomRight(t *testing.T) {
	panel := NewCoordinatorPanel(false, false)
	panel.SetSize(60, 20)

	// Set up coordinator with metrics
	state := &WorkflowUIState{
		CoordinatorMetrics: &metrics.TokenMetrics{
			TokensUsed:  27000,
			TotalTokens: 200000,
		},
		CoordinatorStatus: events.ProcessStatusWorking,
	}
	panel.SetWorkflow("wf-123", state)
	panel.activeTab = TabCoordinator

	view := panel.View()

	// Verify the metrics string appears in the rendered output
	// FormatMetricsDisplay returns "27k/200k" for these values
	require.Contains(t, view, "27k/200k", "metrics should appear in View() output")
}

func TestView_MetricsFitInPanelWidth(t *testing.T) {
	panel := NewCoordinatorPanel(false, false)

	// Use exactly 60-char width as specified in task
	panel.SetSize(60, 20)

	// Set up with both queue count (BottomLeft) and metrics (BottomRight)
	state := &WorkflowUIState{
		CoordinatorMetrics: &metrics.TokenMetrics{
			TokensUsed:  27000,
			TotalTokens: 200000,
		},
		CoordinatorStatus:     events.ProcessStatusWorking,
		CoordinatorQueueCount: 3, // Will show "[3 queued]" in BottomLeft
	}
	panel.SetWorkflow("wf-123", state)
	panel.activeTab = TabCoordinator

	view := panel.View()

	// Verify both queue count and metrics appear without truncation
	// FormatQueueCount returns "[N queued]" format
	require.Contains(t, view, "[3 queued]", "queue count should appear in BottomLeft")
	require.Contains(t, view, "27k/200k", "metrics should appear in BottomRight")

	// Verify no line exceeds panel width (basic overflow check)
	lines := splitLines(view)
	for _, line := range lines {
		// Use visual width (ANSI codes don't count toward visual width)
		// We just verify no obvious overflow - actual visual rendering is what matters
		require.LessOrEqual(t, visualWidth(line), 60,
			"no line should exceed panel width of 60 characters")
	}
}

// splitLines splits a string into lines
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

// visualWidth calculates the visual width of a string, ignoring ANSI escape codes
func visualWidth(s string) int {
	width := 0
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		width++
	}
	return width
}
