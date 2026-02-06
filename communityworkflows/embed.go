// Package communityworkflows embeds community-contributed workflow templates.
//
// Community workflows are contributed via PRs, embedded at compile time alongside
// built-in workflows, but kept in a separate package to distinguish governance
// and origin. Users opt in to specific community workflows through the
// community_workflows config field under orchestration.
package communityworkflows

import (
	"embed"
	"io/fs"
)

// registryTemplates embeds all registry-style community workflow templates from
// the workflows directory. The structure mirrors internal/templates/workflows/:
//   - workflows/<workflow-name>/template.yaml
//   - workflows/<workflow-name>/*.md (workflow-specific templates)
//   - workflows/*.md (shared templates like v1-epic-instructions.md)
//
//go:embed workflows
var registryTemplates embed.FS

// RegistryFS returns the embedded filesystem containing registry-style community
// workflow templates. This is used by the registry service to load community
// workflow registrations alongside built-in and user workflows.
func RegistryFS() fs.FS {
	return registryTemplates
}
