package extensions

import (
	"regexp"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	descriptor "google.golang.org/protobuf/types/descriptorpb"
)

func LoadGoogleAnnotations(method *descriptor.MethodDescriptorProto) *annotations.HttpRule {
	if method.Options != nil {
		h := proto.GetExtension(method.Options, annotations.E_Http)
		return h.(*annotations.HttpRule)
	}

	return nil
}

func GetHttpEndpoint(rule *annotations.HttpRule) (string, string) {
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

func RetrieveParameters(endpoint string) []string {
	var parameters []string
	re := regexp.MustCompile(`{[A-Za-z_.0-9]*}`)

	for _, p := range re.FindAll([]byte(endpoint), -1) {
		parameters = append(parameters, string(p[1:len(p)-1]))
	}

	return parameters
}

func RetrieveParametersFromAdditionalBindings(rule *annotations.HttpRule) []string {
	var parameters []string

	for _, r := range rule.GetAdditionalBindings() {
		if endpoint, _ := GetHttpEndpoint(r); endpoint != "" {
			parameters = append(parameters, RetrieveParameters(endpoint)...)
		}
	}

	return parameters
}

func LoadMethodDefinitions(method *descriptor.MethodDescriptorProto) *HttpMethodExtensions {
	if method.Options != nil {
		v := proto.GetExtension(method.Options, E_Http)
		if val, ok := v.(*HttpMethodExtensions); ok {
			return val
		}
	}

	return nil
}

func LoadEnumDecodingOptions(enum *descriptor.EnumDescriptorProto) *EnumApiExtensions {
	if enum.Options != nil {
		v := proto.GetExtension(enum.Options, E_Api)
		if d, ok := v.(*EnumApiExtensions); ok {
			return d
		}
	}

	return nil
}

func LoadEnumEntry(enumValue *descriptor.EnumValueDescriptorProto) *EnumEntry {
	if enumValue.Options != nil {
		v := proto.GetExtension(enumValue.Options, E_Entry)
		if e, ok := v.(*EnumEntry); ok {
			return e
		}
	}

	return nil
}

func LoadFieldDomain(field *descriptor.FieldDescriptorProto) *FieldDomainOptions {
	if field.Options != nil {
		v := proto.GetExtension(field.Options, E_Domain)
		if val, ok := v.(*FieldDomainOptions); ok {
			return val
		}
	}

	// The field does not have extensions annotations.
	return nil
}

func LoadFieldDatabase(field *descriptor.FieldDescriptorProto) *FieldDatabaseOptions {
	if field.Options != nil {
		v := proto.GetExtension(field.Options, E_Database)
		if val, ok := v.(*FieldDatabaseOptions); ok {
			return val
		}
	}

	// The field does not have extensions annotations.
	return nil
}

func LoadFieldInbound(field *descriptor.FieldDescriptorProto) *FieldInboundOptions {
	if field.Options != nil {
		v := proto.GetExtension(field.Options, E_Inbound)
		if val, ok := v.(*FieldInboundOptions); ok {
			return val
		}
	}

	// The field does not have extensions annotations.
	return nil
}

func LoadFieldOutbound(field *descriptor.FieldDescriptorProto) *FieldOutboundOptions {
	if field.Options != nil {
		v := proto.GetExtension(field.Options, E_Outbound)
		if val, ok := v.(*FieldOutboundOptions); ok {
			return val
		}
	}

	// The field does not have extensions annotations.
	return nil
}

func LoadFieldValidate(field *descriptor.FieldDescriptorProto) *FieldValidateOptions {
	if field.Options != nil {
		v := proto.GetExtension(field.Options, E_Validate)
		if val, ok := v.(*FieldValidateOptions); ok {
			return val
		}
	}

	return nil
}

func LoadMessageDomainOptions(message *descriptor.DescriptorProto) *MessageDomainExpansionOptions {
	if message.Options != nil {
		v := proto.GetExtension(message.Options, E_DomainExpansion)
		if val, ok := v.(*MessageDomainExpansionOptions); ok {
			return val
		}
	}

	return nil
}

func LoadMessageWireExtensionOptions(message *descriptor.DescriptorProto) *MessageWireExpansionOptions {
	if message.Options != nil {
		v := proto.GetExtension(message.Options, E_WireExpansion)
		if val, ok := v.(*MessageWireExpansionOptions); ok {
			return val
		}
	}

	return nil
}

func LoadMessageInboundOptions(message *descriptor.DescriptorProto) *MessageInboundOptions {
	if message.Options != nil {
		v := proto.GetExtension(message.Options, E_InboundOptions)
		if val, ok := v.(*MessageInboundOptions); ok {
			return val
		}
	}

	return nil
}

func LoadMessageOutboundOptions(message *descriptor.DescriptorProto) *MessageOutboundOptions {
	if message.Options != nil {
		v := proto.GetExtension(message.Options, E_OutboundOptions)
		if val, ok := v.(*MessageOutboundOptions); ok {
			return val
		}
	}

	return nil
}

func LoadMethodHttpExtensionOptions(method *descriptor.MethodDescriptorProto) *HttpMethodExtensions {
	if method.Options != nil {
		v := proto.GetExtension(method.Options, E_Http)
		if val, ok := v.(*HttpMethodExtensions); ok {
			return val
		}
	}

	return nil
}
