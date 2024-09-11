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
	"google.golang.org/protobuf/compiler/protogen"

	"github.com/rsfreitas/protoc-gen-mikros-extensions/internal/addon"
	"github.com/rsfreitas/protoc-gen-mikros-extensions/internal/protobuf"
	mtemplate "github.com/rsfreitas/protoc-gen-mikros-extensions/pkg/template"
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
	addonInfos       []*Info
}

type Info struct {
	name      string
	data      []byte
	api       map[string]interface{}
	validator mtemplate.Validator
}

type Options struct {
	// StrictValidators if enabled forces us to use the validator function
	// for a template only if it is declared. Otherwise, the template will be
	// ignored.
	StrictValidators bool
	Kind             mtemplate.Kind
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

	infos, err := loadTemplates(options)
	if err != nil {
		return nil, err
	}

	addons, err := loadAddonsTemplates(options)
	if err != nil {
		return nil, err
	}

	return &Templates{
		strictValidators: options.StrictValidators,
		path:             path,
		packageName:      packageName,
		moduleName:       module,
		filesPrefix:      options.FilesPrefix,
		context:          options.Context,
		templateInfos:    infos,
		addonInfos:       addons,
	}, nil
}

func loadTemplates(options Options) ([]*Info, error) {
	templates, err := options.Files.ReadDir(".")
	if err != nil {
		return nil, err
	}

	var infos []*Info
	for _, t := range templates {
		data, err := options.Files.ReadFile(t.Name())
		if err != nil {
			return nil, err
		}

		helperApi := mtemplate.HelperApi()
		basename := filenameWithoutExtension(t.Name())
		helperApi["templateName"] = func() string {
			return mtemplate.NewName(options.FilesPrefix, basename).String()
		}

		for k, v := range options.HelperFunctions {
			helperApi[k] = v
		}

		infos = append(infos, &Info{
			name:      basename,
			data:      data,
			api:       helperApi,
			validator: options.Context,
		})
	}

	return infos, nil
}

func loadAddonsTemplates(options Options) ([]*Info, error) {
	var infos []*Info
	for _, a := range options.Addons {
		if !canUseAddon(options.Kind, a.Kind()) {
			continue
		}

		println("vai executar addon:", a.Name())
		templates, err := a.Templates().ReadDir(".")
		if err != nil {
			return nil, err
		}

		for _, t := range templates {
			data, err := a.Templates().ReadFile(t.Name())
			if err != nil {
				return nil, err
			}

			helperApi := mtemplate.HelperApi()
			basename := filenameWithoutExtension(t.Name())
			helperApi["addonTemplateName"] = func() string {
				return basename
			}

			infos = append(infos, &Info{
				name:      basename,
				data:      data,
				api:       helperApi,
				validator: a,
			})
		}
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
	Extension    string
	Data         *bytes.Buffer
}

func (t *Templates) Execute() ([]*Generated, error) {
	execute := func(tpl *Info, isAddon bool) (*Generated, error) {
		prefix := t.filesPrefix
		if isAddon {
			prefix = "addon"
		}

		templateValidator, ok := tpl.validator.GetTemplateValidator(t.context, mtemplate.NewName(prefix, tpl.name))
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
		if isAddon {
			templateName = fmt.Sprintf("%s.addon.%s", t.moduleName, tpl.name)
		}

		filename := filepath.Join(
			t.path,
			strings.ReplaceAll(t.packageName, ".", "/"),
			templateName,
		)
		if t.context.Extension() != "" {
			filename += fmt.Sprintf(".%s", t.context.Extension())
		}

		return &Generated{
			Data:         &buf,
			Filename:     filename,
			TemplateName: tpl.name,
			Extension:    t.context.Extension(),
		}, nil
	}

	var gen []*Generated

	for _, tpl := range t.templateInfos {
		g, err := execute(tpl, false)
		if err != nil {
			return nil, err
		}
		if g != nil {
			gen = append(gen, g)
		}
	}

	for _, tpl := range t.addonInfos {
		g, err := execute(tpl, true)
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
