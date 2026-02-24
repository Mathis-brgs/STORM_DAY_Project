FROM golang:1.25-alpine AS build
WORKDIR /app
COPY . .
WORKDIR /app/services/notification
RUN CGO_ENABLED=0 go build -o notification-service ./cmd/main.go

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=build /app/services/notification/notification-service /notification-service
EXPOSE 8080
CMD ["/notification-service"]
