package template

import (
	"bufio"
	"bytes"
	"embed"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/go-playground/validator/v10"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/template/spec"
	"github.com/stoewer/go-strcase"
	"google.golang.org/protobuf/compiler/protogen"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/internal/addon"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf"
)

// Templates is an object that holds information related to a group of
// template files, allowing them to be executed later.
type Templates struct {
	strictValidators bool
	path             string
	packageName      string
	moduleName       string
	filesPrefix      string
	context          Context
	templateInfos    []*Info
}

// Info holds information about a template file.
type Info struct {
	name  string
	data  []byte
	api   map[string]interface{}
	addon *addon.Addon
}

// Options represents the options used to load the templates.
type Options struct {
	// StrictValidators if enabled forces us to use the validator function
	// for a template only if it is declared. Otherwise, the template will be
	// ignored.
	StrictValidators bool
	Kind             spec.Kind
	Path             string
	FilesPrefix      string `validate:"required"`
	Plugin           *protogen.Plugin
	Files            embed.FS `validate:"required"`
	Context          Context  `validate:"required"`
	HelperFunctions  map[string]interface{}
	Addons           []*addon.Addon
}

// Context is an interface that a template file context, i.e., the
// object manipulated inside the template file, must implement.
type Context interface {
	Extension() string
	spec.Validator
}

// Load loads the templates from the given options.
func Load(options Options) (*Templates, error) {
	validate := validator.New()
	if err := validate.Struct(options); err != nil {
		return nil, err
	}

	var (
		path        string
		module      string
		packageName string
	)

	if options.Plugin != nil {
		info, err := protobuf.GetPackageInfo(options.Plugin)
		if err != nil {
			return nil, err
		}

		// module name should not have the version suffix.
		module = strings.TrimSuffix(info.ModuleName, "v1")
		path = info.Path
		packageName = info.PackageName
	}

	if options.Path != "" {
		path = options.Path
	}

	infos, err := loadTemplates(options.Files, options.FilesPrefix, options.HelperFunctions, nil)
	if err != nil {
		return nil, err
	}

	for _, a := range options.Addons {
		if !canUseAddon(options.Kind, a.Addon().Kind()) {
			continue
		}

		addonsInfo, err := loadTemplates(a.Addon().Templates(), a.Addon().Name(), options.HelperFunctions, a)
		if err != nil {
			return nil, err
		}
		infos = append(infos, addonsInfo...)
	}

	return &Templates{
		strictValidators: options.StrictValidators,
		path:             path,
		packageName:      packageName,
		moduleName:       module,
		filesPrefix:      options.FilesPrefix,
		context:          options.Context,
		templateInfos:    infos,
	}, nil
}

func loadTemplates(files embed.FS, prefix string, api map[string]interface{}, addon *addon.Addon) ([]*Info, error) {
	templates, err := files.ReadDir(".")
	if err != nil {
		return nil, err
	}

	var infos []*Info
	for _, t := range templates {
		data, err := files.ReadFile(t.Name())
		if err != nil {
			return nil, err
		}

		basename := filenameWithoutExtension(t.Name())
		helperAPI := spec.DefaultFuncMap()
		for k, v := range api {
			helperAPI[k] = v
		}

		helperAPI["templateName"] = func() string {
			return spec.NewName(prefix, basename).String()
		}

		// Specific addons APIs
		if addon != nil {
			helperAPI["addonName"] = func() string {
				return addon.Addon().Name()
			}
		}

		infos = append(infos, &Info{
			name:  basename,
			data:  data,
			api:   helperAPI,
			addon: addon,
		})
	}

	return infos, nil
}

func canUseAddon(tplKind, addonKind spec.Kind) bool {
	return tplKind == addonKind
}

func filenameWithoutExtension(filename string) string {
	return filename[:len(filename)-len(filepath.Ext(filename))]
}

// Generated holds the template content already parsed, ready to be saved.
type Generated struct {
	Filename     string
	TemplateName string
	Extension    string
	Data         *bytes.Buffer
}

// Execute executes the templates and returns the generated files.
func (t *Templates) Execute() ([]*Generated, error) {
	var gen []*Generated
	for _, tpl := range t.templateInfos {
		g, err := t.executeSingleTemplate(tpl)
		if err != nil {
			return nil, err
		}
		if g != nil {
			gen = append(gen, g)
		}
	}

	return gen, nil
}

func (t *Templates) executeSingleTemplate(tpl *Info) (*Generated, error) {
	if skip, err := t.shouldSkipTemplate(tpl); err != nil || skip {
		return nil, err
	}

	parsedTemplate, err := parse(tpl.name, tpl.data, tpl.api)
	if err != nil {
		return nil, err
	}

	buf, err := t.executeTemplate(parsedTemplate)
	if err != nil {
		return nil, err
	}

	return &Generated{
		Data:         buf,
		Filename:     t.buildOutputFilename(tpl),
		TemplateName: tpl.name,
		Extension:    t.context.Extension(),
	}, nil
}

func (t *Templates) shouldSkipTemplate(tpl *Info) (bool, error) {
	prefix := t.filesPrefix
	if tpl.addon != nil {
		prefix = tpl.addon.Addon().Name()
	}

	tplValidatorProvider := t.context.GetTemplateValidator
	if tpl.addon != nil {
		tplValidatorProvider = tpl.addon.Addon().GetTemplateValidator
	}

	validatorFn, ok := tplValidatorProvider(spec.NewName(prefix, tpl.name), t.context)
	if !ok && t.strictValidators {
		// The validator should not be executed in this case. Since we don't
		// have one for this template, we can skip it.
		return true, nil
	}
	if ok && !validatorFn() {
		// Ignores the template if its validation condition is not
		// satisfied
		return true, nil
	}

	return false, nil
}

func parse(key string, data []byte, helperAPI template.FuncMap) (*template.Template, error) {
	t, err := template.New(key).Funcs(helperAPI).Parse(string(data))
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (t *Templates) executeTemplate(tpl *template.Template) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	if err := tpl.Execute(w, t.context); err != nil {
		return nil, err
	}
	if err := w.Flush(); err != nil {
		return nil, err
	}

	return &buf, nil
}

func (t *Templates) buildOutputFilename(tpl *Info) string {
	// Filename: Path + Package Name + Module Name + Template Name + Extension
	templateName := fmt.Sprintf("%s.%s", t.moduleName, tpl.name)
	if tpl.addon != nil {
		templateName = fmt.Sprintf("%s.%s.%s", t.moduleName, strcase.SnakeCase(tpl.addon.Addon().Name()), tpl.name)
	}

	filename := filepath.Join(
		t.path,
		strings.ReplaceAll(t.packageName, ".", "/"),
		templateName,
	)

	if ext := t.context.Extension(); ext != "" {
		filename += fmt.Sprintf(".%s", ext)
	}

	return filename
}
