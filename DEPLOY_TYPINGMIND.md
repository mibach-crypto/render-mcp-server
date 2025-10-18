# Deploying Render MCP Server for TypingMind

This guide walks you through deploying the Render MCP Server to work with TypingMind.

## Prerequisites

1. A [Render account](https://render.com/register)
2. A [Render API key](https://dashboard.render.com/account/api-keys)
3. A TypingMind account with MCP support

## Deployment Options

### Option 1: Deploy to Render (Recommended)

#### Step 1: Fork or Use This Repository
- Fork this repository to your GitHub account, or
- Use this repository directly: `mibach-crypto/render-mcp-server`

#### Step 2: Create a New Web Service on Render

1. Go to [Render Dashboard](https://dashboard.render.com/)
2. Click "New +" → "Web Service"
3. Connect your GitHub account and select this repository
4. Configure the service:
   - **Name**: `mcp-server` (or your preferred name)
   - **Environment**: `Docker`
   - **Branch**: `main` or `genspark_ai_developer`
   - **Instance Type**: `Free` (for testing) or `Starter` (for production)
   - **Region**: Choose closest to you

#### Step 3: Set Environment Variables

In Render dashboard, add these environment variables:

| Variable | Value | Description |
|----------|-------|-------------|
| `AUTH_TOKEN` | `<generate-secure-token>` | Authentication token for TypingMind (e.g., use `openssl rand -hex 32`) |
| `RENDER_API_KEY` | `<your-render-api-key>` | Your Render API key from account settings |
| `REDIS_URL` | `<optional-redis-url>` | (Optional) Redis connection string for persistent sessions |

#### Step 4: Deploy

1. Click "Create Web Service"
2. Wait for the build and deployment to complete (usually 2-5 minutes)
3. Your service URL will be: `https://your-service-name.onrender.com`

### Option 2: Deploy Using render.yaml (Blueprint)

This repository includes a `render.yaml` file for automated deployment.

1. Go to [Render Dashboard](https://dashboard.render.com/)
2. Click "New +" → "Blueprint"
3. Connect this repository
4. Render will automatically detect the `render.yaml` file
5. Set the required environment variables:
   - `AUTH_TOKEN`: Your secure authentication token
   - `RENDER_API_KEY`: Your Render API key
6. Click "Apply"

This will create:
- A web service running the MCP server
- A Redis Key-Value store for persistent sessions

### Option 3: Local Development/Testing

For local testing before deployment:

```bash
# Clone the repository
git clone https://github.com/mibach-crypto/render-mcp-server.git
cd render-mcp-server

# Copy and configure environment variables
cp .env.example .env
# Edit .env with your values

# Build the Go binary
go build -o render-mcp-server

# Run the server
AUTH_TOKEN=your_token_here \
RENDER_API_KEY=your_api_key \
PORT=8080 \
./render-mcp-server --transport http
```

## Configuring TypingMind

Once your server is deployed:

1. Open TypingMind
2. Go to MCP Connector settings
3. Select "Remote" as the server location
4. Enter your configuration:
   - **Server URL**: 
     - For Render: `https://your-service-name.onrender.com/mcp`
     - For local: `http://localhost:8080/mcp`
   - **Authentication Token**: The same `AUTH_TOKEN` you configured

5. Click "Connect" or "Save"

## Verifying the Connection

### Health Check
Visit `https://your-service-name.onrender.com/health` in your browser. You should see:
```json
{
  "status": "healthy",
  "transport": "http",
  "version": "..."
}
```

### In TypingMind
After connecting, you should be able to:
- List your Render workspaces
- View and manage services
- Query PostgreSQL databases
- Check logs and metrics
- Create new services and databases

## Available MCP Tools

Once connected, TypingMind can use these tools:

### Service Management
- `list_services` - List all your Render services
- `get_service` - Get details about a specific service
- `create_web_service` - Create a new web service
- `create_static_site` - Create a new static site
- `update_environment_variables` - Update service environment variables

### Database Operations
- `list_postgres_instances` - List all PostgreSQL databases
- `get_postgres` - Get database details
- `create_postgres` - Create a new PostgreSQL database
- `query_render_postgres` - Run read-only SQL queries

### Monitoring
- `list_logs` - View application logs
- `get_metrics` - Get performance metrics
- `list_deploys` - View deployment history

### Workspaces
- `list_workspaces` - List available workspaces
- `select_workspace` - Switch between workspaces

## Security Considerations

1. **AUTH_TOKEN**: 
   - Use a strong, random token (e.g., `openssl rand -hex 32`)
   - Never commit this to version control
   - Rotate regularly

2. **RENDER_API_KEY**:
   - Keep this secret
   - Use Render's API key scoping if available
   - Monitor API key usage in Render dashboard

3. **Network Security**:
   - The server validates the AUTH_TOKEN on every request
   - All communication should happen over HTTPS
   - Consider IP whitelisting if needed

## Troubleshooting

### Server won't start
- Check logs in Render dashboard
- Verify environment variables are set correctly
- Ensure Dockerfile builds successfully

### TypingMind can't connect
- Verify the server URL ends with `/mcp`
- Check AUTH_TOKEN matches exactly
- Test the health endpoint manually
- Check Render service logs for connection attempts

### Redis connection issues
- REDIS_URL is optional; server works without it
- If using Redis, ensure the connection string is correct
- Check Redis instance is running and accessible

### Query limitations
- PostgreSQL queries are read-only for security
- Large result sets may be truncated
- Complex queries may timeout

## Performance Tips

1. **Use Redis** for persistent sessions in production
2. **Upgrade instance type** if you experience slow responses
3. **Monitor metrics** to identify bottlenecks
4. **Set appropriate timeouts** for database queries

## Support

- **Issues**: [GitHub Issues](https://github.com/mibach-crypto/render-mcp-server/issues)
- **Render Docs**: [render.com/docs](https://render.com/docs)
- **MCP Spec**: [modelcontextprotocol.io](https://modelcontextprotocol.io)

## Updates and Maintenance

To update your deployed server:

1. Pull latest changes: `git pull origin main`
2. Render will auto-deploy if you have auto-deploy enabled
3. Or manually trigger a deployment in Render dashboard

Remember to check the [CHANGELOG](./CHANGELOG.md) for breaking changes before updating.