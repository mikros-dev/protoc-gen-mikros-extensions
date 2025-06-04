module examples

replace github.com/mikros-dev/protoc-gen-mikros-extensions => ./../..

go 1.23.0

require (
	github.com/mikros-dev/protoc-gen-mikros-extensions v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.71.0
	google.golang.org/protobuf v1.36.6
)

require (
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250324211829-b45e905df463 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250313205543-e70fdf4c4cb4 // indirect
)
