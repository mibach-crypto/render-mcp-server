package authn

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/render-oss/render-mcp-server/pkg/cfg"
)

const apiTokenKey string = "token"

var ErrNotAuthorized = errors.New("resource not found")

func APITokenFromContext(ctx context.Context) string {
	if token, ok := ctx.Value(apiTokenKey).(string); ok {
		return token
	}
	return ""
}

func ContextWithAPIToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, apiTokenKey, token)
}

func ContextWithAPITokenFromHeader(ctx context.Context, req *http.Request) context.Context {
	headerToken := req.Header.Get("Authorization")

	// Note: we strip the "Bearer " prefix if it exists
	// MCP Inspector attaches this prefix automatically, but it's unclear how standard this is
	if len(headerToken) > 7 && headerToken[:7] == "Bearer " {
		headerToken = headerToken[7:]
	}

	// If AUTH_TOKEN is configured on the server, we use it for validation.
	if authToken, ok := os.LookupEnv("AUTH_TOKEN"); ok && authToken != "" {
		// If the client's token matches the expected auth token, we're authenticated.
		if headerToken == authToken {
			return ContextWithAPIToken(ctx, headerToken)
		}
		// If they don't match, the request is unauthenticated.
		return ctx
	}

	// If AUTH_TOKEN is not set, fall back to the original behavior:
	// Place the token from the header into the context, assuming it's a Render API key.
	if headerToken == "" {
		return ctx
	}

	return ContextWithAPIToken(ctx, headerToken)
}

func ContextWithAPITokenFromConfig(ctx context.Context) context.Context {
	token := cfg.GetAPIKey()
	if token == "" {
		log.Fatal("Error getting API token from config")
	}
	return ContextWithAPIToken(ctx, token)
}
