package plugin

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"strings"

	"github.com/bufbuild/protoplugin"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/internal/addon"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/internal/args"
	api_tpl_files "github.com/mikros-dev/protoc-gen-mikros-extensions/internal/template/api"
	test_tpl_files "github.com/mikros-dev/protoc-gen-mikros-extensions/internal/template/testing"
	mcontext "github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/context"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/output"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/template"
	tpl_types "github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/template/types"
)

type execution struct {
	Kind   tpl_types.Kind
	Path   string
	Prefix string
	Files  embed.FS
}

func Handle(
	_ context.Context,
	_ protoplugin.PluginEnv,
	w protoplugin.ResponseWriter,
	r protoplugin.Request,
) error {
	pluginArgs, err := args.NewArgsFromString(r.Parameter())
	if err != nil {
		return err
	}

	plugin, err := protogen.Options{}.New(r.CodeGeneratorRequest())
	if err != nil {
		return err
	}

	if err := handleProtogenPlugin(plugin, pluginArgs); err != nil {
		return err
	}

	response := plugin.Response()
	w.AddCodeGeneratorResponseFiles(response.GetFile()...)
	w.SetSupportedFeatures(uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL) | uint64(pluginpb.CodeGeneratorResponse_FEATURE_SUPPORTS_EDITIONS))
	w.SetFeatureSupportsEditions(descriptorpb.Edition_EDITION_PROTO2, descriptorpb.Edition_EDITION_2024)

	return nil
}

func handleProtogenPlugin(plugin *protogen.Plugin, pluginArgs *args.Args) error {
	// Load our settings
	cfg, err := settings.LoadSettings(pluginArgs.SettingsFilename)
	if err != nil {
		return fmt.Errorf("could not load settings file: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid settings: %w", err)
	}

	// Initializes our output system
	output.Enable(cfg.Debug)

	// Load all addons
	var addons []*addon.Addon
	if cfg.Addons != nil {
		a, err := addon.LoadAddons(cfg.Addons.Path)
		if err != nil {
			return err
		}
		addons = a
	}

	// Build the context and execute templates
	ctx, err := mcontext.BuildContext(mcontext.BuildContextOptions{
		PluginName: pluginArgs.GetPluginName(),
		Settings:   cfg,
		Plugin:     plugin,
		Addons:     addons,
	})
	if err != nil {
		return fmt.Errorf("could not build templates context: %w", err)
	}
	output.Println("processing module:", ctx.ModuleName)

	genTemplates := func(e execution) error {
		templates, err := template.Load(template.Options{
			StrictValidators: true,
			Kind:             e.Kind,
			Path:             e.Path,
			FilesPrefix:      e.Prefix,
			Plugin:           plugin,
			Files:            e.Files,
			Context:          ctx,
			Addons:           addons,
		})
		if err != nil {
			return err
		}

		generated, err := templates.Execute()
		if err != nil {
			return err
		}

		for _, tpl := range generated {
			output.Println("generating source file: ", tpl.Filename)

			content := tpl.Data.String()
			if err := isValidGoSource(content); err != nil {
				return err
			}

			f := plugin.NewGeneratedFile(tpl.Filename, ".")
			f.P(content)
		}

		return nil
	}

	var executions []execution
	if cfg.Templates.Api {
		executions = append(executions, execution{
			Kind:   tpl_types.KindApi,
			Path:   cfg.Templates.ApiPath,
			Prefix: "api",
			Files:  api_tpl_files.Files,
		})
	}
	if cfg.Templates.Test {
		executions = append(executions, execution{
			Kind:   tpl_types.KindTest,
			Path:   cfg.Templates.TestPath,
			Prefix: "testing",
			Files:  test_tpl_files.Files,
		})
	}

	for _, execution := range executions {
		if err := genTemplates(execution); err != nil {
			return fmt.Errorf("could not generate template: %w", err)
		}
	}

	return nil
}

func isValidGoSource(src string) error {
	fileSet := token.NewFileSet()
	_, err := parser.ParseFile(fileSet, "", src, parser.AllErrors)
	if err == nil {
		return nil
	}

	var (
		sb    strings.Builder
		lines = strings.Split(src, "\n")
	)

	sb.WriteString(err.Error() + "\n")
	for i, line := range lines {
		sb.WriteString(fmt.Sprintf("%4d: %s\n", i+1, line))
	}

	return errors.New(sb.String())
}
