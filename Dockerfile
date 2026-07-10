# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/laverna .

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/laverna /usr/local/bin/laverna

USER nonroot:nonroot

EXPOSE 8770

ENTRYPOINT ["/usr/local/bin/laverna"]
CMD ["yomitan", "--host", "0.0.0.0", "--port", "8770", "--https"]
