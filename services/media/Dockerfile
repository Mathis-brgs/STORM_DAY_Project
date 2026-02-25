FROM golang:1.25-alpine AS build
WORKDIR /app
COPY . .
WORKDIR /app/services/media
RUN CGO_ENABLED=0 go build -o media-service ./cmd/main.go

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=build /app/services/media/media-service /media-service
EXPOSE 8080
CMD ["/media-service"]
