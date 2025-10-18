const express = require('express');
const { spawn } = require('child_process');

const app = express();

function coercePositiveNumber(value, fallback) {
  const parsed = Number(value);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : fallback;
}

const HEALTH_PORT = coercePositiveNumber(
  process.env.HEALTH_PORT || process.env.PORT,
  10000,
);
const MCP_PORT = coercePositiveNumber(
  process.env.MCP_PORT || process.env.MCP_SERVER_PORT || process.env.TYPINGMIND_PORT,
  8080,
);
const MCP_HOST = process.env.MCP_HOST || process.env.TYPINGMIND_HOST || '0.0.0.0';
const AUTH_TOKEN = process.env.AUTH_TOKEN || process.env.TYPINGMIND_AUTH_TOKEN;
const NPX_COMMAND = process.env.NPX_COMMAND || 'npx';
const MCP_PACKAGE = process.env.MCP_PACKAGE || '@typingmind/mcp';
const MCP_EXTRA_ARGS = (process.env.MCP_EXTRA_ARGS || '')
  .split(/\s+/)
  .filter(Boolean);
const MCP_RESTART_DELAY_MS = coercePositiveNumber(process.env.MCP_RESTART_DELAY_MS, 5000);

let mcpServerProcess = null;
let mcpServerHealthy = false;
let lastHealthCheck = null;
let restartTimer = null;

const serverState = () => ({
  running: Boolean(mcpServerProcess && mcpServerProcess.exitCode === null),
  healthy: mcpServerHealthy,
  lastCheck: lastHealthCheck,
});

function scheduleRestart() {
  if (restartTimer) {
    clearTimeout(restartTimer);
  }

  restartTimer = setTimeout(() => {
    console.warn('Attempting to restart TypingMind MCP server...');
    startMCPServer();
  }, MCP_RESTART_DELAY_MS);
}

// Health check endpoint for Render
app.get('/health', (req, res) => {
  const status = {
    status: mcpServerHealthy ? 'healthy' : 'degraded',
    timestamp: new Date().toISOString(),
    uptime: process.uptime(),
    mcpServer: serverState(),
    environment: {
      hasAuthToken: Boolean(AUTH_TOKEN),
      healthPort: HEALTH_PORT,
      mcpPort: MCP_PORT,
      mcpHost: MCP_HOST,
      package: MCP_PACKAGE,
    },
  };

  res.status(mcpServerHealthy ? 200 : 503).json(status);
});

// Root endpoint
app.get('/', (req, res) => {
  res.json({
    service: 'TypingMind MCP Server Wrapper',
    version: '1.1.0',
    status: mcpServerHealthy ? 'running' : 'starting',
    endpoints: {
      health: '/health',
      info: '/',
      mcp: `TypingMind MCP running on ${MCP_HOST}:${MCP_PORT}`,
    },
    configuration: {
      healthPort: HEALTH_PORT,
      mcpPort: MCP_PORT,
      mcpHost: MCP_HOST,
      package: MCP_PACKAGE,
    },
  });
});

function stopMCPServer(signal = 'SIGTERM') {
  if (!mcpServerProcess) {
    return;
  }

  console.log(`Forwarding ${signal} to TypingMind MCP server (pid: ${mcpServerProcess.pid})`);
  mcpServerProcess.kill(signal);
}

// Start the TypingMind MCP server as a child process
function startMCPServer() {
  if (!AUTH_TOKEN) {
    console.error('ERROR: AUTH_TOKEN environment variable is not set');
    process.exit(1);
  }

  if (mcpServerProcess) {
    console.warn('MCP server process already running, skipping restart');
    return;
  }

  const childEnv = {
    ...process.env,
    PORT: String(MCP_PORT),
    HOST: MCP_HOST,
    AUTH_TOKEN,
  };

  const args = ['--yes', MCP_PACKAGE, ...MCP_EXTRA_ARGS, AUTH_TOKEN];

  console.log(
    `Launching TypingMind MCP server via ${NPX_COMMAND} ${[
      MCP_PACKAGE,
      ...MCP_EXTRA_ARGS,
    ].join(' ')} on ${childEnv.HOST}:${childEnv.PORT}`,
  );

  mcpServerHealthy = false;
  lastHealthCheck = null;

  mcpServerProcess = spawn(NPX_COMMAND, args, {
    stdio: ['ignore', 'pipe', 'pipe'],
    env: childEnv,
  });

  mcpServerProcess.stdout.on('data', (data) => {
    const message = data.toString().trim();
    if (message) {
      console.log('[MCP Server]:', message);
    }
    mcpServerHealthy = true;
    lastHealthCheck = new Date().toISOString();
  });

  mcpServerProcess.stderr.on('data', (data) => {
    const message = data.toString().trim();
    if (message) {
      console.error('[MCP Server Error]:', message);
    }
  });

  mcpServerProcess.on('error', (error) => {
    console.error('Failed to start MCP server:', error);
    mcpServerHealthy = false;
    mcpServerProcess = null;
    scheduleRestart();
  });

  mcpServerProcess.on('exit', (code, signal) => {
    console.warn(`MCP server exited with code ${code} and signal ${signal}`);
    mcpServerHealthy = false;
    mcpServerProcess = null;

    if (code === 0) {
      console.log('MCP server exited normally, not scheduling restart');
      return;
    }

    scheduleRestart();
  });
}

function shutdown(signal) {
  console.log(`${signal} received, shutting down gracefully...`);
  stopMCPServer(signal);
  server.close(() => {
    console.log('Health server closed');
    process.exit(0);
  });
}

process.on('SIGTERM', () => shutdown('SIGTERM'));
process.on('SIGINT', () => shutdown('SIGINT'));
process.on('exit', () => stopMCPServer('SIGTERM'));

// Start both servers
const server = app.listen(HEALTH_PORT, '0.0.0.0', () => {
  console.log(`Health check server running on port ${HEALTH_PORT}`);
  console.log(
    `Environment: HEALTH_PORT=${HEALTH_PORT}, MCP_PORT=${MCP_PORT}, MCP_HOST=${MCP_HOST}, AUTH_TOKEN=${AUTH_TOKEN ? 'SET' : 'NOT SET'}`,
  );
  startMCPServer();
});

console.log('TypingMind MCP Server Wrapper initialized');
console.log('This wrapper provides /health endpoint for Render while running the MCP server');
