package protobuf

import (
	"errors"
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

func getPackageProtoFiles(plugin *protogen.Plugin) (map[string]*protogen.File, error) {
	var (
		files = make(map[string]*protogen.File)
	)

	info, err := GetPackageInfo(plugin)
	if err != nil {
		return nil, err
	}

	for name, file := range plugin.FilesByPath {
		if isProtoFileFromCurrentPackage(file, info.ModuleName) {
			files[strings.TrimSuffix(filepath.Base(name), ".proto")] = file
		}
	}

	if len(files) == 0 {
		return nil, errors.New("could not find a supported .proto file")
	}

	return files, nil
}

// PackageInfo contains the package name, module name and path of the current
// file being processed.
type PackageInfo struct {
	PackageName string
	ModuleName  string
	Path        string
}

// GetPackageInfo try to retrieve the golang module name from the list of .proto
// files.
func GetPackageInfo(plugin *protogen.Plugin) (*PackageInfo, error) {
	if len(plugin.Files) == 0 {
		return nil, errors.New("cannot find the module name without .proto files")
	}

	// The last file in the slice is always the main .proto file that is being
	// "compiled" by protoc.
	file := plugin.Files[len(plugin.Files)-1]

	path := strings.ReplaceAll(file.GoImportPath.String(), "\"", "")
	return &PackageInfo{
		PackageName: file.Proto.GetPackage(),
		ModuleName:  string(file.GoPackageName),
		Path:        path,
	}, nil
}

func isProtoFileFromCurrentPackage(file *protogen.File, packageName string) bool {
	return string(file.GoPackageName) == packageName
}
