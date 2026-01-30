package mention

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	m := New()
	assert.False(t, m.IsActive())
	assert.Empty(t, m.Query())
	assert.Nil(t, m.Selected())
}

func TestModel_SetProcesses(t *testing.T) {
	m := New()
	processes := []Process{
		{ID: "coordinator", Role: "Coordinator"},
		{ID: "worker-1", Role: "Worker"},
		{ID: "worker-2", Role: "Worker"},
	}
	m = m.SetProcesses(processes)
	assert.Len(t, m.processes, 3)
}

func TestModel_Activate(t *testing.T) {
	m := New()
	processes := []Process{
		{ID: "coordinator", Role: "Coordinator"},
		{ID: "worker-1", Role: "Worker"},
	}
	m = m.SetProcesses(processes)
	m = m.Activate(5)

	assert.True(t, m.IsActive())
	assert.Equal(t, "", m.Query())
	assert.Equal(t, 5, m.queryStart)
	assert.Len(t, m.filtered, 2)
}

func TestModel_Deactivate(t *testing.T) {
	m := New()
	m = m.Activate(0)
	m = m.Deactivate()

	assert.False(t, m.IsActive())
	assert.Empty(t, m.Query())
}

func TestModel_UpdateQuery_Filters(t *testing.T) {
	m := New()
	processes := []Process{
		{ID: "coordinator", Role: "Coordinator"},
		{ID: "worker-1", Role: "Worker"},
		{ID: "worker-2", Role: "Worker"},
	}
	m = m.SetProcesses(processes)
	m = m.Activate(0)

	// Filter by "work"
	m, hasMatches := m.UpdateQuery("work")
	assert.True(t, hasMatches)
	assert.Len(t, m.filtered, 2) // worker-1 and worker-2

	// Filter by "1"
	m, hasMatches = m.UpdateQuery("1")
	assert.True(t, hasMatches)
	assert.Len(t, m.filtered, 1) // worker-1

	// Filter with no matches
	m, hasMatches = m.UpdateQuery("xyz")
	assert.False(t, hasMatches)
	assert.Len(t, m.filtered, 0)
}

func TestModel_Next_Prev(t *testing.T) {
	m := New()
	processes := []Process{
		{ID: "coordinator", Role: "Coordinator"},
		{ID: "worker-1", Role: "Worker"},
		{ID: "worker-2", Role: "Worker"},
	}
	m = m.SetProcesses(processes)
	m = m.Activate(0)

	assert.Equal(t, 0, m.cursor)

	m = m.Next()
	assert.Equal(t, 1, m.cursor)

	m = m.Next()
	assert.Equal(t, 2, m.cursor)

	m = m.Next() // Wraps around
	assert.Equal(t, 0, m.cursor)

	m = m.Prev() // Wraps around to end
	assert.Equal(t, 2, m.cursor)

	m = m.Prev()
	assert.Equal(t, 1, m.cursor)
}

func TestModel_Selected(t *testing.T) {
	m := New()
	processes := []Process{
		{ID: "coordinator", Role: "Coordinator"},
		{ID: "worker-1", Role: "Worker"},
	}
	m = m.SetProcesses(processes)
	m = m.Activate(0)

	selected := m.Selected()
	require.NotNil(t, selected)
	assert.Equal(t, "coordinator", selected.ID)

	m = m.Next()
	selected = m.Selected()
	require.NotNil(t, selected)
	assert.Equal(t, "worker-1", selected.ID)
}

func TestModel_HandleKey_CtrlN(t *testing.T) {
	m := New()
	processes := []Process{
		{ID: "coordinator", Role: "Coordinator"},
		{ID: "worker-1", Role: "Worker"},
	}
	m = m.SetProcesses(processes)
	m = m.Activate(0)

	// Use down arrow which is also handled
	m, consumed, selected := m.HandleKey(tea.KeyMsg{Type: tea.KeyDown})

	assert.True(t, consumed)
	assert.Nil(t, selected)
	assert.Equal(t, 1, m.cursor)
}

func TestModel_HandleKey_CtrlP(t *testing.T) {
	m := New()
	processes := []Process{
		{ID: "coordinator", Role: "Coordinator"},
		{ID: "worker-1", Role: "Worker"},
	}
	m = m.SetProcesses(processes)
	m = m.Activate(0)
	m = m.Next() // Move to worker-1

	// Use up arrow which is also handled
	m, consumed, selected := m.HandleKey(tea.KeyMsg{Type: tea.KeyUp})

	assert.True(t, consumed)
	assert.Nil(t, selected)
	assert.Equal(t, 0, m.cursor)
}

func TestModel_HandleKey_Enter(t *testing.T) {
	m := New()
	processes := []Process{
		{ID: "coordinator", Role: "Coordinator"},
		{ID: "worker-1", Role: "Worker"},
	}
	m = m.SetProcesses(processes)
	m = m.Activate(0)

	m, consumed, selected := m.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})

	assert.True(t, consumed)
	require.NotNil(t, selected)
	assert.Equal(t, "coordinator", selected.ID)
	assert.False(t, m.IsActive()) // Should be deactivated after selection
}

func TestModel_HandleKey_Escape(t *testing.T) {
	m := New()
	m = m.Activate(0)

	m, consumed, selected := m.HandleKey(tea.KeyMsg{Type: tea.KeyEscape})

	assert.True(t, consumed)
	assert.Nil(t, selected)
	assert.False(t, m.IsActive())
}

func TestModel_HandleKey_NotActive(t *testing.T) {
	m := New()

	m, consumed, selected := m.HandleKey(tea.KeyMsg{Type: tea.KeyCtrlN})

	assert.False(t, consumed)
	assert.Nil(t, selected)
}

func TestModel_HandleKey_Tab_NotConsumed(t *testing.T) {
	m := New()
	m = m.Activate(0)

	m, consumed, selected := m.HandleKey(tea.KeyMsg{Type: tea.KeyTab})

	assert.False(t, consumed) // Tab should be passed through for channel cycling
	assert.Nil(t, selected)
}

func TestModel_View_Empty(t *testing.T) {
	m := New()
	view := m.View(50)
	assert.Empty(t, view) // Not active, no view
}

func TestModel_View_Active(t *testing.T) {
	m := New()
	processes := []Process{
		{ID: "coordinator", Role: "Coordinator"},
		{ID: "worker-1", Role: "Worker"},
	}
	m = m.SetProcesses(processes)
	m = m.Activate(0)

	view := m.View(50)
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "@coordinator")
	assert.Contains(t, view, "@worker-1")
}

func TestModel_View_NoMatches(t *testing.T) {
	m := New()
	processes := []Process{
		{ID: "coordinator", Role: "Coordinator"},
	}
	m = m.SetProcesses(processes)
	m = m.Activate(0)
	m, _ = m.UpdateQuery("xyz") // No matches

	view := m.View(50)
	assert.Empty(t, view) // No matches, no view
}
