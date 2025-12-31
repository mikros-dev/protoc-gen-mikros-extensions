//go:generate protoc -I . --go_out=pkg/protobuf/extensions --go_opt=module=github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf/extensions proto/mikros_extensions.proto
package main
