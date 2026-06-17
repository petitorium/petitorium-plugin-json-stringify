// Package main provides the json-stringify plugin for Petitorium.
// This plugin finds JSON object keys ending with {{json-stringify:wrap}} and
// replaces them with their stringified JSON value.
//
// Tag syntax inside a JSON key:
//
//	"serviceData{{json-stringify:wrap}}": { ... }
//	"serviceData{{json-stringify:wrap indent="false"}}": { ... }
//
// To build: CGO_ENABLED=0 go build -o json-stringify .
package main

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/hashicorp/go-plugin"

	"github.com/petitorium/petitorium-plugin-sdk/shared"
	"github.com/petitorium/petitorium-plugin-sdk/types"
)

// tagSuffixRegex matches the json-stringify tag at the end of a JSON key.
// It tolerates optional whitespace and parameters.
var tagSuffixRegex = regexp.MustCompile(`\{\{\s*json-stringify:wrap\s*(?:\s+\w+="[^"]*")*\s*\}\}`)

// paramRegex extracts key="value" pairs from a tag string.
var paramRegex = regexp.MustCompile(`(\w+)="([^"]*)"`)

// JSONStringify is a plugin that stringifies JSON values into escaped strings.
type JSONStringify struct{}

// Name returns the plugin name.
func (p *JSONStringify) Name() string {
	return "json-stringify"
}

// Version returns the plugin version.
func (p *JSONStringify) Version() string {
	return "1.0.0"
}

// Description returns the plugin description.
func (p *JSONStringify) Description() string {
	return "Stringifies JSON object values into escaped JSON strings"
}

// Hooks returns the hook types this plugin implements.
func (p *JSONStringify) Hooks() []types.HookType {
	return []types.HookType{types.PreSend}
}

// ExecuteHook executes a specific hook with the given context.
func (p *JSONStringify) ExecuteHook(hookType types.HookType, ctx *types.HookContext) (*types.HookContext, error) {
	if hookType != types.PreSend {
		return ctx, nil
	}
	if ctx.Request == nil || ctx.Request.Body == "" {
		return ctx, nil
	}

	var body interface{}
	if err := json.Unmarshal([]byte(ctx.Request.Body), &body); err != nil {
		// Body is not valid JSON; leave it untouched.
		return ctx, nil
	}

	processed := processValue(body)

	b, err := json.Marshal(processed)
	if err != nil {
		return ctx, fmt.Errorf("json-stringify: failed to marshal body: %w", err)
	}

	ctx.Request.Body = string(b)
	return ctx, nil
}

// processValue recursively walks a JSON structure and stringifies values
// whose keys end with the json-stringify tag suffix.
func processValue(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		// First recurse into values so nested tags are resolved bottom-up.
		for k, vv := range val {
			val[k] = processValue(vv)
		}

		result := make(map[string]interface{}, len(val))
		for k, vv := range val {
			baseKey, params, ok := extractTagInfo(k)
			if ok {
				str, err := stringifyValue(vv, params)
				if err != nil {
					// On error preserve the original key/value.
					result[k] = vv
					continue
				}
				result[baseKey] = str
				continue
			}
			result[k] = vv
		}
		return result

	case []interface{}:
		for i, vv := range val {
			val[i] = processValue(vv)
		}
		return val

	default:
		return v
	}
}

// extractTagInfo checks whether a JSON key ends with the json-stringify tag.
// It returns the base key (everything before the tag), the parsed parameters,
// and a boolean indicating a match.
func extractTagInfo(key string) (string, map[string]string, bool) {
	loc := tagSuffixRegex.FindStringIndex(key)
	if loc == nil || loc[1] != len(key) {
		return "", nil, false
	}
	baseKey := key[:loc[0]]
	tagPart := key[loc[0]:loc[1]]
	return baseKey, parseParams(tagPart), true
}

// stringifyValue marshals a value to a JSON string.
// If params["indent"] == "false" the output is compact; otherwise it is
// pretty-printed with tabs.
func stringifyValue(v interface{}, params map[string]string) (string, error) {
	indent := params["indent"]
	if indent == "false" {
		b, err := json.Marshal(v)
		return string(b), err
	}
	b, err := json.MarshalIndent(v, "", "\t")
	return string(b), err
}

// parseParams extracts key="value" pairs from a raw tag string.
func parseParams(rawTag string) map[string]string {
	params := make(map[string]string)
	matches := paramRegex.FindAllStringSubmatch(rawTag, -1)
	for _, m := range matches {
		if len(m) == 3 {
			params[m[1]] = m[2]
		}
	}
	return params
}

// GetTagDetails implements types.TagEditorCapable.
func (p *JSONStringify) GetTagDetails(rawTag, context string) (*types.TagDetailsResponse, error) {
	params := parseParams(rawTag)
	indent := params["indent"]
	if indent == "" {
		indent = "true"
	}

	return &types.TagDetailsResponse{
		DisplayLabel: "JSON Stringify",
		PluginName:   "json-stringify",
		Action:       "wrap",
		Editable:     true,
		Schema: &types.TagEditorSchema{
			Fields: []types.TagField{
				{
					Key:          "indent",
					Label:        "Pretty Print (Indent)",
					FieldType:    "checkbox",
					DefaultValue: indent,
				},
			},
		},
	}, nil
}

// UpdateTag implements types.TagEditorCapable.
func (p *JSONStringify) UpdateTag(rawTag string, values map[string]string) (*types.UpdateTagResponse, error) {
	indent := values["indent"]
	if indent == "" {
		indent = "true"
	}

	newTag := fmt.Sprintf(`{{json-stringify:wrap indent="%s"}}`, indent)
	return &types.UpdateTagResponse{NewRawTag: newTag}, nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		Plugins: map[string]plugin.Plugin{
			"json-stringify": &shared.PetitoriumPlugin{Impl: &JSONStringify{}},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}

// Compile-time interface checks.
var (
	_ types.Plugin           = &JSONStringify{}
	_ types.TagEditorCapable = &JSONStringify{}
)
