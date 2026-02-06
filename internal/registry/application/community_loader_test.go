package registry

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"

	"github.com/zjrosen/perles/internal/registry/domain"
)

// validCommunityYAML returns a template.yaml with two workflow registrations for testing.
func validCommunityYAML() string {
	return `registry:
  - namespace: "workflow"
    key: "joke-contest"
    version: "v1"
    name: "Joke Contest"
    description: "A community joke contest workflow"
    labels:
      - "community"
      - "fun"
    nodes:
      - key: "research"
        name: "Research Phase"
        template: "research.md"
        outputs:
          - key: "research"
            file: "research.md"
  - namespace: "workflow"
    key: "code-review"
    version: "v1"
    name: "Code Review"
    description: "A community code review workflow"
    labels:
      - "community"
      - "review"
    nodes:
      - key: "review"
        name: "Review Phase"
        template: "review.md"
        outputs:
          - key: "review"
            file: "review.md"
`
}

func TestLoadCommunityRegistryFromFS_NilSource(t *testing.T) {
	regs, fsys, err := LoadCommunityRegistryFromFS(nil)

	require.NoError(t, err)
	require.Nil(t, regs)
	require.Nil(t, fsys)
}

func TestLoadCommunityRegistryFromFS_EmptyEnabledIDs(t *testing.T) {
	communityFS := fstest.MapFS{
		"workflows/joke-contest/template.yaml": &fstest.MapFile{Data: []byte(validCommunityYAML())},
	}

	source := &CommunitySource{
		FS:         communityFS,
		EnabledIDs: []string{},
	}

	regs, fsys, err := LoadCommunityRegistryFromFS(source)

	require.NoError(t, err)
	require.Nil(t, regs)
	require.Nil(t, fsys)
}

func TestLoadCommunityRegistryFromFS_LoadsFilteredRegistrations(t *testing.T) {
	communityFS := fstest.MapFS{
		"workflows/joke-contest/template.yaml": &fstest.MapFile{Data: []byte(validCommunityYAML())},
		"workflows/joke-contest/research.md":   &fstest.MapFile{Data: []byte("# Research")},
		"workflows/joke-contest/review.md":     &fstest.MapFile{Data: []byte("# Review")},
	}

	source := &CommunitySource{
		FS:         communityFS,
		EnabledIDs: []string{"workflow/joke-contest"},
	}

	regs, fsys, err := LoadCommunityRegistryFromFS(source)

	require.NoError(t, err)
	require.NotNil(t, fsys)
	require.Len(t, regs, 1, "should only return the enabled workflow")

	reg := regs[0]
	require.Equal(t, "workflow", reg.Namespace())
	require.Equal(t, "joke-contest", reg.Key())
	require.Equal(t, registry.SourceCommunity, reg.Source())
}

func TestLoadCommunityRegistryFromFS_UnmatchedIDWarns(t *testing.T) {
	communityFS := fstest.MapFS{
		"workflows/joke-contest/template.yaml": &fstest.MapFile{Data: []byte(validCommunityYAML())},
		"workflows/joke-contest/research.md":   &fstest.MapFile{Data: []byte("# Research")},
		"workflows/joke-contest/review.md":     &fstest.MapFile{Data: []byte("# Review")},
	}

	source := &CommunitySource{
		FS:         communityFS,
		EnabledIDs: []string{"workflow/nonexistent"},
	}

	regs, fsys, err := LoadCommunityRegistryFromFS(source)

	// Should not error - WARN is logged instead
	require.NoError(t, err)
	require.NotNil(t, fsys, "should return FS even when no IDs match")
	require.Empty(t, regs, "no registrations should match a nonexistent ID")
}

func TestLoadCommunityRegistryFromFS_ZeroRegistrations(t *testing.T) {
	// FS with only a README.md - no template.yaml files
	communityFS := fstest.MapFS{
		"workflows/README.md": &fstest.MapFile{Data: []byte("# Community Workflows")},
	}

	source := &CommunitySource{
		FS:         communityFS,
		EnabledIDs: []string{"workflow/something"},
	}

	regs, fsys, err := LoadCommunityRegistryFromFS(source)

	// Zero registrations triggers the "no workflow registrations found" error path
	// which should be WARN+skip, never crash
	require.NoError(t, err)
	require.NotNil(t, fsys, "should return FS even with zero registrations")
	require.Nil(t, regs, "should return nil registrations for zero-registration FS")
}

func TestLoadCommunityRegistryFromFS_MalformedYAML(t *testing.T) {
	communityFS := fstest.MapFS{
		"workflows/broken/template.yaml": &fstest.MapFile{Data: []byte(`registry:
  - namespace: broken
    key: [this is invalid yaml
`)},
	}

	source := &CommunitySource{
		FS:         communityFS,
		EnabledIDs: []string{"workflow/broken"},
	}

	regs, fsys, err := LoadCommunityRegistryFromFS(source)

	// Malformed YAML should be WARN+skip, never crash
	require.NoError(t, err)
	require.NotNil(t, fsys, "should return FS even with malformed YAML")
	require.Nil(t, regs, "should return nil registrations for malformed YAML")
}

func TestLoadCommunityRegistryFromFS_BareKeyNormalization(t *testing.T) {
	// Bare keys (without namespace/) should be auto-prefixed with "workflow/"
	communityFS := fstest.MapFS{
		"workflows/joke-contest/template.yaml": &fstest.MapFile{Data: []byte(validCommunityYAML())},
		"workflows/joke-contest/research.md":   &fstest.MapFile{Data: []byte("# Research")},
		"workflows/joke-contest/review.md":     &fstest.MapFile{Data: []byte("# Review")},
	}

	source := &CommunitySource{
		FS:         communityFS,
		EnabledIDs: []string{"joke-contest"}, // bare key, no "workflow/" prefix
	}

	regs, fsys, err := LoadCommunityRegistryFromFS(source)

	require.NoError(t, err)
	require.NotNil(t, fsys)
	require.Len(t, regs, 1, "bare key should resolve to workflow/joke-contest")

	reg := regs[0]
	require.Equal(t, "workflow", reg.Namespace())
	require.Equal(t, "joke-contest", reg.Key())
	require.Equal(t, registry.SourceCommunity, reg.Source())
}

func TestLoadCommunityRegistryFromFS_MixedBareAndQualifiedKeys(t *testing.T) {
	// Mix of bare keys and fully-qualified keys should both work
	communityFS := fstest.MapFS{
		"workflows/joke-contest/template.yaml": &fstest.MapFile{Data: []byte(validCommunityYAML())},
		"workflows/joke-contest/research.md":   &fstest.MapFile{Data: []byte("# Research")},
		"workflows/joke-contest/review.md":     &fstest.MapFile{Data: []byte("# Review")},
	}

	source := &CommunitySource{
		FS:         communityFS,
		EnabledIDs: []string{"joke-contest", "workflow/code-review"}, // mixed formats
	}

	regs, fsys, err := LoadCommunityRegistryFromFS(source)

	require.NoError(t, err)
	require.NotNil(t, fsys)
	require.Len(t, regs, 2, "both bare and qualified keys should resolve")

	keys := make(map[string]bool)
	for _, r := range regs {
		keys[r.Key()] = true
	}
	require.True(t, keys["joke-contest"], "bare key joke-contest should resolve")
	require.True(t, keys["code-review"], "qualified key workflow/code-review should resolve")
}

func TestNormalizeCommunityID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"joke-contest", "workflow/joke-contest"},
		{"workflow/joke-contest", "workflow/joke-contest"},
		{"custom-ns/my-workflow", "custom-ns/my-workflow"},
		{"simple", "workflow/simple"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			require.Equal(t, tt.expected, normalizeCommunityID(tt.input))
		})
	}
}

func TestLoadCommunityRegistryFromFS_AllIDsMatch(t *testing.T) {
	communityFS := fstest.MapFS{
		"workflows/joke-contest/template.yaml": &fstest.MapFile{Data: []byte(validCommunityYAML())},
		"workflows/joke-contest/research.md":   &fstest.MapFile{Data: []byte("# Research")},
		"workflows/joke-contest/review.md":     &fstest.MapFile{Data: []byte("# Review")},
	}

	source := &CommunitySource{
		FS:         communityFS,
		EnabledIDs: []string{"workflow/joke-contest", "workflow/code-review"},
	}

	regs, fsys, err := LoadCommunityRegistryFromFS(source)

	require.NoError(t, err)
	require.NotNil(t, fsys)
	require.Len(t, regs, 2, "both enabled IDs should be returned")

	// Verify both registrations are present
	keys := make(map[string]bool)
	for _, r := range regs {
		keys[r.Key()] = true
		require.Equal(t, registry.SourceCommunity, r.Source())
	}
	require.True(t, keys["joke-contest"], "joke-contest should be in results")
	require.True(t, keys["code-review"], "code-review should be in results")
}
