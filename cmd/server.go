package cmd

import (
	"log"
	"os"

	"github.com/mark3labs/mcp-go/server"
	"github.com/render-oss/render-mcp-server/pkg/authn"
	"github.com/render-oss/render-mcp-server/pkg/cfg"
	"github.com/render-oss/render-mcp-server/pkg/auth"
	"github.com/render-oss/render-mcp-server/pkg/client"
	"github.com/render-oss/render-mcp-server/pkg/config"
	"github.com/render-oss/render-mcp-server/pkg/deploy"
	"github.com/render-oss/render-mcp-server/pkg/keyvalue"
	"github.com/render-oss/render-mcp-server/pkg/logs"
	"github.com/render-oss/render-mcp-server/pkg/metrics"
	"github.com/render-oss/render-mcp-server/pkg/multicontext"
	"github.com/render-oss/render-mcp-server/pkg/owner"
	"github.com/render-oss/render-mcp-server/pkg/postgres"
	"github.com/render-oss/render-mcp-server/pkg/service"
	"github.com/render-oss/render-mcp-server/pkg/session"
)

func Serve(transport string) *server.MCPServer {
	// Create MCP server
	s := server.NewMCPServer(
		"render-mcp-server",
		cfg.Version,
	)

	c, err := client.NewDefaultClient()
	if err != nil {
		if err == config.ErrLogin {
			auth.AddTools(s)
		} else {
			// TODO: We can't create a client unless we're logged in, so we should handle that error case.
			panic(err)
		}
	} else {
		owner.AddTools(s, c)
		service.AddTools(s, c)
		deploy.AddTools(s, c)
		postgres.AddTools(s, c)
		keyvalue.AddTools(s, c)
		logs.AddTools(s, c)
		metrics.AddTools(s, c)
	}

	if transport == "http" {
		var sessionStore session.Store
		if redisURL, ok := os.LookupEnv("REDIS_URL"); ok {
			log.Print("using Redis session store\n")
			sessionStore, err = session.NewRedisStore(redisURL)
			if err != nil {
				log.Fatalf("failed to initialize Redis session store: %v", err)
			}
		} else {
			log.Print("using in-memory session store\n")
			sessionStore = session.NewInMemoryStore()
		}
		port := os.Getenv("PORT")
		if port == "" {
			port = "10000"
		}

		err := server.
			NewStreamableHTTPServer(s, server.WithHTTPContextFunc(multicontext.MultiHTTPContextFunc(
				session.ContextWithHTTPSession(sessionStore),
				authn.ContextWithAPITokenFromHeader,
			))).
			Start(":" + port)
		if err != nil {
			log.Fatalf("Starting Streamable server: %v\n:", err)
		}
	} else {
		err := server.ServeStdio(s, server.WithStdioContextFunc(multicontext.MultiStdioContextFunc(
			session.ContextWithStdioSession,
			authn.ContextWithAPITokenFromConfig,
		)))
		if err != nil {
			log.Fatalf("Starting STDIO server: %v\n", err)
		}
	}

	return s
}
