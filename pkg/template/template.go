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
	"github.com/stoewer/go-strcase"
	"google.golang.org/protobuf/compiler/protogen"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/internal/addon"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf"
	tpl_types "github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/template/types"
)

type Context interface {
	// SetTemplateKind receives the current template kind that is being executed
	// in order to the context handle its output properly.
	SetTemplateKind(kind tpl_types.Kind)

	// Validator is the place where each supported template is validated to be
	// executed or not.
	tpl_types.Validator
}

// Templates is an object that holds information related to a group of
// template files, allowing them to be executed later.
type Templates struct {
	strictValidators bool
	basePath         string
	packageName      string
	moduleName       string
	templateInfos    []*Info
}

type Info struct {
	name  string
	kind  tpl_types.Kind
	data  []byte
	api   map[string]interface{}
	addon *addon.Addon
}

type Options struct {
	// StrictValidators if enabled forces us to use the validator function
	// for a template only if it is declared. Otherwise, the template will be
	// ignored.
	StrictValidators bool
	Kind             tpl_types.Kind
	Plugin           *protogen.Plugin
	Files            embed.FS `validate:"required"`
	HelperFunctions  map[string]interface{}
	Addons           []*addon.Addon
}

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
		mName, pName, p, err := protobuf.GetPackageNameAndPath(options.Plugin)
		if err != nil {
			return nil, err
		}

		// module name should not have the version suffix.
		module = strings.TrimSuffix(mName, "v1")
		path = p
		packageName = pName
	}

	infos, err := loadTemplates(options.Files, options.Kind, options.HelperFunctions, nil)
	if err != nil {
		return nil, err
	}

	for _, a := range options.Addons {
		if !canUseAddon(options.Kind, a.Addon().Kind()) {
			continue
		}

		addonsInfo, err := loadTemplates(a.Addon().Templates(), a.Addon().Kind(), options.HelperFunctions, a)
		if err != nil {
			return nil, err
		}
		infos = append(infos, addonsInfo...)
	}

	return &Templates{
		strictValidators: options.StrictValidators,
		basePath:         path,
		packageName:      packageName,
		moduleName:       module,
		templateInfos:    infos,
	}, nil
}

func loadTemplates(files embed.FS, kind tpl_types.Kind, api map[string]interface{}, addon *addon.Addon) ([]*Info, error) {
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
		helperApi := tpl_types.HelperApi()
		for k, v := range api {
			helperApi[k] = v
		}

		helperApi["templateName"] = func() string {
			return tpl_types.NewName(kind, basename).String()
		}

		// Specific addons APIs
		if addon != nil {
			helperApi["addonName"] = func() string {
				return addon.Addon().Name()
			}
		}

		infos = append(infos, &Info{
			name:  basename,
			kind:  kind,
			data:  data,
			api:   helperApi,
			addon: addon,
		})
	}

	return infos, nil
}

func canUseAddon(tplKind, addonKind tpl_types.Kind) bool {
	return tplKind == addonKind
}

func filenameWithoutExtension(filename string) string {
	return filename[:len(filename)-len(filepath.Ext(filename))]
}

type ExecuteOptions struct {
	Context      Context `validate:"required"`
	Path         string
	ModuleName   string
	SingleModule bool
}

// Generated holds the template content already parsed, ready to be saved.
type Generated struct {
	Filename     string
	TemplateName string
	Data         *bytes.Buffer
}

func (t *Templates) Execute(options ExecuteOptions) ([]*Generated, error) {
	execute := func(tpl *Info) (*Generated, error) {
		kind := tpl.kind
		if tpl.addon != nil {
			kind = tpl.addon.Addon().Kind()
		}

		tplValidator := options.Context.GetTemplateValidator
		if tpl.addon != nil {
			tplValidator = tpl.addon.Addon().GetTemplateValidator
		}

		templateValidator, ok := tplValidator(tpl_types.NewName(kind, tpl.name), options.Context)
		if !ok && t.strictValidators {
			// The validator should not be executed in this case, since we don't
			// have one for this template, we can skip it.
			return nil, nil
		}
		if ok && !templateValidator() {
			// Ignores the template if its validation condition is not
			// satisfied
			return nil, nil
		}

		parsedTemplate, err := parse(tpl.name, tpl.data, tpl.api)
		if err != nil {
			return nil, err
		}

		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)

		// Informs the current context which kind of template is being executed.
		options.Context.SetTemplateKind(kind)

		if err := parsedTemplate.Execute(w, options.Context); err != nil {
			return nil, err
		}
		if err := w.Flush(); err != nil {
			return nil, err
		}

		return &Generated{
			Data:         &buf,
			Filename:     t.generateTemplateName(tpl, options, kind),
			TemplateName: tpl.name,
		}, nil
	}

	var gen []*Generated
	for _, tpl := range t.templateInfos {
		g, err := execute(tpl)
		if err != nil {
			return nil, err
		}
		if g != nil {
			gen = append(gen, g)
		}
	}

	return gen, nil
}

func parse(key string, data []byte, helperApi template.FuncMap) (*template.Template, error) {
	t, err := template.New(key).Funcs(helperApi).Parse(string(data))
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (t *Templates) generateTemplateName(tpl *Info, options ExecuteOptions, kind tpl_types.Kind) string {
	prefix := t.moduleName + "."
	if tpl.kind == tpl_types.KindRust {
		// Rust generated files should not have the module name as prefix,
		// because it will change the rust module name if used.
		prefix = ""
	}

	// Filename: Path + Package Name + Module Name + Template Name + Extension
	templateName := fmt.Sprintf("%s%s", prefix, tpl.name)
	if tpl.addon != nil {
		templateName = fmt.Sprintf("%s%s.%s", prefix, strcase.SnakeCase(tpl.addon.Addon().Name()), tpl.name)
	}

	filename := templateName
	if tpl.kind == tpl_types.KindRust {
		if options.SingleModule {
			filename = filepath.Join(options.ModuleName, templateName)
		}
	}
	if !options.SingleModule {
		path := t.basePath
		if options.Path != "" {
			path = options.Path
		}

		filename = filepath.Join(
			path,
			strings.ReplaceAll(t.packageName, ".", "/"),
			templateName,
		)
	}

	if ext := templateExtension(tpl.name, kind); ext != "" {
		filename += fmt.Sprintf(".%s", ext)
	}

	return filename
}

func templateExtension(name string, kind tpl_types.Kind) string {
	// Use the template extension, if it has one.
	if ext := filepath.Ext(name); ext != "" {
		return ""
	}

	return kind.Extension()
}
