package context

import (
	"fmt"
	"slices"
	"strings"

	"github.com/stoewer/go-strcase"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/mapping"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf/extensions"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
)

// Method represents a method to be used inside templates by its context.
type Method struct {
	Name                  string
	RequestType           string
	ResponseType          string
	AdditionalHTTPMethods []HTTPRule
	Request               *Message
	PathArguments         []*MethodField
	QueryArguments        []*MethodField
	HeaderArguments       []*MethodField
	ProtoMethod           *protobuf.Method

	prefixServiceName bool
	moduleName        string
	endpoint          *Endpoint
	service           *extensions.MikrosServiceExtensions
	method            *extensions.MikrosMethodExtensions
}

// HTTPRule represents a HTTP rule defined inside a method.
type HTTPRule struct {
	Method   string
	Endpoint string
}

// MethodField represents a field of a method.
type MethodField struct {
	GoName    string
	ProtoName string
	CastType  string
}

func loadMethods(pkg *protobuf.Protobuf, messages []*Message, cfg *settings.Settings) ([]*Method, error) {
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
			return m.Name == method.RequestType.Name && m.Type == mapping.WireInput
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
			RequestType:           method.RequestType.Name,
			ResponseType:          method.ResponseType.Name,
			AdditionalHTTPMethods: getAdditionalHTTPRules(method),
			Request:               msg,
			PathArguments:         path,
			QueryArguments:        getQueryArguments(msg, endpoint, methodExtensions),
			HeaderArguments:       header,
			ProtoMethod:           method,
			prefixServiceName:     cfg.Templates.Routes.PrefixServiceName,
			moduleName:            pkg.ModuleName,
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
				return nil, fmt.Errorf(
					"field '%s' declared in path arguments not found inside message '%s' definition",
					name,
					m.Name,
				)
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

func getHeaderArguments(
	m *Message,
	methodExtensions *extensions.MikrosMethodExtensions,
) ([]*MethodField, error) {
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

func getQueryArguments(
	m *Message,
	endpoint *Endpoint,
	methodExtensions *extensions.MikrosMethodExtensions,
) []*MethodField {
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

func getParametersToFilter(
	m *Message,
	endpoint *Endpoint,
	methodExtensions *extensions.MikrosMethodExtensions,
) []string {
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

func getAdditionalHTTPRules(method *protobuf.Method) []HTTPRule {
	var rules []HTTPRule

	if googleHTTP := extensions.LoadGoogleAnnotations(method.Proto); googleHTTP != nil {
		for _, r := range googleHTTP.GetAdditionalBindings() {
			method, endpoint := extensions.GetHTTPEndpoint(r)
			rules = append(rules, HTTPRule{
				Method:   method,
				Endpoint: endpoint,
			})
		}
	}

	return rules
}

// Validate validates if the method is properly configured.
func (m *Method) Validate() error {
	if m.service != nil {
		if authorization := m.service.GetAuthorization(); authorization != nil {
			isCustomMode := authorization.GetMode() == extensions.AuthorizationMode_AUTHORIZATION_MODE_CUSTOM
			isAuthNameEmpty := authorization.GetCustomAuthName() == ""
			if isCustomMode && isAuthNameEmpty {
				return fmt.Errorf("custom auth name is required when mode is AUTHORIZATION_MODE_CUSTOM")
			}
		}
	}

	return nil
}

// HTTPMethod returns the HTTP method of the method.
func (m *Method) HTTPMethod() string {
	if m.endpoint != nil {
		return m.endpoint.Method
	}

	return ""
}

// Endpoint returns the endpoint of the method.
func (m *Method) Endpoint() string {
	if m.endpoint != nil {
		endpoint := m.endpoint.Path
		if m.prefixServiceName {
			endpoint = fmt.Sprintf("/%v%v", strcase.KebabCase(m.moduleName), endpoint)
		}

		return endpoint
	}

	return ""
}

// HasRequiredBody returns true if the method has a required body.
func (m *Method) HasRequiredBody() bool {
	if m.endpoint != nil {
		return m.endpoint.Body != "" && !m.ParseRequestInService()
	}

	return false
}

// AuthModeKey returns the key of the auth mode. If the mode is
// AUTHORIZATION_MODE_CUSTOM, it returns the custom auth name.
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

// AuthModeValue returns the value of the auth mode. If the mode is
// AUTHORIZATION_MODE_CUSTOM, it returns the auth arguments.
func (m *Method) AuthModeValue() string {
	if m.method == nil {
		return ""
	}

	http := m.method.GetHttp()
	if http == nil {
		return ""
	}

	var args []string
	for _, arg := range http.GetAuthArg() {
		if strings.HasSuffix(arg, "@header") {
			argument := fmt.Sprintf(`string(ctx.Request.Header.Peek("%s"))`, strings.TrimSuffix(arg, "@header"))
			args = append(args, argument)
			continue
		}

		args = append(args, fmt.Sprintf(`"%s"`, arg))
	}

	return `[]string{` + strings.Join(args, `,`) + `}`
}

// HasQueryArguments returns true if the method has query arguments.
func (m *Method) HasQueryArguments() bool {
	return len(m.QueryArguments) > 0
}

// HasHeaderArguments returns true if the method has header arguments.
func (m *Method) HasHeaderArguments() bool {
	return len(m.HeaderArguments) > 0
}

// HasAuth returns true if the method has auth.
func (m *Method) HasAuth() bool {
	if m.service != nil {
		if authorization := m.service.GetAuthorization(); authorization != nil {
			return authorization.GetMode() != extensions.AuthorizationMode_AUTHORIZATION_MODE_NO_AUTH
		}
	}

	return false
}

// ParseRequestInService returns true if the request should be parsed in the
// service.
func (m *Method) ParseRequestInService() bool {
	if m.method != nil {
		if http := m.method.GetHttp(); http != nil {
			return http.GetParseRequestInService()
		}
	}

	return false
}
