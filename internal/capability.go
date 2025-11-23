package internal

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type Capability struct {
	Tool    mcp.Tool
	Handler server.ToolHandlerFunc
}

func AppendCapability(server *server.MCPServer, capabilities ...Capability) {
	for _, c := range capabilities {
		server.AddTool(c.Tool, c.Handler)
	}
}
