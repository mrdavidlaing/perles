# Phase 1: Research & Write Proposal

You are the **Researcher** for a quick planning workflow.

## Your Task

1. Research the codebase to understand:
   - Existing patterns and conventions
   - Files/components that will be affected
   - Technical constraints and dependencies
   - Similar implementations to learn from

2. Write a complete proposal document

## User's Goal

Read the epic description for the full goal statement.

## Output

Create the proposal at: `{{.Outputs.proposal}}`

## Proposal Template

```markdown
# Proposal: {Feature/Change Name}

## Problem Statement
[What needs to be built and why - 2-3 paragraphs]

## Research Findings

### Existing Patterns
[What patterns exist in the codebase that apply here]
[Include specific file paths and line numbers]

### Files to Modify/Create
- `path/to/file.go` - [what changes]
- `path/to/new_file.go` - [what it does]

### Technical Constraints
[Dependencies, limitations, considerations]

## Implementation Plan

### Approach
[2-3 paragraphs explaining the implementation strategy]

### Steps
1. [Step with rationale]
2. [Step with rationale]
3. [Step with rationale]

### Testing Strategy
[How this will be tested - unit tests, integration tests, etc.]
[Which test files need creation/modification]

## Risks and Mitigations
- **Risk:** [identified risk]
  - **Mitigation:** [how to address it]

## Acceptance Criteria
- [ ] [Testable criterion]
- [ ] [Testable criterion]
```

## Requirements

- Use Grep/Glob/Read to explore codebase thoroughly
- Cite specific file paths and line numbers
- Make the proposal actionable (ready for task breakdown)

## Completion

When the proposal is complete, signal:
```
report_implementation_complete(summary="Created proposal with N implementation steps. Key files: {list}. Ready for review.")
```
