FROM node:20-alpine

WORKDIR /app

# Install npx (comes with node)
# No build step needed - npx will download @typingmind/mcp on first run

# The auth token is passed as the first argument
CMD npx @typingmind/mcp ${AUTH_TOKEN}
