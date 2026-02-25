FROM golang:1.25-alpine AS build
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o /app/media-service ./cmd/main.go

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=build /app/media-service /media-service
EXPOSE 8080
CMD ["/media-service"]
