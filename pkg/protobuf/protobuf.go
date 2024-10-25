package protobuf

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

type Protobuf struct {
	ModuleName   string
	PackageName  string
	FullPath     string
	Service      *Service
	Messages     []*Message
	Enums        []*Enum
	PackageFiles map[string]*protogen.File
	Files        map[string]*protogen.File
}

type ParseOptions struct {
	Plugin *protogen.Plugin
}

func Parse(options ParseOptions) (*Protobuf, error) {
	moduleName, packageName, path, err := GetPackageNameAndPath(options.Plugin)
	if err != nil {
		return nil, err
	}

	packageProtoFiles, err := getPackageProtoFiles(options.Plugin)
	if err != nil {
		return nil, err
	}

	packageFiles := make(map[string]*protogen.File)
	for k, v := range packageProtoFiles {
		packageFiles[k] = v
	}

	files := make(map[string]*protogen.File)
	for name, f := range options.Plugin.FilesByPath {
		if string(f.GoPackageName) != moduleName {
			files[name] = f
		}
	}

	return &Protobuf{
		ModuleName:  moduleName,
		PackageName: packageName,
		FullPath:    path,
		Service: parseService(&parseServiceOptions{
			Files: packageProtoFiles,
		}),
		Messages: parseMessages(&parseMessagesOptions{
			ModuleName: moduleName,
			Files:      packageProtoFiles,
		}),
		Enums: parseEnums(&parseEnumsOptions{
			Files: packageProtoFiles,
		}),
		PackageFiles: packageFiles,
		Files:        files,
	}, nil
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
