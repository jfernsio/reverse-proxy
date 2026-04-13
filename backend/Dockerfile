FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy dependency files first (caching = faster rebuilds)
COPY go.mod ./
#download depndecies
RUN go mod download
#copy source code all filess
COPY . .

#build bonary
RUN CGO_ENABLED=0 GOOS=linux \
    go build \
    -ldflags="-s -w" \
    -trimpath \
    -o /backend-servers .

#runtime stage

FROM alpine:latest

# Add curl for healthcheck
RUN apk add --no-cache curl

#copy binary from builder stage
COPY --from=builder /backend-servers /backend-servers

#create non-root user for security
RUN adduser -D -u 10001 appuser
USER appuser

#expose port 
EXPOSE 9000 9001

#healthcheck to ensure server is running
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:9000/ || exit 1
  
#start the binary server
ENTRYPOINT ["/backend-servers"]    
