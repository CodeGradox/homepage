# syntax=docker/dockerfile:1

# Build stage: compile a static binary with all assets embedded.
ARG GO_VERSION=1.26
FROM golang:${GO_VERSION}-alpine AS build

WORKDIR /src

# Download modules first so this layer caches independently of source changes.
# (This project is stdlib-only, so it's essentially a no-op, but keeps the
# build correct if dependencies are ever added.)
COPY go.mod go.sum* ./
RUN go mod download

COPY . .

# CGO off so the binary is fully static and runs on a scratch/distroless base.
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /homepage .

# Final stage: a minimal, non-root image. distroless/static ships CA certs and a
# nonroot user but nothing else — no shell, no package manager.
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /homepage /homepage

EXPOSE 8080
ENV PORT=8080
USER nonroot:nonroot

ENTRYPOINT ["/homepage"]
