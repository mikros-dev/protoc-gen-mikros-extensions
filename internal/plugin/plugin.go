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
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/ctxutil"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/log"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/template"
	tpl_context "github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/template/context"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/template/spec"
)

type execution struct {
	Kind   spec.Kind
	Path   string
	Prefix string
	Files  embed.FS
}

// Handle is the entrypoint of the plugin.
func Handle(
	ctx context.Context,
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

	if err := handleProtogenPlugin(ctx, plugin, pluginArgs); err != nil {
		return err
	}

	response := plugin.Response()
	w.AddCodeGeneratorResponseFiles(response.GetFile()...)
	w.SetSupportedFeatures(
		uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL) |
			uint64(pluginpb.CodeGeneratorResponse_FEATURE_SUPPORTS_EDITIONS),
	)

	w.SetFeatureSupportsEditions(descriptorpb.Edition_EDITION_PROTO2, descriptorpb.Edition_EDITION_2024)

	return nil
}

func handleProtogenPlugin(ctx context.Context, plugin *protogen.Plugin, pluginArgs *args.Args) error {
	cfg, addons, err := loadConfigAndAddons(pluginArgs)
	if err != nil {
		return err
	}

	// Initializes our output system
	logger := log.New(log.LoggerOptions{
		Verbose: cfg.Debug,
	})
	ctx = ctxutil.WithLogger(ctx, logger)

	// Build the context and execute templates
	tplContext, err := tpl_context.BuildContext(tpl_context.BuildContextOptions{
		PluginName: pluginArgs.GetPluginName(),
		Settings:   cfg,
		Plugin:     plugin,
		Addons:     addons,
	})
	if err != nil {
		return fmt.Errorf("could not build templates context: %w", err)
	}
	logger.Println("processing module:", tplContext.ModuleName)

	executions := buildExecutions(cfg)
	for _, execution := range executions {
		if err := generateTemplates(ctx, plugin, tplContext, addons, execution); err != nil {
			return fmt.Errorf("could not generate template: %w", err)
		}
	}

	return nil
}

func loadConfigAndAddons(pluginArgs *args.Args) (*settings.Settings, []*addon.Addon, error) {
	cfg, err := settings.LoadSettings(pluginArgs.SettingsFilename)
	if err != nil {
		return nil, nil, fmt.Errorf("could not load settings file: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, nil, fmt.Errorf("invalid settings: %w", err)
	}

	var addonsList []*addon.Addon
	if cfg.Addons != nil {
		a, err := addon.LoadAddons(cfg.Addons.Path)
		if err != nil {
			return nil, nil, err
		}
		addonsList = a
	}

	return cfg, addonsList, nil
}

func buildExecutions(cfg *settings.Settings) []execution {
	var executions []execution
	if cfg.Templates.API {
		executions = append(executions, execution{
			Kind:   spec.KindAPI,
			Path:   cfg.Templates.APIPath,
			Prefix: "api",
			Files:  api_tpl_files.Files,
		})
	}
	if cfg.Templates.Test {
		executions = append(executions, execution{
			Kind:   spec.KindTest,
			Path:   cfg.Templates.TestPath,
			Prefix: "testing",
			Files:  test_tpl_files.Files,
		})
	}

	return executions
}

func generateTemplates(
	ctx context.Context,
	plugin *protogen.Plugin,
	tplContext *tpl_context.Context,
	addons []*addon.Addon,
	e execution,
) error {
	logger := ctxutil.LoggerFromContext(ctx)
	templates, err := template.Load(template.Options{
		StrictValidators: true,
		Kind:             e.Kind,
		Path:             e.Path,
		FilesPrefix:      e.Prefix,
		Plugin:           plugin,
		Files:            e.Files,
		Context:          tplContext,
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
		logger.Println("generating source file: ", tpl.Filename)

		content := tpl.Data.String()
		if err := isValidGoSource(content); err != nil {
			return err
		}

		f := plugin.NewGeneratedFile(tpl.Filename, ".")
		f.P(content)
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

	_, _ = sb.WriteString(err.Error() + "\n")
	for i, line := range lines {
		_, _ = sb.WriteString(fmt.Sprintf("%4d: %s\n", i+1, line))
	}

	return errors.New(sb.String())
}
