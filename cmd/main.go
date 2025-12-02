package main

import (
	"applovin-max-mcp/internal"

	"github.com/mark3labs/mcp-go/server"
)

func main() {
	mcpServer := server.NewMCPServer(
		"Applovin Max Demo",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	internal.AppendCapability(mcpServer,
		internal.NewRevenueReportCapability(),
		internal.NewCohortRequestCapability(),
	)

	err := server.ServeStdio(mcpServer)
	if err != nil {
		panic(err)
	}
}
