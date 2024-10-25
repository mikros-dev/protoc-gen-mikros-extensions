package protobuf

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	descriptor "google.golang.org/protobuf/types/descriptorpb"
)

type Service struct {
	Name    string
	Methods []*Method
	Proto   *descriptor.ServiceDescriptorProto
}

type parseServiceOptions struct {
	Files map[string]*protogen.File
}

func (p *parseServiceOptions) GetApiFile() *protogen.File {
	for _, f := range p.Files {
		// Get the first which has a service definition.
		if len(f.Services) > 0 {
			return f
		}
	}

	return nil
}

func parseService(options *parseServiceOptions) *Service {
	api := options.GetApiFile()

	// Don't do anything if the package does not have an _api.proto file.
	if api == nil {
		return nil
	}

	var (
		service = api.Proto.Service[0]
		methods = make([]*Method, len(service.GetMethod()))
	)

	for i, method := range service.GetMethod() {
		methods[i] = parseMethod(method)
	}

	return &Service{
		Name:    service.GetName(),
		Methods: methods,
		Proto:   service,
	}
}

func (s *Service) IsHTTP() bool {
	for _, m := range s.Methods {
		if m.HTTPMethod != "" {
			return true
		}
	}

	return false
}

func (s *Service) String() string {
	methods := make([]string, len(s.Methods))
	for i, m := range s.Methods {
		methods[i] = m.String()
	}

	return fmt.Sprintf(`{name:%v, methods:[%v]}`,
		s.Name,
		strings.Join(methods, ","))
}
