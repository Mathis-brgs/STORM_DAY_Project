module gateway

go 1.25.6

replace github.com/Mathis-brgs/storm-project/services/message => ../message

require (
	github.com/Mathis-brgs/storm-project/services/message v0.0.0
	github.com/go-chi/chi/v5 v5.2.5
	github.com/golang/protobuf v1.5.4
	github.com/lxzan/gws v1.8.9
	github.com/nats-io/nats.go v1.48.0
)

require (
	github.com/dolthub/maphash v0.1.0 // indirect
	github.com/klauspost/compress v1.18.3 // indirect
	github.com/nats-io/nkeys v0.4.11 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
