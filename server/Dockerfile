# Build context is server/ (see render.yaml). Two stages: build the API with
# the Go toolchain, then copy only the binary into a small image.

# Must be at least the `go` line in go.mod (1.25.4).
FROM golang:1.25-alpine AS build
WORKDIR /src

# Download modules first, so code changes don't redo this layer.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# CGO off gives a static binary that runs without libc in the distroless image.
# -trimpath and -s -w drop local paths and debug info to keep it small.
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

# distroless/static has CA certificates (needed for the TLS link to Atlas) and
# no shell or package manager. :nonroot runs the app as an unprivileged user.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/api /api

# Render sets PORT itself; 8080 is only the fallback when it doesn't.
EXPOSE 8080
ENTRYPOINT ["/api"]
