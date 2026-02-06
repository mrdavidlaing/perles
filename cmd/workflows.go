package cmd

import (
	"fmt"

	"github.com/zjrosen/perles/internal/orchestration/workflow"

	"github.com/spf13/cobra"
)

var workflowsCmd = &cobra.Command{
	Use:   "workflows",
	Short: "List available workflow templates",
	Long:  `Display all workflow templates available for orchestration mode, including built-in and user-defined workflows.`,
	RunE:  runWorkflows,
}

func init() {
	rootCmd.AddCommand(workflowsCmd)
}

func runWorkflows(cmd *cobra.Command, args []string) error {
	// Load workflow registry using the unified config-aware pipeline.
	// This picks up built-in, community (if enabled), and user workflows
	// with proper precedence and config overrides.
	registry, err := workflow.NewRegistryWithConfig(cfg.Orchestration)
	if err != nil {
		return fmt.Errorf("loading workflows: %w", err)
	}

	// Get workflows grouped by source
	builtinWorkflows := registry.ListBySource(workflow.SourceBuiltIn)
	communityWorkflows := registry.ListBySource(workflow.SourceCommunity)
	userDefinedWorkflows := registry.ListBySource(workflow.SourceUser)

	// Print built-in workflows
	fmt.Println("Built-in Workflows:")
	if len(builtinWorkflows) == 0 {
		fmt.Println("  (none)")
	} else {
		maxLen := maxIDLen(builtinWorkflows)
		for _, wf := range builtinWorkflows {
			fmt.Printf("  %-*s  %s\n", maxLen, wf.ID, wf.Description)
		}
	}

	fmt.Println()

	// Print community workflows
	fmt.Println("Community Workflows:")
	if len(communityWorkflows) == 0 {
		fmt.Println("  (none enabled — configure in orchestration.community_workflows)")
	} else {
		maxLen := maxIDLen(communityWorkflows)
		for _, wf := range communityWorkflows {
			fmt.Printf("  %-*s  %s\n", maxLen, wf.ID, wf.Description)
		}
	}

	fmt.Println()

	// Print user workflows
	userDir := workflow.UserWorkflowDir()
	fmt.Printf("User Workflows (%s):\n", userDir)
	if len(userDefinedWorkflows) == 0 {
		fmt.Println("  (none)")
	} else {
		maxLen := maxIDLen(userDefinedWorkflows)
		for _, wf := range userDefinedWorkflows {
			fmt.Printf("  %-*s  %s\n", maxLen, wf.ID, wf.Description)
		}
	}

	fmt.Println()
	fmt.Println("Use workflows in orchestration mode with Ctrl+P")

	return nil
}

// maxIDLen returns the length of the longest workflow ID in the slice.
func maxIDLen(workflows []workflow.Workflow) int {
	maxLen := 0
	for _, wf := range workflows {
		if len(wf.ID) > maxLen {
			maxLen = len(wf.ID)
		}
	}
	return maxLen
}
