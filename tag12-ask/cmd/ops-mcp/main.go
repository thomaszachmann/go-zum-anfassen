// ops-mcp: MCP-Server (stdio) mit Ops-Tools für KI-Assistenten.
package main

import (
	"context"
	"log"
	"os"

	"example.com/ask/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type diskIn struct {
	Path string `json:"path" jsonschema:"Pfad, z. B. / oder /var"`
}

func diskFree(_ context.Context, _ *mcp.CallToolRequest, in diskIn) (
	*mcp.CallToolResult, tools.Disk, error) {
	d, err := tools.DiskFree(in.Path)
	return nil, d, err // Fehler landen als isError beim Client
}

func newServer() *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: "ops-mcp", Version: "v0.1.0"}, nil)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "disk_free",
		Description: "Freier Speicherplatz eines Pfads auf diesem Rechner",
	}, diskFree)
	return s
}

func main() {
	log.SetOutput(os.Stderr) // stdout gehört dem JSON-RPC-Protokoll
	if err := newServer().Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
