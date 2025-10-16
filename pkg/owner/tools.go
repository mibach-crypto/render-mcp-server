package owner

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/render-oss/render-mcp-server/pkg/client"
	"github.com/render-oss/render-mcp-server/pkg/pointers"
	"github.com/render-oss/render-mcp-server/pkg/session"
	"github.com/render-oss/render-mcp-server/pkg/validate"
)

func AddTools(s *server.MCPServer, c *client.ClientWithResponses) {
	ownerRepo := NewRepo(c)

	tool, handler := listWorkspaces(ownerRepo)
	s.AddTool(*tool, handler)

	tool, handler = selectWorkspace()
	s.AddTool(*tool, handler)

	tool, handler = getSelectedWorkspace()
	s.AddTool(*tool, handler)
}

func listWorkspaces(ownerRepo *Repo) (*mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("list_workspaces",
		mcp.WithDescription("List the workspaces that you have access to"),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:          "List workspaces",
			ReadOnlyHint:   pointers.From(true),
			IdempotentHint: pointers.From(true),
			OpenWorldHint:  pointers.From(true),
		}),
	)
	return &tool,
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			workspaces, err := ownerRepo.ListOwners(ctx, ListInput{})
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			respJSON, err := json.Marshal(workspaces)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			resultText := ""

			if len(workspaces) == 1 {
				err = session.FromContext(ctx).SetWorkspace(ctx, workspaces[0].Id)
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				resultText = "Only one workspace found, automatically selected it"
			}

			resultText += string(respJSON)
			return mcp.NewToolResultText(resultText), nil
		}
}

func selectWorkspace() (*mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("select_workspace",
		mcp.WithDescription("Select a workspace to use for all actions. This tool should "+
			"only be used after explicitly asking the user to select one, it should not be invoked "+
			"as part of an automated process. Having the wrong workspace selected can lead to "+
			"destructive actions being performed on unintended resources."),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:          "Select workspace",
			IdempotentHint: pointers.From(true),
			OpenWorldHint:  pointers.From(true),
		}),
		mcp.WithString("ownerID",
			mcp.Required(),
			mcp.Description("The ID of the owner to select"),
		),
	)
	return &tool,
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			ownerID, err := validate.RequiredToolParam[string](request, "ownerID")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			err = session.FromContext(ctx).SetWorkspace(ctx, ownerID)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			return mcp.NewToolResultText("Workspace selected"), nil
		}
}

func getSelectedWorkspace() (*mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("get_selected_workspace",
		mcp.WithDescription("Get the currently selected workspace"),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:          "Get selected workspace",
			ReadOnlyHint:   pointers.From(true),
			IdempotentHint: pointers.From(true),
			OpenWorldHint:  pointers.From(true),
		}),
	)
	return &tool,
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			workspace, err := session.FromContext(ctx).GetWorkspace(ctx)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			return mcp.NewToolResultText(
				fmt.Sprintf("The currently selected workspace is: %s", workspace),
			), nil
		}
}
