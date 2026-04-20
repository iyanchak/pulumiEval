# ---------- Build Stage ----------
FROM golang:1.25-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git

COPY . .

# 1. FORCE CGO_ENABLED=0 (Creates a standalone binary that needs no libraries)
RUN CGO_ENABLED=0 go build -o /app/counter ./*.go

# ---------- Final Stage ----------
FROM alpine:latest
RUN apk add --no-cache ca-certificates && update-ca-certificates

COPY --from=builder /app/counter /counter

# 2. EXPLICITLY set permissions so Kubernetes can run it
RUN chmod +x /counter

# Create the data directory and give it permissions
RUN mkdir -p /data && chmod 777 /data

# Set the working directory
WORKDIR /data
ENTRYPOINT ["/counter"]