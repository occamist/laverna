FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/laverna .

FROM cgr.dev/chainguard/static:latest

COPY --from=builder /out/laverna /usr/local/bin/laverna

EXPOSE 8770

ENTRYPOINT ["/usr/local/bin/laverna"]
CMD ["yomitan", "--host", "0.0.0.0", "--port", "8770", "--https"]
