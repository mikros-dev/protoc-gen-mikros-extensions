// Package addon defines the public contracts and helpers required to build
// extensible addons for protoc-gen-mikros-extensions.
//
// # Overview
//
// An addon integrates custom code-generation behavior into the plugin. It
// declares a unique name, exposes its template files, specifies the kind of
// templates it generates, provides template-specific imports and context,
// and participates in template validation.
//
// Key concepts
//
//   - Addon: core interface an addon must implement to be discovered and used
//     by the generator (name, templates, kind, imports, context, validation).
//
//   - OutboundExtension: an optional interface that lets an addon inject custom
//     code into generated “IntoOutbound” conversion functions.
//
//   - Import: describes template imports (with optional alias) that an addon
//     may contribute during code generation.
//
//   - The extensions subpackage: utilities for working with protobuf option
//     extensions through reflection, commonly used by addons to detect and
//     retrieve custom options from descriptors.
package addon
