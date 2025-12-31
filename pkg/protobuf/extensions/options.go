package extensions

import (
	"regexp"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	descriptor "google.golang.org/protobuf/types/descriptorpb"
)

// LoadGoogleAnnotations loads the HTTP rule from the method options.
func LoadGoogleAnnotations(method *descriptor.MethodDescriptorProto) *annotations.HttpRule {
	if method.Options != nil {
		h := proto.GetExtension(method.Options, annotations.E_Http)
		return h.(*annotations.HttpRule)
	}

	return nil
}

// GetHTTPEndpoint returns the endpoint and method from the HTTP rule.
func GetHTTPEndpoint(rule *annotations.HttpRule) (string, string) {
	var (
		endpoint string
		method   string
	)

	switch rule.GetPattern().(type) {
	case *annotations.HttpRule_Get:
		endpoint = rule.GetGet()
		method = "GET"

	case *annotations.HttpRule_Post:
		endpoint = rule.GetPost()
		method = "POST"

	case *annotations.HttpRule_Put:
		endpoint = rule.GetPut()
		method = "PUT"

	case *annotations.HttpRule_Delete:
		endpoint = rule.GetDelete()
		method = "DELETE"

	case *annotations.HttpRule_Patch:
		endpoint = rule.GetPatch()
		method = "PATCH"
	}

	return endpoint, method
}

// RetrieveParameters returns the parameters from the endpoint.
func RetrieveParameters(endpoint string) []string {
	var parameters []string
	re := regexp.MustCompile(`{[A-Za-z_.0-9]*}`)

	for _, p := range re.FindAll([]byte(endpoint), -1) {
		parameters = append(parameters, string(p[1:len(p)-1]))
	}

	return parameters
}

// RetrieveParametersFromAdditionalBindings returns the parameters from the
// additional bindings of an HTTP rule.
func RetrieveParametersFromAdditionalBindings(rule *annotations.HttpRule) []string {
	var parameters []string

	for _, r := range rule.GetAdditionalBindings() {
		if endpoint, _ := GetHTTPEndpoint(r); endpoint != "" {
			parameters = append(parameters, RetrieveParameters(endpoint)...)
		}
	}

	return parameters
}

// LoadMethodExtensions loads the Mikros extensions from the method options.
func LoadMethodExtensions(method *descriptor.MethodDescriptorProto) *MikrosMethodExtensions {
	if method.Options != nil {
		v := proto.GetExtension(method.Options, E_MethodOptions)
		if val, ok := v.(*MikrosMethodExtensions); ok {
			return val
		}
	}

	return nil
}

// LoadEnumExtensions loads the Mikros extensions from the enum options.
func LoadEnumExtensions(enum *descriptor.EnumDescriptorProto) *MikrosEnumExtensions {
	if enum.Options != nil {
		v := proto.GetExtension(enum.Options, E_EnumOptions)
		if d, ok := v.(*MikrosEnumExtensions); ok {
			return d
		}
	}

	return nil
}

// LoadEnumValueExtensions loads the Mikros extensions from the enum value options.
func LoadEnumValueExtensions(enumValue *descriptor.EnumValueDescriptorProto) *MikrosEnumValueExtensions {
	if enumValue.Options != nil {
		v := proto.GetExtension(enumValue.Options, E_EnumValueOptions)
		if e, ok := v.(*MikrosEnumValueExtensions); ok {
			return e
		}
	}

	return nil
}

// LoadFieldExtensions loads the Mikros extensions from the field options.
func LoadFieldExtensions(field *descriptor.FieldDescriptorProto) *MikrosFieldExtensions {
	if field.Options != nil {
		v := proto.GetExtension(field.Options, E_FieldOptions)
		if val, ok := v.(*MikrosFieldExtensions); ok {
			return val
		}
	}

	return nil
}

// LoadMessageExtensions loads the Mikros extensions from the message options.
func LoadMessageExtensions(message *descriptor.DescriptorProto) *MikrosMessageExtensions {
	if message.Options != nil {
		v := proto.GetExtension(message.Options, E_MessageOptions)
		if val, ok := v.(*MikrosMessageExtensions); ok {
			return val
		}
	}

	return nil
}

// LoadServiceExtensions loads the Mikros extensions from the service options.
func LoadServiceExtensions(service *descriptor.ServiceDescriptorProto) *MikrosServiceExtensions {
	if service.Options != nil {
		v := proto.GetExtension(service.Options, E_ServiceOptions)
		if val, ok := v.(*MikrosServiceExtensions); ok {
			return val
		}
	}

	return nil
}
