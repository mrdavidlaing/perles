package registry

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegistration_Getters(t *testing.T) {
	chain, err := NewChain().
		Node("research", "Research", "v1-research.md").
		Node("propose", "Propose", "v1-proposal.md").
		Build()
	require.NoError(t, err)

	reg := newRegistration("workflow", "planning-standard", "v1", "Standard Planning Workflow", "Three-phase workflow: Research, Propose, Plan", "", "", "", chain, nil, nil, SourceBuiltIn)

	require.Equal(t, "workflow", reg.Namespace())
	require.Equal(t, "planning-standard", reg.Key())
	require.Equal(t, "v1", reg.Version())
	require.Equal(t, "Standard Planning Workflow", reg.Name())
	require.Equal(t, "Three-phase workflow: Research, Propose, Plan", reg.Description())
	require.Len(t, reg.DAG().Nodes(), 2)
}

func TestRegistration_EmptyFields(t *testing.T) {
	chain, err := NewChain().
		Node("plan", "Plan", "v1-plan.md").
		Build()
	require.NoError(t, err)

	// Registration allows empty name/description - validation is in builder
	reg := newRegistration("workflow", "simple", "v1", "", "", "", "", "", chain, nil, nil, SourceBuiltIn)

	require.Equal(t, "workflow", reg.Namespace())
	require.Equal(t, "simple", reg.Key())
	require.Equal(t, "v1", reg.Version())
	require.Equal(t, "", reg.Name())
	require.Equal(t, "", reg.Description())
	require.Len(t, reg.DAG().Nodes(), 1)
}

func TestRegistration_EpicTemplate(t *testing.T) {
	chain, err := NewChain().
		Node("plan", "Plan", "v1-plan.md").
		Build()
	require.NoError(t, err)

	reg := newRegistration("workflow", "test", "v1", "Test", "Desc", "v1-epic-template.md", "", "", chain, nil, nil, SourceBuiltIn)
	require.Equal(t, "v1-epic-template.md", reg.EpicTemplate())

	// Empty initial prompt
	regNoPrompt := newRegistration("workflow", "test2", "v1", "Test", "Desc", "", "", "", chain, nil, nil, SourceBuiltIn)
	require.Equal(t, "", regNoPrompt.EpicTemplate())
}

func TestRegistration_SystemPrompt_ReturnsValue(t *testing.T) {
	chain, err := NewChain().
		Node("plan", "Plan", "v1-plan.md").
		Build()
	require.NoError(t, err)

	reg := newRegistration("workflow", "test", "v1", "Test", "Desc", "", "epic_driven.md", "", chain, nil, nil, SourceBuiltIn)
	require.Equal(t, "epic_driven.md", reg.SystemPrompt())

	// Empty system prompt is allowed at domain level
	regNoPrompt := newRegistration("workflow", "test2", "v1", "Test", "Desc", "", "", "", chain, nil, nil, SourceBuiltIn)
	require.Equal(t, "", regNoPrompt.SystemPrompt())
}

func TestRegistration_DAGAccess(t *testing.T) {
	chain, err := NewChain().
		Node("research", "Research", "v1-research.md").
		Node("propose", "Propose", "v1-proposal.md").
		Node("plan", "Plan", "v1-plan.md").
		Build()
	require.NoError(t, err)

	reg := newRegistration("workflow", "planning-standard", "v1", "Standard", "Description", "", "", "", chain, nil, nil, SourceBuiltIn)

	nodes := reg.DAG().Nodes()
	require.Len(t, nodes, 3)
	require.Equal(t, "research", nodes[0].Key())
	require.Equal(t, "Research", nodes[0].Name())
	require.Equal(t, "v1-research.md", nodes[0].Template())
	require.Equal(t, "propose", nodes[1].Key())
	require.Equal(t, "plan", nodes[2].Key())
}

// Source tests

func TestSource_String(t *testing.T) {
	tests := []struct {
		source   Source
		expected string
	}{
		{SourceBuiltIn, "built-in"},
		{SourceCommunity, "community"},
		{SourceUser, "user"},
		{Source(99), "unknown"}, // Test unknown value
	}

	for _, tc := range tests {
		require.Equal(t, tc.expected, tc.source.String())
	}
}

func TestRegistration_Source(t *testing.T) {
	chain, err := NewChain().
		Node("plan", "Plan", "v1-plan.md").
		Build()
	require.NoError(t, err)

	// Test SourceBuiltIn
	regBuiltIn := newRegistration("workflow", "test", "v1", "Test", "Desc", "", "", "", chain, nil, nil, SourceBuiltIn)
	require.Equal(t, SourceBuiltIn, regBuiltIn.Source())

	// Test SourceUser
	regUser := newRegistration("workflow", "test2", "v1", "Test", "Desc", "", "", "", chain, nil, nil, SourceUser)
	require.Equal(t, SourceUser, regUser.Source())
}

func TestSource_DefaultValue(t *testing.T) {
	// Zero value of Source should be SourceBuiltIn (iota starts at 0)
	var s Source
	require.Equal(t, SourceBuiltIn, s)
	require.Equal(t, "built-in", s.String())
}

func TestRegistration_IsEpicDriven(t *testing.T) {
	epicIDArg, err := NewArgument("epic_id", "Epic ID", "The epic to work on", ArgumentTypeText, true, "")
	require.NoError(t, err)

	goalArg, err := NewArgument("goal", "Goal", "What to achieve", ArgumentTypeTextarea, true, "")
	require.NoError(t, err)

	tests := []struct {
		name     string
		args     []*Argument
		hasChain bool
		want     bool
	}{
		{
			name:     "epic-driven: single epic_id arg, no chain",
			args:     []*Argument{epicIDArg},
			hasChain: false,
			want:     true,
		},
		{
			name:     "not epic-driven: epic_id arg with chain",
			args:     []*Argument{epicIDArg},
			hasChain: true,
			want:     false,
		},
		{
			name:     "not epic-driven: multiple args",
			args:     []*Argument{epicIDArg, goalArg},
			hasChain: false,
			want:     false,
		},
		{
			name:     "not epic-driven: different arg name",
			args:     []*Argument{goalArg},
			hasChain: false,
			want:     false,
		},
		{
			name:     "not epic-driven: no args",
			args:     nil,
			hasChain: false,
			want:     false,
		},
		{
			name:     "not epic-driven: standard workflow with chain",
			args:     []*Argument{goalArg},
			hasChain: true,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var chain *Chain
			if tt.hasChain {
				chain, err = NewChain().
					Node("step1", "Step 1", "step1.md").
					Build()
				require.NoError(t, err)
			}

			reg := newRegistration("workflow", "test", "v1", "Test", "Desc", "", "", "", chain, nil, tt.args, SourceBuiltIn)
			require.Equal(t, tt.want, reg.IsEpicDriven())
		})
	}
}
