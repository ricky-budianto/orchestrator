FROM golang:1.25-alpine3.23 AS builder

RUN apk update && apk add --no-cache git

# Prepare Git

RUN --mount=type=secret,id=GITHUB_USERNAME \
    --mount=type=secret,id=GITHUB_TOKEN \
    export GITHUB_USERNAME=$(cat /run/secrets/GITHUB_USERNAME) && \
    export GITHUB_TOKEN=$(cat /run/secrets/GITHUB_TOKEN) && \
    git config --global url."https://${GITHUB_USERNAME}:${GITHUB_TOKEN}@github.com".insteadOf "https://github.com"

# Move to working directory /build
WORKDIR /build

# Copy and download dependency using go mod
COPY go.mod go.sum ./

RUN go mod download

# Copy the code into the container
COPY . .

# Build the application with telemetry support
RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o main .
# RUN upx --best --lzma main

# Move to /dist directory as the place for resulting binary folder
WORKDIR /dist

# Copy binary from build to main folder
RUN cp  /build/default.env ./config.env
RUN cp /build/main .
# RUN cp /build/assets .

# Build a small image
FROM gcr.io/distroless/static-debian11

#perpare user and tz
USER nonroot:nonroot
ENV TZ=Asia/Jakarta

# copy application
COPY --from=builder /dist/main /dist/config.env /

# Export necessary ports
EXPOSE 3000
# Metrics endpoint exposed on the same port as the main service

# Command to run
ENTRYPOINT ["/main"]