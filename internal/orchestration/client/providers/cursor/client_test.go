package cursor

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zjrosen/perles/internal/orchestration/client"
)

func TestCursorClient_Type(t *testing.T) {
	c := NewClient()
	require.Equal(t, client.ClientCursor, c.Type())
}

func TestCursorClient_Registration(t *testing.T) {
	// Verify Cursor client is registered via init() and can be created
	require.True(t, client.IsRegistered(client.ClientCursor), "ClientCursor should be registered via init()")

	c, err := client.NewClient(client.ClientCursor)
	require.NoError(t, err)
	require.NotNil(t, c)
	require.Equal(t, client.ClientCursor, c.Type())
}
