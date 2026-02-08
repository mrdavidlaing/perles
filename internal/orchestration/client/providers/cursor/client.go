package cursor

import (
	"context"

	"github.com/zjrosen/perles/internal/orchestration/client"
)

func init() {
	client.RegisterClient(client.ClientCursor, func() client.HeadlessClient {
		return NewClient()
	})
}

// CursorClient implements client.HeadlessClient for Cursor Agent CLI.
type CursorClient struct{}

// NewClient creates a new CursorClient.
func NewClient() *CursorClient {
	return &CursorClient{}
}

// Type returns the client type identifier.
func (c *CursorClient) Type() client.ClientType {
	return client.ClientCursor
}

// Spawn creates and starts a headless Cursor process.
// If cfg.SessionID is set, resumes an existing session.
// If cfg.SessionID is empty, creates a new session.
func (c *CursorClient) Spawn(ctx context.Context, cfg client.Config) (client.HeadlessProcess, error) {
	cursorCfg := configFromClient(cfg)
	if cfg.SessionID != "" {
		return Resume(ctx, cfg.SessionID, cursorCfg)
	}
	return Spawn(ctx, cursorCfg)
}

// Ensure CursorClient implements client.HeadlessClient at compile time.
var _ client.HeadlessClient = (*CursorClient)(nil)
