FROM golang:1.25-alpine AS build
WORKDIR /app

# Copy the entire project context
COPY . .

# Build from the gateway service directory
WORKDIR /app/services/gateway
RUN CGO_ENABLED=0 go build -o gateway-service ./cmd/main.go

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
# Copy the binary from the build stage (relative to the WORKDIR of the build stage)
COPY --from=build /app/services/gateway/gateway-service /gateway-service
EXPOSE 8080
CMD ["/gateway-service"]
