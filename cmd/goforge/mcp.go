package main

import (
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"

	"github.com/MMaZX/goforge/mcp"
)

func newMCPCmd(flags *globalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Run the GoForge MCP server over stdio",
		RunE: func(cmd *cobra.Command, args []string) error {
			engine, conn, cfg, err := loadEngine(cmd, flags)
			if err != nil {
				return err
			}
			defer conn.Close()

			server := mcp.NewServer(engine, cfg, version)
			return server.Run(cmd.Context(), &sdk.StdioTransport{})
		},
	}
}
