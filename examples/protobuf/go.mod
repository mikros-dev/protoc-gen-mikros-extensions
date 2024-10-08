module examples

replace github.com/mikros-dev/protoc-gen-mikros-extensions => ./../..

go 1.22.4

toolchain go1.22.6

require (
	github.com/fasthttp/router v1.5.2
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	github.com/mikros-dev/protoc-gen-mikros-extensions v0.0.0-00010101000000-000000000000
	github.com/valyala/fasthttp v1.55.0
	google.golang.org/genproto/googleapis/api v0.0.0-20240903143218-8af14fe29dc1
	google.golang.org/grpc v1.66.2
	google.golang.org/protobuf v1.34.2
)

require (
	github.com/andybalholm/brotli v1.1.0 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/savsgio/gotils v0.0.0-20240704082632-aef3928b8a38 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240827150818-7e3bb234dfed // indirect
)
