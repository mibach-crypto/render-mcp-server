# syntax=docker/dockerfile:1
FROM golang:1.23-alpine AS build

ARG VERSION="dev"

WORKDIR /build

# Install git for private module support
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 \
    go build -ldflags="-s -w -X github.com/render-oss/render-mcp-server/pkg/cfg.Version=${VERSION}" \
    -o /bin/render-mcp-server main.go

FROM gcr.io/distroless/base-debian12

WORKDIR /server

COPY --from=build /bin/render-mcp-server ./render-mcp-server

ENV RENDER_CONFIG_PATH=/config/mcp-server.yaml

EXPOSE 10000

ENTRYPOINT ["/server/render-mcp-server"]
CMD ["--transport", "http"]
