# Build context must be repo root so replace github.com/.../message => ../message resolves.
FROM golang:1.25-alpine AS build
WORKDIR /app
COPY services/gateway/ ./
COPY services/message /message
RUN go mod download
RUN CGO_ENABLED=0 go build -o /app/gateway-service ./cmd/main.go

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=build /app/gateway-service /gateway-service
EXPOSE 8080
CMD ["/gateway-service"]
