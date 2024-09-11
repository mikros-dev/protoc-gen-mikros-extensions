package protobuf

import (
	"errors"
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

func getProtoFiles(plugin *protogen.Plugin) (map[string]*protogen.File, error) {
	var (
		files = make(map[string]*protogen.File)
	)

	moduleName, _, _, err := GetPackageNameAndPath(plugin)
	if err != nil {
		return nil, err
	}

	for name, file := range plugin.FilesByPath {
		if isProtoFileFromCurrentPackage(file, moduleName) {
			files[strings.TrimSuffix(filepath.Base(name), ".proto")] = file
		}
	}

	if len(files) == 0 {
		return nil, errors.New("could not find a supported .proto file")
	}

	return files, nil
}

// GetPackageNameAndPath try to retrieve the golang module name from the list of .proto
// files.
func GetPackageNameAndPath(plugin *protogen.Plugin) (string, string, string, error) {
	if len(plugin.Files) == 0 {
		return "", "", "", errors.New("cannot find the module name without .proto files")
	}

	// The last file in the slice is always the main .proto file that is being
	// "compiled" by protoc.
	file := plugin.Files[len(plugin.Files)-1]

	path := strings.ReplaceAll(file.GoImportPath.String(), "\"", "")
	return string(file.GoPackageName), file.Proto.GetPackage(), path, nil
}

func isProtoFileFromCurrentPackage(file *protogen.File, packageName string) bool {
	return string(file.GoPackageName) == packageName
}
