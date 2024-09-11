package protobuf

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	descriptor "google.golang.org/protobuf/types/descriptorpb"
)

type Protobuf struct {
	ModuleName  string
	PackageName string
	FullPath    string
	Service     *Service
	Messages    []*Message
	Enums       []*Enum
	Files       map[string]*descriptor.FileDescriptorProto
}

type ParseOptions struct {
	Plugin *protogen.Plugin
}

func Parse(options ParseOptions) (*Protobuf, error) {
	moduleName, packageName, path, err := GetPackageNameAndPath(options.Plugin)
	if err != nil {
		return nil, err
	}

	protoFiles, err := getProtoFiles(options.Plugin)
	if err != nil {
		return nil, err
	}

	files := make(map[string]*descriptor.FileDescriptorProto)
	for k, v := range protoFiles {
		files[k] = getFileDescriptor(v)
	}

	return &Protobuf{
		ModuleName:  moduleName,
		PackageName: packageName,
		FullPath:    path,
		Files:       files,
		Enums: parseEnums(&parseEnumsOptions{
			Files: protoFiles,
		}),
		Service: parseService(&parseServiceOptions{
			Files: protoFiles,
		}),
		Messages: parseMessages(&parseMessagesOptions{
			ModuleName: moduleName,
			Files:      protoFiles,
		}),
	}, nil
}

func getFileDescriptor(file *protogen.File) *descriptor.FileDescriptorProto {
	if file != nil {
		return file.Proto
	}

	return nil
}

func (p *Protobuf) String() string {
	enums := make([]string, len(p.Enums))
	for i, e := range p.Enums {
		enums[i] = e.String()
	}

	s := fmt.Sprintf(`{name:%v, path:%v, enums:[%v]`,
		p.ModuleName,
		p.FullPath,
		strings.Join(enums, ","))

	if p.Service != nil {
		s += ", service:" + p.Service.String()
	}
	if p.Messages != nil {
		messages := make([]string, len(p.Messages))
		for i, m := range p.Messages {
			messages[i] = m.String()
		}

		s += ", messages:[" + strings.Join(messages, ",") + "]"
	}

	return s + "}"
}
