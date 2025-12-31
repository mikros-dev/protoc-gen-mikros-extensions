package context

import (
	"strings"

	descriptor "google.golang.org/protobuf/types/descriptorpb"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/mikros_extensions"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf"
)

// Endpoint represents an endpoint loaded from protobuf Google annotations.
type Endpoint struct {
	Body           string
	Path           string
	Method         string
	Parameters     []string
	HTTPExtensions *mikros_extensions.HttpMethodExtensions
}

func getEndpoint(method *protobuf.Method) *Endpoint {
	googleHTTP := mikros_extensions.LoadGoogleAnnotations(method.Proto)
	if googleHTTP == nil {
		return nil
	}

	e := &Endpoint{
		Body: googleHTTP.GetBody(),
	}

	if endpoint, m := mikros_extensions.GetHTTPEndpoint(googleHTTP); endpoint != "" {
		e.Path = endpoint
		e.Method = m
		e.Parameters = mikros_extensions.RetrieveParameters(endpoint)
		e.Parameters = append(e.Parameters, mikros_extensions.RetrieveParametersFromAdditionalBindings(googleHTTP)...)
	}

	m := mikros_extensions.LoadMethodExtensions(method.Proto)
	if m != nil {
		if op := m.GetHttp(); op != nil {
			e.HTTPExtensions = op
		}
	}

	return e
}

func getFieldLocation(field *descriptor.FieldDescriptorProto, endpoint *Endpoint) FieldLocation {
	if endpoint == nil {
		return FieldLocationBody
	}

	if isEndpointParameter(field.GetName(), endpoint) {
		return FieldLocationPath
	}

	if isHeaderParameter(field.GetName(), endpoint) {
		return FieldLocationHeader
	}

	if strings.Contains(endpoint.Body, field.GetName()) || endpoint.Body == "*" {
		return FieldLocationBody
	}

	return FieldLocationQuery
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
	if endpoint != nil && endpoint.HTTPExtensions != nil {
		for _, n := range endpoint.HTTPExtensions.GetHeader() {
			if name == n {
				return true
			}
		}
	}

	return false
}
