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

	"github.com/rsfreitas/protoc-gen-mikros-extensions/internal/args"
	api_tpl_files "github.com/rsfreitas/protoc-gen-mikros-extensions/internal/assets/api"
	test_tpl_files "github.com/rsfreitas/protoc-gen-mikros-extensions/internal/assets/testing"
	mcontext "github.com/rsfreitas/protoc-gen-mikros-extensions/internal/context"
	"github.com/rsfreitas/protoc-gen-mikros-extensions/internal/output"
	"github.com/rsfreitas/protoc-gen-mikros-extensions/internal/settings"
	"github.com/rsfreitas/protoc-gen-mikros-extensions/internal/template"
)

type execution struct {
	Path   string
	Prefix string
	File   embed.FS
}

func Handle(
	_ context.Context,
	_ protoplugin.PluginEnv,
	w protoplugin.ResponseWriter,
	r protoplugin.Request,
) error {
	plugin, err := protogen.Options{}.New(r.CodeGeneratorRequest())
	if err != nil {
		return err
	}

	plugin.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL) | uint64(pluginpb.CodeGeneratorResponse_FEATURE_SUPPORTS_EDITIONS)
	plugin.SupportedEditionsMinimum = descriptorpb.Edition_EDITION_PROTO2
	plugin.SupportedEditionsMaximum = descriptorpb.Edition_EDITION_2023

	pluginArgs, err := args.NewArgsFromString(r.Parameter())
	if err != nil {
		return err
	}
	if err := pluginArgs.Validate(); err != nil {
		return fmt.Errorf("could not load plugin arguments: %w", err)
	}

	if err := handleProtogenPlugin(plugin, pluginArgs); err != nil {
		return err
	}

	response := plugin.Response()
	w.AddCodeGeneratorResponseFiles(response.GetFile()...)

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

	// Build the context and execute templates
	ctx, err := mcontext.BuildContext(mcontext.BuildContextOptions{
		PluginName: pluginArgs.GetPluginName(),
		Settings:   cfg,
		Plugin:     plugin,
	})
	if err != nil {
		return fmt.Errorf("could not build templates context: %w", err)
	}
	output.Println("processing module:", ctx.ModuleName)

	genTemplates := func(path, prefix string, files embed.FS) error {
		templates, err := template.LoadTemplates(template.Options{
			StrictValidators: true,
			Path:             path,
			FilesPrefix:      prefix,
			Plugin:           plugin,
			Files:            files,
			Context:          ctx,
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
			Path:   cfg.Templates.ApiPath,
			Prefix: "api",
			File:   api_tpl_files.Files,
		})
	}
	if cfg.Templates.Test {
		executions = append(executions, execution{
			Path:   cfg.Templates.TestPath,
			Prefix: "testing",
			File:   test_tpl_files.Files,
		})
	}

	for _, execution := range executions {
		if err := genTemplates(execution.Path, execution.Prefix, execution.File); err != nil {
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
