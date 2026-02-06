package registry

// Source indicates where a registration originated from.
type Source int

const (
	// SourceBuiltIn indicates a registration bundled with the application.
	SourceBuiltIn Source = iota
	// SourceCommunity indicates a community-contributed registration.
	SourceCommunity
	// SourceUser indicates a registration from the user's configuration directory.
	SourceUser
)

// String returns a human-readable representation of the Source.
func (s Source) String() string {
	switch s {
	case SourceBuiltIn:
		return "built-in"
	case SourceCommunity:
		return "community"
	case SourceUser:
		return "user"
	default:
		return "unknown"
	}
}

// Registration represents a registered workflow namespace+version
type Registration struct {
	namespace    string      // e.g., "workflow"
	key          string      // e.g., "planning-standard"
	version      string      // e.g., "v1"
	name         string      // e.g., "Standard Planning Workflow"
	description  string      // e.g., "Three-phase workflow: Research, Propose, Plan"
	epicTemplate string      // template filename for epic content (e.g., "v1-research-proposal-epic.md")
	systemPrompt string      // template filename for system prompt content (e.g., "epic_driven.md")
	artifactPath string      // path prefix for artifacts (default: ".spec")
	dag          *Chain      // DAG-based workflow chain (replaces flat chain)
	labels       []string    // e.g., ["lang:go", "category:workflow"]
	arguments    []*Argument // user-configurable parameters for workflow
	source       Source      // origin of registration (built-in or user)
}

// newRegistration creates a registration (used by builder)
func newRegistration(namespace, key, version, name, description, epicTemplate, systemPrompt, artifactPath string, dag *Chain, labels []string, arguments []*Argument, source Source) *Registration {
	return &Registration{
		namespace:    namespace,
		key:          key,
		version:      version,
		name:         name,
		description:  description,
		epicTemplate: epicTemplate,
		systemPrompt: systemPrompt,
		artifactPath: artifactPath,
		dag:          dag,
		labels:       labels,
		arguments:    arguments,
		source:       source,
	}
}

// Key returns the registration key (unique identifier per type)
func (r *Registration) Key() string {
	return r.key
}

// Name returns the human-readable name
func (r *Registration) Name() string {
	return r.name
}

// Description returns the description for AI agents
func (r *Registration) Description() string {
	return r.description
}

// Namespace returns the registration namespace
func (r *Registration) Namespace() string {
	return r.namespace
}

// Version returns the registration version
func (r *Registration) Version() string {
	return r.version
}

// DAG returns the DAG-based workflow chain
func (r *Registration) DAG() *Chain {
	return r.dag
}

// Labels returns the registration labels
func (r *Registration) Labels() []string {
	return r.labels
}

// EpicTemplate returns the template filename for epic content
func (r *Registration) EpicTemplate() string {
	return r.epicTemplate
}

// SystemPrompt returns the template filename for system prompt content
func (r *Registration) SystemPrompt() string {
	return r.systemPrompt
}

// ArtifactPath returns the path prefix for artifacts.
// Returns empty string if not explicitly set.
func (r *Registration) ArtifactPath() string {
	return r.artifactPath
}

// Arguments returns the workflow's user-configurable parameters.
func (r *Registration) Arguments() []*Argument {
	return r.arguments
}

// Source returns the registration's source (built-in or user).
func (r *Registration) Source() Source {
	return r.source
}

// IsEpicDriven returns true if this workflow uses an existing epic from the tracker
// rather than creating one. An epic-driven workflow has a single "epic_id" argument
// and no DAG nodes (tasks come from the BD tracker).
func (r *Registration) IsEpicDriven() bool {
	// Must have exactly one argument named "epic_id"
	if len(r.arguments) != 1 {
		return false
	}
	if r.arguments[0].Key() != "epic_id" {
		return false
	}
	// Must have no nodes (or empty DAG)
	if r.dag != nil && len(r.dag.Nodes()) > 0 {
		return false
	}
	return true
}
