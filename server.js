const express = require('express');
const { spawn } = require('child_process');
const http = require('http');

const app = express();
const PORT = process.env.PORT || 10000;
const AUTH_TOKEN = process.env.AUTH_TOKEN;

let mcpServerProcess = null;
let mcpServerHealthy = false;
let lastHealthCheck = null;

// Health check endpoint for Render
app.get('/health', (req, res) => {
    const status = {
          status: 'healthy',
          timestamp: new Date().toISOString(),
          uptime: process.uptime(),
          mcpServer: {
                  running: mcpServerProcess !== null && mcpServerProcess.exitCode === null,
                  healthy: mcpServerHealthy,
                  lastCheck: lastHealthCheck
          },
          environment: {
                  hasAuthToken: !!AUTH_TOKEN,
                  port: PORT
          }
    };

          console.log('Health check:', JSON.stringify(status));
    res.status(200).json(status);
});

// Root endpoint
app.get('/', (req, res) => {
    res.json({
          service: 'TypingMind MCP Server Wrapper',
          version: '1.0.0',
          status: 'running',
          endpoints: {
                  health: '/health',
                  info: '/',
                  mcp: 'TypingMind MCP running on port 10000'
          }
    });
});

// Start the TypingMind MCP server as a child process
function startMCPServer() {
    if (!AUTH_TOKEN) {
          console.error('ERROR: AUTH_TOKEN environment variable is not set');
          process.exit(1);
    }

  console.log('Starting TypingMind MCP server...');

  mcpServerProcess = spawn('npx', ['@typingmind/mcp', AUTH_TOKEN], {
        stdio: ['ignore', 'pipe', 'pipe'],
        env: { ...process.env }
  });

  mcpServerProcess.stdout.on('data', (data) => {
        console.log('[MCP Server]:', data.toString().trim());
        // Server is producing output, likely healthy
                                 mcpServerHealthy = true;
        lastHealthCheck = new Date().toISOString();
  });

  mcpServerProcess.stderr.on('data', (data) => {
        console.error('[MCP Server Error]:', data.toString().trim());
  });

  mcpServerProcess.on('error', (error) => {
        console.error('Failed to start MCP server:', error);
        mcpServerHealthy = false;
  });

  mcpServerProcess.on('exit', (code, signal) => {
        console.log(`MCP server exited with code ${code} and signal ${signal}`);
        mcpServerHealthy = false;
        mcpServerProcess = null;

                          // Restart after 5 seconds if it crashes
                          setTimeout(() => {
                                  console.log('Restarting MCP server...');
                                  startMCPServer();
                          }, 5000);
  });

  // Give the server time to start
  setTimeout(() => {
        mcpServerHealthy = true;
        lastHealthCheck = new Date().toISOString();
        console.log('MCP server startup period completed');
  }, 3000);
}

// Graceful shutdown
process.on('SIGTERM', () => {
    console.log('SIGTERM received, shutting down gracefully...');
    if (mcpServerProcess) {
          mcpServerProcess.kill('SIGTERM');
    }
    server.close(() => {
          console.log('Server closed');
          process.exit(0);
    });
});

process.on('SIGINT', () => {
    console.log('SIGINT received, shutting down gracefully...');
    if (mcpServerProcess) {
          mcpServerProcess.kill('SIGINT');
    }
    server.close(() => {
          console.log('Server closed');
          process.exit(0);
    });
});

// Start both servers
const server = app.listen(PORT, '0.0.0.0', () => {
    console.log(`Health check server running on port ${PORT}`);
    console.log(`Environment: PORT=${PORT}, AUTH_TOKEN=${AUTH_TOKEN ? 'SET' : 'NOT SET'}`);
    startMCPServer();
});

console.log('TypingMind MCP Server Wrapper initialized');
console.log('This wrapper provides /health endpoint for Render while running the MCP server');
