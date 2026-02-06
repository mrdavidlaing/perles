package communityworkflows

import (
	"fmt"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zjrosen/perles/internal/templates"
)

func TestRegistryFS_JokeContestExists(t *testing.T) {
	fsys := RegistryFS()

	// The joke-contest workflow should be accessible at workflows/joke-contest/template.yaml.
	data, err := fs.ReadFile(fsys, "workflows/joke-contest/template.yaml")
	require.NoError(t, err, "should be able to read joke-contest template.yaml via RegistryFS")
	require.NotEmpty(t, data, "joke-contest template.yaml should not be empty")
}

func TestRegistryFS_SharedTemplatesExist(t *testing.T) {
	fsys := RegistryFS()

	for _, name := range []string{"v1-human-review.md", "v1-epic-instructions.md"} {
		data, err := fs.ReadFile(fsys, "workflows/"+name)
		require.NoError(t, err, "shared template %s should be readable via RegistryFS", name)
		require.NotEmpty(t, data, "shared template %s should not be empty", name)
	}
}

func TestRegistryFS_ReadmeExists(t *testing.T) {
	fsys := RegistryFS()

	data, err := fs.ReadFile(fsys, "workflows/README.md")
	require.NoError(t, err, "workflows/README.md placeholder should exist")
	require.NotEmpty(t, data, "workflows/README.md should not be empty")
}

func TestSharedTemplateSyncCheck(t *testing.T) {
	communityFS := RegistryFS()
	builtinFS := templates.RegistryFS()

	sharedTemplates := []string{"v1-human-review.md", "v1-epic-instructions.md"}

	for _, name := range sharedTemplates {
		t.Run(name, func(t *testing.T) {
			path := "workflows/" + name

			communityData, err := fs.ReadFile(communityFS, path)
			require.NoError(t, err, "failed to read %s from communityworkflows.RegistryFS()", name)

			builtinData, err := fs.ReadFile(builtinFS, path)
			require.NoError(t, err, "failed to read %s from templates.RegistryFS()", name)

			require.Equal(t, builtinData, communityData,
				fmt.Sprintf("community copy of %s has diverged from built-in; "+
					"run 'cp internal/templates/workflows/%s communityworkflows/workflows/%s' to sync",
					name, name, name))
		})
	}
}
