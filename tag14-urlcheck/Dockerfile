# Build-Stage: kompiliert ein statisches Binary
FROM golang:1.26.9-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -trimpath \
    -ldflags "-s -w -X main.version=${VERSION}" -o /out/urlcheck ./cmd/urlcheck

# Laufzeit: nur das Binary, CA-Zertifikate und ein nonroot-User
FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/urlcheck /urlcheck
USER nonroot:nonroot
ENTRYPOINT ["/urlcheck"]
