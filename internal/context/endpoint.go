package context

import (
	"strings"

	descriptor "google.golang.org/protobuf/types/descriptorpb"

	"github.com/rsfreitas/protoc-gen-mikros-extensions/internal/protobuf"
	"github.com/rsfreitas/protoc-gen-mikros-extensions/mikros/extensions"
)

type Endpoint struct {
	Body           string
	Path           string
	Method         string
	Parameters     []string
	HttpExtensions *extensions.HttpMethodExtensions
}

func getEndpoint(method *protobuf.Method) *Endpoint {
	googleHttp := extensions.LoadGoogleAnnotations(method.Proto)
	if googleHttp == nil {
		return nil
	}

	e := &Endpoint{
		Body: googleHttp.GetBody(),
	}

	if endpoint, m := extensions.GetHttpEndpoint(googleHttp); endpoint != "" {
		e.Path = endpoint
		e.Method = m
		e.Parameters = extensions.RetrieveParameters(endpoint)
		e.Parameters = append(e.Parameters, extensions.RetrieveParametersFromAdditionalBindings(googleHttp)...)
	}

	if op := extensions.LoadMethodDefinitions(method.Proto); op != nil {
		e.HttpExtensions = op
	}

	return e
}

func getFieldLocation(field *descriptor.FieldDescriptorProto, endpoint *Endpoint) FieldLocation {
	if endpoint == nil {
		return FieldLocation_Body
	}

	if isEndpointParameter(field.GetName(), endpoint) {
		return FieldLocation_Path
	}

	if isHeaderParameter(field.GetName(), endpoint) {
		return FieldLocation_Header
	}

	if strings.Contains(endpoint.Body, field.GetName()) || endpoint.Body == "*" {
		return FieldLocation_Body
	}

	return FieldLocation_Query
}

func isEndpointParameter(name string, endpoint *Endpoint) bool {
	if endpoint != nil {
		for _, p := range endpoint.Parameters {
			if p == name {
				return true
			}
		}
	}

	return false
}

func isHeaderParameter(name string, endpoint *Endpoint) bool {
	if endpoint != nil && endpoint.HttpExtensions != nil {
		for _, n := range endpoint.HttpExtensions.GetHeader() {
			if name == n {
				return true
			}
		}
	}

	return false
}
