package imports

import (
	"cmp"
	"regexp"
	"slices"
	"strings"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/template/spec"
)

// Context represents the context template information specific for imports.
type Context struct {
	HasValidatableMessage   bool
	OutboundHasBitflagField bool
	UseCommonConverters     bool
	ModuleName              string
	FullPath                string
	Methods                 []*Method
	DomainMessages          []*Message
	OutboundMessages        []*Message
	ValidatableMessages     []*Message
	WireExtensions          []*Message
	WireInputMessages       []*Message
}

// Message represents a message.
type Message struct {
	ValidationNeedsCustomRuleOptions bool
	IsWireInputKind                  bool
	Receiver                         string
	Fields                           []*Field
	ProtoMessage                     *protobuf.Message
}

// Field represents a field inside a Message.
type Field struct {
	IsArray                        bool
	IsProtobufTimestamp            bool
	IsOutboundBitflag              bool
	IsMessage                      bool
	OutboundHide                   bool
	ConversionDomainToWire         string
	ConversionWireToDomain         string
	ConversionWireOutputToOutbound string
	DomainType                     string
	WireType                       string
	OutboundType                   string
	TestingBinding                 string
	TestingCall                    string
	ValidationCall                 string
	ProtoField                     *protobuf.Field
}

// Method represents a method declared inside a service.
type Method struct {
	HasRequiredBody    bool
	HasQueryArguments  bool
	HasHeaderArguments bool
}

// Import represents an import statement inside a template.
type Import struct {
	Alias string
	Name  string
}

// Importer represents the API that an internal template file must implement to
// be able to load its imports.
type Importer interface {
	// Name must return the template name.
	Name() spec.Name

	// Load must return a slice of imports for the template.
	Load(ctx *Context, cfg *settings.Settings) []*Import
}

// LoadTemplateImports loads the imports for the templates.
func LoadTemplateImports(ctx *Context, cfg *settings.Settings) map[spec.Name][]*Import {
	var (
		generatedImports = make(map[spec.Name][]*Import)
		templates        = []Importer{
			&Domain{},
			&Enum{},
			&CustomAPI{},
			&HTTPServer{},
			&Routes{},
			&Wire{},
			&WireInput{},
			&Outbound{},
			&Common{},
			&Validation{},
			&Testing{},
			&TestingHTTPServer{},
		}
	)

	for _, tpl := range templates {
		generatedImports[tpl.Name()] = tpl.Load(ctx, cfg)
	}

	return generatedImports
}

func toSlice(ipt map[string]*Import) []*Import {
	var (
		s     = make([]*Import, len(ipt))
		index = 0
	)

	for _, i := range ipt {
		s[index] = i
		index++
	}

	slices.SortFunc(s, func(a, b *Import) int {
		if a.Alias != "" && b.Alias != "" {
			return cmp.Compare(a.Alias, b.Alias)
		}
		if a.Alias != "" && b.Alias == "" {
			return cmp.Compare(a.Alias, b.Name)
		}
		if a.Alias == "" && b.Alias != "" {
			return cmp.Compare(a.Name, b.Alias)
		}

		return cmp.Compare(a.Name, b.Name)
	})

	return s
}

func addConvertersIfNeeded(imports map[string]*Import, cfg *settings.Settings, binding string) {
	// Import user converters package?
	if i, ok := needsUserConvertersPackage(cfg, binding); ok {
		imports["converters"] = i
	}
}

func needsUserConvertersPackage(cfg *settings.Settings, conversionCall string) (*Import, bool) {
	if cfg.Templates.Common != nil {
		for _, dep := range cfg.Templates.Common.API {
			var moduleName string
			if dep.Import != nil {
				moduleName = dep.Import.ModuleName()
			}

			if strings.HasPrefix(conversionCall, moduleName) {
				return &Import{
					Alias: dep.Import.Alias,
					Name:  dep.Import.Name,
				}, true
			}
		}
	}

	return nil, false
}

func addTimeIfNeeded(imports map[string]*Import, f *Field) bool {
	// Import time package?
	if f.IsProtobufTimestamp {
		imports["time"] = packages["time"]
		return true
	}

	return false
}

func addProtoTimestampIfNeeded(imports map[string]*Import, wireType string) bool {
	// Import proto timestamp package?
	if strings.HasPrefix(wireType, "ts.") || strings.HasPrefix(wireType, "*ts.") {
		imports["prototimestamp"] = packages["prototimestamp"]
		return true
	}

	return false
}

func addModuleIfNeeded(imports map[string]*Import, binding, fieldType, currentModule, receiver, fullPath string) {
	if module, ok := needsImportAnotherProtoModule(binding, fieldType, currentModule, receiver); ok {
		imports[module] = importAnotherModule(module, currentModule, fullPath)
	}
}

// needsImportAnotherProtoModule checks if a conversion call that is being made must have
// another module imported.
func needsImportAnotherProtoModule(conversionCall, fieldType, moduleName, receiver string) (string, bool) {
	if m, ok := checkImportNeededFromConversionCall(conversionCall, moduleName, receiver); ok {
		return m, ok
	}

	if m, ok := checkImportNeededFromFieldType(fieldType); ok {
		return m, ok
	}

	return "", false
}

func checkImportNeededFromConversionCall(conversionCall, moduleName, receiver string) (string, bool) {
	// Don't bother checking if it is not a function call
	if !strings.Contains(conversionCall, "(") {
		return "", false
	}

	var (
		parts          = strings.Split(conversionCall, ".")
		ignoredModules = []string{moduleName, receiver, "converters"}
	)

	// The conversion should have a module as prefix, like "something.", and
	// it should split to more than 5 parts because the module name usually
	// repeat in it.
	if len(parts) == 0 || len(parts) < 5 || slices.Contains(ignoredModules, parts[0]) {
		return "", false
	}

	return parts[0], true
}

var (
	mapTypeRe = regexp.MustCompile(`^map\[\S+]`)
)

func checkImportNeededFromFieldType(fieldType string) (string, bool) {
	if strings.HasPrefix(fieldType, "map[") {
		fieldType = mapTypeRe.ReplaceAllString(fieldType, "")
	}
	if strings.HasPrefix(fieldType, "[]") {
		fieldType = strings.TrimPrefix(fieldType, "[]")
	}

	fieldType = strings.TrimPrefix(fieldType, "*")
	parts := strings.Split(fieldType, ".")

	// fieldType must have two parts here 'module.Wire' to require another
	// module to be imported
	if len(parts) == 2 {
		return parts[0], true
	}

	return "", false
}

func importAnotherModule(moduleName, currentModuleName, importPath string) *Import {
	return &Import{
		Alias: moduleName,
		Name:  strings.ReplaceAll(importPath, currentModuleName, moduleName),
	}
}
