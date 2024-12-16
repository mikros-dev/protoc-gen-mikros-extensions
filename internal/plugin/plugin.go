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
	"github.com/mikros-dev/protoc-gen-mikros-extensions/internal/template"
	go_tpl_files "github.com/mikros-dev/protoc-gen-mikros-extensions/internal/template/golang"
	rust_tpl_files "github.com/mikros-dev/protoc-gen-mikros-extensions/internal/template/rust"
	test_tpl_files "github.com/mikros-dev/protoc-gen-mikros-extensions/internal/template/testing"
	mcontext "github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/context"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/output"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
	mtemplate "github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/template"
)

type execution struct {
	SingleModule bool
	Kind         mtemplate.Kind
	Path         string
	ModuleName   string
	Files        embed.FS
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
		templates, err := template.LoadTemplates(template.LoadTemplatesOptions{
			StrictValidators: true,
			Kind:             e.Kind,
			Plugin:           plugin,
			Files:            e.Files,
			Addons:           addons,
		})
		if err != nil {
			return err
		}

		generated, err := templates.Execute(template.ExecuteOptions{
			Context:      ctx,
			Path:         e.Path,
			SingleModule: e.SingleModule,
			ModuleName:   e.ModuleName,
		})
		if err != nil {
			return err
		}

		for _, tpl := range generated {
			output.Println("generating source file: ", tpl.Filename)
			content := tpl.Data.String()

			if e.Kind != mtemplate.KindRust {
				if err := isValidGoSource(content); err != nil {
					return err
				}
			}

			f := plugin.NewGeneratedFile(tpl.Filename, ".")
			f.P(content)
		}

		return nil
	}

	var executions []execution
	if cfg.Templates.Go {
		executions = append(executions, execution{
			Kind:  mtemplate.KindGo,
			Path:  cfg.Templates.GoPath,
			Files: go_tpl_files.Files,
		})
	}
	if cfg.Templates.Test {
		executions = append(executions, execution{
			Kind:  mtemplate.KindTest,
			Path:  cfg.Templates.TestPath,
			Files: test_tpl_files.Files,
		})
	}
	if cfg.Templates.RustEnabled() {
		executions = append(executions, execution{
			Kind:         mtemplate.KindRust,
			Path:         cfg.Templates.Rust.Path,
			Files:        rust_tpl_files.Files,
			SingleModule: cfg.Templates.Rust.SingleModule,
			ModuleName:   cfg.Templates.Rust.ModuleName,
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
