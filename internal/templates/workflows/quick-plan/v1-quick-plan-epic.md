# Quick Plan: {{.Slug}}

## Overview

A lightweight 4-worker planning cycle that produces a researched proposal with actionable beads tasks. Single workers per phase with review gates ensure quality without overhead.

## Goal

{{.Args.goal}}

## Worker Roles

| Worker | Role | Responsibility |
|--------|------|----------------|
| worker-1 | Researcher | Deep research and write complete proposal |
| worker-2 | Proposal Reviewer | Verify research accuracy and completeness |
| worker-3 | Planner | Break proposal into beads epic and tasks |
| worker-4 | Task Reviewer | Ensure tasks are well-structured and implementable |

## Workflow Phases

```
Phase 1:  worker-1 (Researcher)        → Research & write proposal
Phase 2:  worker-2 (Proposal Reviewer) → Review proposal
Phase 2B: worker-1 (Researcher)        → Address feedback (if needed)
Phase 3:  worker-3 (Planner)           → Create epic & tasks
Phase 4:  worker-4 (Task Reviewer)     → Review tasks
Phase 4B: worker-3 (Planner)           → Address feedback (if needed)
Phase 5:  worker-1 (Researcher)        → Final summary
```

## Review Gates

| Gate | Pass Condition | Failure Action |
|------|----------------|----------------|
| Proposal Review | Research accurate, approach feasible | Researcher revises |
| Task Review | Tasks clear, tests included, deps correct | Planner revises |

## Output Artifacts

- `{{.Outputs.proposal}}` - Research and implementation proposal

## Key Principles

1. **Research citations must be accurate** - File paths must exist, patterns must be real
2. **Tests are included with tasks** - NOT deferred to later
3. **Review loops are quality gates** - Don't skip or rush past them
4. **Tasks should be independently executable** - Clear scope and acceptance criteria

## Execution Instructions

1. **Spawn all 4 workers** at the start
2. **Follow phase order** - Review gates enforce quality
3. **Handle review loops** - If reviewer says CHANGES NEEDED, loop back
4. **Mark tasks complete immediately** when workers signal completion

## Success Criteria

- [ ] Proposal has concrete file paths and patterns cited
- [ ] Review loops catch real issues (not rubber-stamped)
- [ ] Each task includes test requirements
- [ ] Tasks are small enough for single-session completion
- [ ] Dependencies between tasks are logical
- [ ] Epic links to proposal document
