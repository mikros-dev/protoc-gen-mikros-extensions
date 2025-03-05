package context

import (
	"strings"

	descriptor "google.golang.org/protobuf/types/descriptorpb"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/mikros/extensions"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf"
)

type Endpoint struct {
	Body           string
	Path           string
	Method         string
	Parameters     []string
	HttpExtensions *mikros_extensions.HttpMethodExtensions
}

func getEndpoint(method *protobuf.Method) *Endpoint {
	googleHttp := mikros_extensions.LoadGoogleAnnotations(method.Proto)
	if googleHttp == nil {
		return nil
	}

	e := &Endpoint{
		Body: googleHttp.GetBody(),
	}

	if endpoint, m := mikros_extensions.GetHttpEndpoint(googleHttp); endpoint != "" {
		e.Path = endpoint
		e.Method = m
		e.Parameters = mikros_extensions.RetrieveParameters(endpoint)
		e.Parameters = append(e.Parameters, mikros_extensions.RetrieveParametersFromAdditionalBindings(googleHttp)...)
	}

	m := mikros_extensions.LoadMethodExtensions(method.Proto)
	if m != nil {
		if op := m.GetHttp(); op != nil {
			e.HttpExtensions = op
		}
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
