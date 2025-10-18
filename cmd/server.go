package cmd

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/server"
	"github.com/render-oss/render-mcp-server/pkg/auth"
	"github.com/render-oss/render-mcp-server/pkg/authn"
	"github.com/render-oss/render-mcp-server/pkg/cfg"
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
		startTime := time.Now()
		host := firstNonEmptyEnv([]string{"HOST", "MCP_HOST", "TYPINGMIND_HOST"}, "0.0.0.0")
		port := firstNonEmptyEnv([]string{"PORT", "MCP_PORT", "TYPINGMIND_PORT"}, "10000")
		listenAddr := net.JoinHostPort(host, port)

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

		mux := http.NewServeMux()
		httpServer := &http.Server{
			Addr:    listenAddr,
			Handler: mux,
		}

		healthHandler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			status := map[string]any{
				"status":              "ok",
				"version":             cfg.Version,
				"transport":           transport,
				"timestamp":           time.Now().UTC().Format(time.RFC3339),
				"uptimeSeconds":       time.Since(startTime).Seconds(),
				"authTokenConfigured": authTokenConfigured(),
				"endpoints": map[string]string{
					"mcp":    "/mcp",
					"health": "/health",
				},
				"listener": map[string]string{
					"host": host,
					"port": port,
				},
			}

			if err := json.NewEncoder(w).Encode(status); err != nil {
				log.Printf("failed to encode health response: %v\n", err)
			}
		}

		mux.HandleFunc("/health", healthHandler)

		streamableServer := server.NewStreamableHTTPServer(
			s,
			server.WithHTTPContextFunc(multicontext.MultiHTTPContextFunc(
				session.ContextWithHTTPSession(sessionStore),
				authn.ContextWithAPITokenFromHeader,
			)),
			server.WithStreamableHTTPServer(httpServer),
		)

		mux.Handle("/mcp", streamableServer)

		log.Printf("Starting HTTP MCP server on %s\n", listenAddr)

		err := streamableServer.Start(listenAddr)
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

func firstNonEmptyEnv(keys []string, fallback string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return fallback
}

func authTokenConfigured() bool {
	value, ok := os.LookupEnv("AUTH_TOKEN")
	return ok && value != ""
}
