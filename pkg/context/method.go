package context

import (
	"fmt"
	"slices"
	"strings"

	"github.com/rsfreitas/protoc-gen-mikros-extensions/internal/converters"
	"github.com/rsfreitas/protoc-gen-mikros-extensions/internal/protobuf"
	"github.com/rsfreitas/protoc-gen-mikros-extensions/mikros/extensions"
)

type Method struct {
	Name                  string
	AdditionalHTTPMethods []HttpRule
	Request               *Message
	PathArguments         []*MethodField
	QueryArguments        []*MethodField
	HeaderArguments       []*MethodField

	endpoint *Endpoint
	service  *extensions.MikrosServiceExtensions
	method   *extensions.MikrosMethodExtensions
}

type HttpRule struct {
	Method   string
	Endpoint string
}

type MethodField struct {
	GoName    string
	ProtoName string
	CastType  string
}

func loadMethods(pkg *protobuf.Protobuf, messages []*Message) ([]*Method, error) {
	if pkg.Service == nil {
		return nil, nil
	}

	var (
		methods = make([]*Method, len(pkg.Service.Methods))
		service = extensions.LoadServiceExtensions(pkg.Service.Proto)
	)

	for i, method := range pkg.Service.Methods {
		var (
			msg              *Message
			endpoint         = getEndpoint(method)
			methodExtensions = extensions.LoadMethodExtensions(method.Proto)
		)

		index := slices.IndexFunc(messages, func(m *Message) bool {
			return m.Name == method.RequestType.Name && m.Type == converters.WireInputMessage
		})
		if index != -1 {
			msg = messages[index]
		}

		path, err := getPathArguments(msg, endpoint)
		if err != nil {
			return nil, err
		}

		header, err := getHeaderArguments(msg, methodExtensions)
		if err != nil {
			return nil, err
		}

		if err := validateBodyArguments(msg, endpoint); err != nil {
			return nil, err
		}

		m := &Method{
			Name:                  method.Name,
			AdditionalHTTPMethods: getAdditionalHttpRules(method),
			Request:               msg,
			PathArguments:         path,
			QueryArguments:        getQueryArguments(msg, endpoint, methodExtensions),
			HeaderArguments:       header,
			endpoint:              endpoint,
			service:               service,
			method:                methodExtensions,
		}
		if err := m.Validate(); err != nil {
			return nil, err
		}
		methods[i] = m
	}

	return methods, nil
}

func getPathArguments(m *Message, endpoint *Endpoint) ([]*MethodField, error) {
	var fields []*MethodField

	if endpoint != nil {
		for _, name := range endpoint.Parameters {
			index := slices.IndexFunc(m.Fields, func(f *Field) bool {
				return f.ProtoName == name
			})
			if index == -1 {
				return nil, fmt.Errorf("field '%s' declared in path arguments not found inside message '%s' definition", name, m.Name)
			}

			field := m.Fields[index]
			fields = append(fields, &MethodField{
				GoName:    field.GoName,
				ProtoName: field.ProtoName,
				CastType:  field.GoType,
			})
		}
	}

	return fields, nil
}

func getHeaderArguments(m *Message, methodExtensions *extensions.MikrosMethodExtensions) ([]*MethodField, error) {
	var fields []*MethodField

	if methodExtensions == nil {
		return fields, nil
	}

	if httpExtension := methodExtensions.GetHttp(); httpExtension != nil {
		for _, header := range httpExtension.GetHeader() {
			index := slices.IndexFunc(m.Fields, func(f *Field) bool {
				return f.ProtoName == header
			})
			if index == -1 {
				return nil, fmt.Errorf("header field '%s' not found inside message '%s' definition", header, m.Name)
			}

			field := m.Fields[index]
			fields = append(fields, &MethodField{
				GoName:    field.GoName,
				ProtoName: field.ProtoName,
				CastType:  field.GoType,
			})
		}
	}

	return fields, nil
}

func getQueryArguments(m *Message, endpoint *Endpoint, methodExtensions *extensions.MikrosMethodExtensions) []*MethodField {
	var fields []*MethodField

	if endpoint != nil {
		var (
			filteredParameters = getParametersToFilter(m, endpoint, methodExtensions)
			queryParameters    []string
		)

		for _, field := range m.Fields {
			index := slices.IndexFunc(filteredParameters, func(f string) bool {
				return f == field.ProtoName
			})
			if index == -1 {
				queryParameters = append(queryParameters, field.ProtoName)
			}
		}

		for _, p := range queryParameters {
			index := slices.IndexFunc(m.Fields, func(f *Field) bool {
				return f.ProtoName == p
			})
			if index != -1 {
				field := m.Fields[index]
				fields = append(fields, &MethodField{
					GoName:    field.GoName,
					ProtoName: field.ProtoName,
					CastType:  field.GoType,
				})
			}
		}
	}

	return fields
}

func getParametersToFilter(m *Message, endpoint *Endpoint, methodExtensions *extensions.MikrosMethodExtensions) []string {
	parameters := getBodyParameters(m, endpoint)

	if endpoint != nil {
		parameters = append(parameters, endpoint.Parameters...)
	}
	if methodExtensions != nil {
		if httpExtensions := methodExtensions.GetHttp(); httpExtensions != nil {
			parameters = append(parameters, httpExtensions.GetHeader()...)
		}
	}

	return parameters
}

func getBodyParameters(m *Message, endpoint *Endpoint) []string {
	var parameters []string

	if endpoint != nil {
		// All fields should be loaded from the body
		if endpoint.Body == "*" {
			for _, f := range m.Fields {
				parameters = append(parameters, f.ProtoName)
			}
		}
		if endpoint.Body != "*" && len(endpoint.Body) > 0 {
			parameters = append(parameters, strings.Split(endpoint.Body, " ")...)
		}
	}

	return parameters
}

func validateBodyArguments(m *Message, endpoint *Endpoint) error {
	// Checks if body parameters were declared inside the inbound message.
	if endpoint != nil && endpoint.Body != "*" && endpoint.Body != "" {
		parameters := strings.Split(endpoint.Body, " ")
		for _, param := range parameters {
			index := slices.IndexFunc(m.Fields, func(f *Field) bool {
				return f.ProtoName == param
			})
			if index == -1 {
				return fmt.Errorf("body field '%s' not found inside message '%s' definition", param, m.Name)
			}
		}
	}

	return nil
}

func getAdditionalHttpRules(method *protobuf.Method) []HttpRule {
	var rules []HttpRule

	if googleHttp := extensions.LoadGoogleAnnotations(method.Proto); googleHttp != nil {
		for _, r := range googleHttp.GetAdditionalBindings() {
			method, endpoint := extensions.GetHttpEndpoint(r)
			rules = append(rules, HttpRule{
				Method:   method,
				Endpoint: endpoint,
			})
		}
	}

	return rules
}

func (m *Method) Validate() error {
	if m.service != nil {
		if authorization := m.service.GetAuthorization(); authorization != nil {
			if authorization.GetMode() == extensions.AuthorizationMode_AUTHORIZATION_MODE_CUSTOM && authorization.GetCustomAuthName() == "" {
				return fmt.Errorf("custom auth name is required when mode is AUTHORIZATION_MODE_CUSTOM")
			}
		}
	}

	return nil
}

func (m *Method) HTTPMethod() string {
	if m.endpoint != nil {
		return m.endpoint.Method
	}

	return ""
}

func (m *Method) Endpoint() string {
	if m.endpoint != nil {
		return m.endpoint.Path
	}

	return ""
}

func (m *Method) HasRequiredBody() bool {
	if m.endpoint != nil {
		return m.endpoint.Body != ""
	}

	return false
}

func (m *Method) AuthModeKey() string {
	if m.service != nil {
		if authorization := m.service.GetAuthorization(); authorization != nil {
			mode := authorization.GetMode()
			if mode == extensions.AuthorizationMode_AUTHORIZATION_MODE_CUSTOM {
				return authorization.GetCustomAuthName()
			}
		}
	}

	return ""
}

func (m *Method) AuthModeValue() string {
	if m.method != nil {
		if http := m.method.GetHttp(); http != nil {
			var args []string
			for _, arg := range http.GetAuthArg() {
				if strings.HasSuffix(arg, "@header") {
					args = append(args, fmt.Sprintf(`string(ctx.Request.Header.Peek("%s"))`, strings.TrimSuffix(arg, "@header")))
					continue
				}

				args = append(args, fmt.Sprintf(`"%s"`, arg))
			}

			return `[]string{` + strings.Join(args, `,`) + `}`
		}
	}

	return ""
}

func (m *Method) HasQueryArguments() bool {
	return len(m.QueryArguments) > 0
}

func (m *Method) HasHeaderArguments() bool {
	return len(m.HeaderArguments) > 0
}

func (m *Method) HasAuth() bool {
	if m.service != nil {
		if authorization := m.service.GetAuthorization(); authorization != nil {
			return authorization.GetMode() != extensions.AuthorizationMode_AUTHORIZATION_MODE_NO_AUTH
		}
	}

	return false
}
