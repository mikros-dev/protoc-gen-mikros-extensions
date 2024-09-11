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

	"github.com/rsfreitas/protoc-gen-mikros-extensions/internal/protobuf"
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

type Info struct {
	name string
	data []byte
	api  map[string]interface{}
}

type Options struct {
	// StrictValidators if enabled forces us to use the validator function
	// for a template only if it is declared. Otherwise, the template will be
	// ignored.
	StrictValidators bool
	Path             string
	FilesPrefix      string `validate:"required"`
	Plugin           *protogen.Plugin
	Files            embed.FS `validate:"required"`
	Context          Context  `validate:"required"`
	HelperFunctions  map[string]interface{}
}

// Context is an interface that a template file context, i.e., the
// object manipulated inside the template file, must implement.
type Context interface {
	ValidateForExecution(name Name) (Validator, bool)
	Extension() string
}

type Validator func() bool

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

		helperApi := buildDefaultHelperApi()
		basename := filenameWithoutExtension(t.Name())
		helperApi["templateName"] = func() string {
			return NewName(options.FilesPrefix, basename).String()
		}

		for k, v := range options.HelperFunctions {
			helperApi[k] = v
		}

		infos = append(infos, &Info{
			name: basename,
			data: data,
			api:  helperApi,
		})
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

func buildDefaultHelperApi() map[string]interface{} {
	return template.FuncMap{
		"toLowerCamelCase": strcase.ToLowerCamel,
		"firstLower": func(s string) string {
			c := s[0]
			return strings.ToLower(string(c))
		},
		"toSnake":     strcase.ToSnake,
		"toCamelCase": strcase.ToCamel,
		"toKebab":     strcase.ToKebab,
		"trimSuffix":  strings.TrimSuffix,
	}
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
	var gen []*Generated

	for _, tpl := range t.templateInfos {
		templateValidator, ok := t.context.ValidateForExecution(NewName(t.filesPrefix, tpl.name))
		if !ok && t.strictValidators {
			// The validator should not be executed in this case, since we don't
			// have one for this template, we can skip it.
			continue
		}
		if ok && !templateValidator() {
			// Ignores the template if its validation condition is not
			// satisfied
			continue
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
		filename := filepath.Join(
			t.path,
			strings.ReplaceAll(t.packageName, ".", "/"),
			fmt.Sprintf("%s.%s", t.moduleName, tpl.name),
		)
		if t.context.Extension() != "" {
			filename += fmt.Sprintf(".%s", t.context.Extension())
		}

		gen = append(gen, &Generated{
			Data:         &buf,
			Filename:     filename,
			TemplateName: tpl.name,
			Extension:    t.context.Extension(),
		})
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
