FROM golang:1.25-alpine AS build
WORKDIR /app
COPY . .
WORKDIR /app/services/message
RUN CGO_ENABLED=0 go build -o message-service ./cmd/message-service/main.go

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=build /app/services/message/message-service /message-service
EXPOSE 8080
CMD ["/message-service"]
