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
	"github.com/iancoleman/strcase"
	"google.golang.org/protobuf/compiler/protogen"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/internal/addon"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf"
	mtemplate "github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/template"
)

// Templates is an object that holds information related to a group of
// template files, allowing them to be executed later.
type Templates struct {
	strictValidators bool
	path             string
	packageName      string
	moduleName       string
	context          Context
	templateInfos    []*Info
}

type Info struct {
	name  string
	kind  mtemplate.Kind
	data  []byte
	api   map[string]interface{}
	addon *addon.Addon
}

type Options struct {
	// StrictValidators if enabled forces us to use the validator function
	// for a template only if it is declared. Otherwise, the template will be
	// ignored.
	StrictValidators bool
	Kind             mtemplate.Kind
	Path             string
	Plugin           *protogen.Plugin
	Files            embed.FS `validate:"required"`
	Context          Context  `validate:"required"`
	HelperFunctions  map[string]interface{}
	Addons           []*addon.Addon
}

// Context is an interface that a template file context, i.e., the
// object manipulated inside the template file, must implement.
type Context interface {
	mtemplate.Validator
}

func LoadTemplates(options Options) (*Templates, error) {
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

	if options.Path != "" {
		path = options.Path
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
		path:             path,
		packageName:      packageName,
		moduleName:       module,
		context:          options.Context,
		templateInfos:    infos,
	}, nil
}

func loadTemplates(files embed.FS, kind mtemplate.Kind, api map[string]interface{}, addon *addon.Addon) ([]*Info, error) {
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
		helperApi := mtemplate.HelperApi()
		for k, v := range api {
			helperApi[k] = v
		}

		helperApi["templateName"] = func() string {
			return mtemplate.NewName(kind, basename).String()
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

func canUseAddon(tplKind, addonKind mtemplate.Kind) bool {
	return tplKind == addonKind
}

func filenameWithoutExtension(filename string) string {
	return filename[:len(filename)-len(filepath.Ext(filename))]
}

// Generated holds the template content already parsed, ready to be saved.
type Generated struct {
	Filename     string
	TemplateName string
	Data         *bytes.Buffer
}

func (t *Templates) Execute() ([]*Generated, error) {
	execute := func(tpl *Info) (*Generated, error) {
		kind := tpl.kind
		if tpl.addon != nil {
			kind = tpl.addon.Addon().Kind()
		}

		tplValidator := t.context.GetTemplateValidator
		if tpl.addon != nil {
			tplValidator = tpl.addon.Addon().GetTemplateValidator
		}

		templateValidator, ok := tplValidator(mtemplate.NewName(kind, tpl.name), t.context)
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

		if err := parsedTemplate.Execute(w, t.context); err != nil {
			return nil, err
		}
		if err := w.Flush(); err != nil {
			return nil, err
		}

		// Filename: Path + Package Name + Module Name + Template Name + Extension
		templateName := fmt.Sprintf("%s.%s", t.moduleName, tpl.name)
		if tpl.addon != nil {
			templateName = fmt.Sprintf("%s.%s.%s", t.moduleName, strcase.ToSnake(tpl.addon.Addon().Name()), tpl.name)
		}

		filename := filepath.Join(
			t.path,
			strings.ReplaceAll(t.packageName, ".", "/"),
			templateName,
		)
		if kind.Extension() != "" {
			filename += fmt.Sprintf(".%s", kind.Extension())
		}

		return &Generated{
			Data:         &buf,
			Filename:     filename,
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
