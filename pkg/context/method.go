package context

import (
	"fmt"
	"slices"
	"strings"

	"github.com/stoewer/go-strcase"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/internal/translation"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/converters"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/mikros_extensions"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
	tpl_types "github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/template/types"
)

type Method struct {
	Name                  string
	RequestType           string
	ResponseType          string
	AdditionalHTTPMethods []HttpRule
	Request               *Message
	PathArguments         []*MethodField
	QueryArguments        []*MethodField
	HeaderArguments       []*MethodField
	BodyArguments         []*MethodField
	ProtoMethod           *protobuf.Method

	prefixServiceName bool
	moduleName        string
	endpoint          *Endpoint
	service           *mikros_extensions.MikrosServiceExtensions
	method            *mikros_extensions.MikrosMethodExtensions
}

type HttpRule struct {
	Method   string
	Endpoint string
}

type MethodField struct {
	GoName    string
	ProtoName string
	CastType  string
	Field     *Field
}

func loadMethods(pkg *protobuf.Protobuf, messages []*Message, cfg *settings.Settings) ([]*Method, error) {
	if pkg.Service == nil {
		return nil, nil
	}

	var (
		methods = make([]*Method, len(pkg.Service.Methods))
		service = mikros_extensions.LoadServiceExtensions(pkg.Service.Proto)
	)

	for i, method := range pkg.Service.Methods {
		var (
			msg              *Message
			endpoint         = getEndpoint(method)
			methodExtensions = mikros_extensions.LoadMethodExtensions(method.Proto)
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

		body, err := getBodyParameters(msg, endpoint, methodExtensions)
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
			AdditionalHTTPMethods: getAdditionalHttpRules(method),
			Request:               msg,
			PathArguments:         path,
			QueryArguments:        getQueryArguments(msg, endpoint, methodExtensions),
			HeaderArguments:       header,
			BodyArguments:         body,
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
				return nil, fmt.Errorf("field '%s' declared in path arguments not found inside message '%s' definition", name, m.Name)
			}

			field := m.Fields[index]
			fields = append(fields, &MethodField{
				GoName:    field.GoName,
				ProtoName: field.ProtoName,
				CastType:  field.GoType,
				Field:     field,
			})
		}
	}

	return fields, nil
}

func getHeaderArguments(m *Message, methodExtensions *mikros_extensions.MikrosMethodExtensions) ([]*MethodField, error) {
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
				Field:     field,
			})
		}
	}

	return fields, nil
}

func getQueryArguments(m *Message, endpoint *Endpoint, methodExtensions *mikros_extensions.MikrosMethodExtensions) []*MethodField {
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
					Field:     field,
				})
			}
		}
	}

	return fields
}

func getParametersToFilter(m *Message, endpoint *Endpoint, methodExtensions *mikros_extensions.MikrosMethodExtensions) []string {
	parameters := getBodyParametersFromEndpoint(m, endpoint, methodExtensions)

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

func getBodyParameters(m *Message, endpoint *Endpoint, methodExtensions *mikros_extensions.MikrosMethodExtensions) ([]*MethodField, error) {
	var fields []*MethodField

	if endpoint != nil {
		var (
			parameters = getBodyParametersFromEndpoint(m, endpoint, methodExtensions)
		)

		// Remove path and header parameters if any

		for _, name := range parameters {
			index := slices.IndexFunc(m.Fields, func(f *Field) bool {
				return f.ProtoName == name
			})
			if index == -1 {
				return nil, fmt.Errorf("header field '%s' not found inside message '%s' definition", name, m.Name)
			}

			field := m.Fields[index]
			fields = append(fields, &MethodField{
				GoName:    field.GoName,
				ProtoName: field.ProtoName,
				CastType:  field.GoType,
				Field:     field,
			})
		}
	}

	return fields, nil
}

func getBodyParametersFromEndpoint(m *Message, endpoint *Endpoint, methodExtensions *mikros_extensions.MikrosMethodExtensions) []string {
	var parameters []string

	if endpoint != nil {
		// Is the method using all fields as body arguments?
		if endpoint.Body == "*" {
			for _, f := range m.Fields {
				parameters = append(parameters, f.ProtoName)
			}
		}
		// Or it is using specific fields for that?
		if endpoint.Body != "*" && len(endpoint.Body) > 0 {
			parameters = append(parameters, strings.Split(endpoint.Body, " ")...)
		}

		// Remove path and header parameters if any
		for _, param := range endpoint.Parameters {
			parameters = slices.DeleteFunc(parameters, func(p string) bool {
				return p == param
			})
		}
		if methodExtensions != nil {
			if httpExtensions := methodExtensions.GetHttp(); httpExtensions != nil {
				for _, param := range httpExtensions.GetHeader() {
					parameters = slices.DeleteFunc(parameters, func(p string) bool {
						return p == param
					})
				}
			}
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

	if googleHttp := mikros_extensions.LoadGoogleAnnotations(method.Proto); googleHttp != nil {
		for _, r := range googleHttp.GetAdditionalBindings() {
			method, endpoint := mikros_extensions.GetHttpEndpoint(r)
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
			if authorization.GetMode() == mikros_extensions.AuthorizationMode_AUTHORIZATION_MODE_CUSTOM && authorization.GetCustomAuthName() == "" {
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
		endpoint := m.endpoint.Path
		if m.prefixServiceName {
			endpoint = fmt.Sprintf("/%v%v", strcase.KebabCase(m.moduleName), endpoint)
		}

		return endpoint
	}

	return ""
}

func (m *Method) EndpointByTemplateKind(kind tpl_types.Kind) string {
	if endpoint := m.Endpoint(); endpoint != "" {
		if kind == tpl_types.KindRust {
			return translation.RustEndpoint(endpoint)
		}

		return endpoint
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
			if mode == mikros_extensions.AuthorizationMode_AUTHORIZATION_MODE_CUSTOM {
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
			return authorization.GetMode() != mikros_extensions.AuthorizationMode_AUTHORIZATION_MODE_NO_AUTH
		}
	}

	return false
}

func (m *Method) HasPathArguments() bool {
	return len(m.PathArguments) > 0
}

func (m *Method) GetPathParameterNames() string {
	if m.HasPathArguments() {
		names := make([]string, len(m.PathArguments))
		for i, arg := range m.PathArguments {
			names[i] = arg.ProtoName
		}

		parameters := strings.Join(names, ", ")

		// More than one parameter should use a tuple.
		if len(names) > 1 {
			parameters = "(" + parameters + ")"
		}

		return parameters
	}

	return ""
}

func (m *Method) GetPathParameterTypesByTemplateKind(kind tpl_types.Kind) string {
	if m.HasPathArguments() {
		types := make([]string, len(m.PathArguments))
		for i, arg := range m.PathArguments {
			types[i] = arg.Field.TypeByTemplateKind(kind)
		}

		parameters := strings.Join(types, ", ")

		// More than one parameter should use a tuple.
		if len(types) > 1 {
			parameters = "(" + parameters + ")"
		}

		return parameters
	}

	return ""
}

func (m *Method) HasBodyArguments() bool {
	return len(m.BodyArguments) > 0
}
