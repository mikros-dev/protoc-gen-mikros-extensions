package protobuf

import (
	"fmt"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
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
	Comment      Comment
	Proto        *descriptor.MethodDescriptorProto
}

func parseMethod(proto *descriptor.MethodDescriptorProto, schema *protogen.Method) *Method {
	var (
		httpMethod string
		endpoint   string
	)

	if googleAPI := getGoogleHTTPAPIIfAny(proto); googleAPI != nil {
		httpMethod, endpoint = getMethodAndEndpoint(googleAPI)
	}

	return &Method{
		Name:         proto.GetName(),
		RequestType:  protoName(proto.GetInputType()),
		ResponseType: protoName(proto.GetOutputType()),
		HTTPMethod:   httpMethod,
		Endpoint:     endpoint,
		Comment:      parseMethodComment(schema),
		Proto:        proto,
	}
}

// getGoogleHTTPAPIIfAny gets the google.api.http extension of a method if exists.
func getGoogleHTTPAPIIfAny(msg *descriptor.MethodDescriptorProto) *annotations.HttpRule {
	if msg.Options != nil {
		if h := proto.GetExtension(msg.Options, annotations.E_Http); h != nil {
			return h.(*annotations.HttpRule)
		}
	}

	return nil
}

// getMethodAndEndpoint translates a google.api.http notation of a request
// type to our supported type.
func getMethodAndEndpoint(rule *annotations.HttpRule) (string, string) {
	var (
		method   = ""
		endpoint = ""
	)

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

func parseMethodComment(method *protogen.Method) Comment {
	if method == nil {
		return Comment{}
	}

	detached := make([]string, 0, len(method.Comments.LeadingDetached))
	for _, c := range method.Comments.LeadingDetached {
		detached = append(detached, string(c))
	}

	return Comment{
		Leading:         string(method.Comments.Leading),
		Trailing:        string(method.Comments.Trailing),
		LeadingDetached: detached,
	}
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
