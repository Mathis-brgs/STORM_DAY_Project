FROM golang:1.25-alpine AS build
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o /cmd/main ./cmd/main.go

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=build /cmd/main /cmd/main
EXPOSE 8080
CMD ["/cmd/main"]
