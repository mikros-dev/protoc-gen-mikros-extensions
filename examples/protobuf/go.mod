module examples

replace github.com/mikros-dev/protoc-gen-mikros-extensions => ./../..

go 1.23.0

require (
	github.com/fasthttp/router v1.5.2
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	github.com/mikros-dev/protoc-gen-mikros-extensions v0.0.0-00010101000000-000000000000
	github.com/valyala/fasthttp v1.55.0
	google.golang.org/genproto/googleapis/api v0.0.0-20250324211829-b45e905df463
	google.golang.org/grpc v1.71.0
	google.golang.org/protobuf v1.36.6
)

require (
	github.com/andybalholm/brotli v1.1.0 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/savsgio/gotils v0.0.0-20240704082632-aef3928b8a38 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250313205543-e70fdf4c4cb4 // indirect
)
