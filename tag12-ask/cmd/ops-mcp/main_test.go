package main

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestDiskFreeTool(t *testing.T) {
	ctx := context.Background()
	st, ct := mcp.NewInMemoryTransports()
	if _, err := newServer().Connect(ctx, st, nil); err != nil {
		t.Fatal(err)
	}
	c := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "v0"}, nil)
	cs, err := c.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cs.Close()

	tests := []struct {
		path    string
		wantErr bool
	}{{"/", false}, {"/gibt/es/nicht", true}}
	for _, tt := range tests {
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "disk_free", Arguments: map[string]any{"path": tt.path}})
		if err != nil {
			t.Fatal(err)
		}
		if res.IsError != tt.wantErr {
			t.Errorf("%s: IsError=%v", tt.path, res.IsError)
		}
		t.Logf("%s -> %v", tt.path, res.Content[0].(*mcp.TextContent).Text)
	}
}
