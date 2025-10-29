package protobuf

import (
	"fmt"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	descriptor "google.golang.org/protobuf/types/descriptorpb"
)

// Method represents a method loaded from a protobuf service.
type Method struct {
	Name         string
	RequestType  *ProtoName
	ResponseType *ProtoName
	HTTPMethod   string
	Endpoint     string
	Proto        *descriptor.MethodDescriptorProto
}

func parseMethod(method *descriptor.MethodDescriptorProto) *Method {
	var (
		httpMethod string
		endpoint   string
	)

	if googleAPI := getGoogleHTTPAPIIfAny(method); googleAPI != nil {
		httpMethod, endpoint = getMethodAndEndpoint(googleAPI)
	}

	return &Method{
		Name:         method.GetName(),
		RequestType:  protoName(method.GetInputType()),
		ResponseType: protoName(method.GetOutputType()),
		HTTPMethod:   httpMethod,
		Endpoint:     endpoint,
		Proto:        method,
	}
}

// getGoogleHTTPAPIIfAny gets the google.api.http extension of a method if exists.
func getGoogleHTTPAPIIfAny(msg *descriptor.MethodDescriptorProto) *annotations.HttpRule {
	if msg.Options != nil {
		h := proto.GetExtension(msg.Options, annotations.E_Http)
		return h.(*annotations.HttpRule)
	}

	return nil
}

// getMethodAndEndpoint translates a google.api.http notation of a request
// type to our supported type.
func getMethodAndEndpoint(rule *annotations.HttpRule) (string, string) {
	method := ""
	endpoint := ""

	switch rule.GetPattern().(type) {
	case *annotations.HttpRule_Get:
		method = "GET"
		endpoint = rule.GetGet()

	case *annotations.HttpRule_Post:
		method = "POST"
		endpoint = rule.GetPost()

	case *annotations.HttpRule_Put:
		method = "PUT"
		endpoint = rule.GetPut()

	case *annotations.HttpRule_Delete:
		method = "DELETE"
		endpoint = rule.GetDelete()

	case *annotations.HttpRule_Patch:
		method = "PATCH"
		endpoint = rule.GetPatch()
	}

	return method, endpoint
}

// HasHTTPBody returns true if the method has a body.
func (m *Method) HasHTTPBody() bool {
	return m.HTTPMethod == "POST" || m.HTTPMethod == "PUT"
}

func (m *Method) String() string {
	s := fmt.Sprintf(`{name:%v, request:%v, response:%v`,
		m.Name,
		m.RequestType,
		m.ResponseType)

	if m.HTTPMethod != "" {
		s += fmt.Sprintf(`, type:%v, endpoint:%v`,
			m.HTTPMethod,
			m.Endpoint)
	}

	return s + "}"
}
