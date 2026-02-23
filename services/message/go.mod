module github.com/Mathis-brgs/storm-project/services/message

go 1.25.6

require (
	github.com/lib/pq v1.11.2
	github.com/nats-io/nats.go v1.48.0
)

replace github.com/Mathis-brgs/storm-project/pkg => ../../pkg

require (
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/nats-io/nkeys v0.4.11 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
