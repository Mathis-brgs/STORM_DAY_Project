FROM golang:1.25-alpine AS build
WORKDIR /app
COPY . .
WORKDIR /app/services/notification
RUN CGO_ENABLED=0 go build -o /services/notification/cmd/main.go ./cmd/main.go

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=build /services/notification/cmd/main.go /services/notification/cmd/main.go
EXPOSE 8080
CMD ["/services/notification/cmd/main.go"]
