module github.com/nomuken/william/services/server

go 1.24.0

require (
	connectrpc.com/connect v1.13.0
	github.com/golang-migrate/migrate/v4 v4.18.3
	github.com/lib/pq v1.10.9
	google.golang.org/protobuf v1.36.0
)

require (
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	go.uber.org/atomic v1.7.0 // indirect
)

replace github.com/nomuken/william => ../..
