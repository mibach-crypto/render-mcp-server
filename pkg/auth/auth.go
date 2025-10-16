package auth

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/render-oss/render-mcp-server/pkg/config"
)

func AddTools(s *server.MCPServer) {
	tool := mcp.NewTool("login",
		mcp.WithDescription("Authenticate with the Render API. You can get an API key from https://dashboard.render.com/account/api-keys."),
		mcp.WithString("apiKey",
			mcp.Required(),
			mcp.Description("Your Render API key."),
		),
	)
	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		apiKey, err := request.RequireString("apiKey")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		err = config.SetAPIConfig(config.APIConfig{
			APIKey: apiKey,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to set api key: %w", err)
		}
		return mcp.NewToolResultText("successfully authenticated"), nil
	}
	s.AddTool(tool, handler)
}