# Petitorium JSON Stringify Plugin

A [Petitorium](https://github.com/petitorium/petitorium) plugin that stringifies JSON object values into escaped JSON strings before a request is sent.

## Tag Syntax

Append the tag to a JSON object key to stringify its value:

```json
{
  "serviceData{{json-stringify:wrap}}": {
    "serviceId": "12345",
    "name": "Example Service"
  }
}
```

### Parameters

| Parameter | Type   | Required | Description                                           |
| --------- | ------ | -------- | ----------------------------------------------------- |
| `indent`  | string | No       | `true` (default) for pretty-printed, `false` for compact |

### Examples

**Default pretty-printed output:**
```json
"serviceData{{json-stringify:wrap}}": { "id": 1 }
```

**Compact output:**
```json
"serviceData{{json-stringify:wrap indent=\"false\"}}": { "id": 1 }
```

## How It Works

1. The plugin hooks into the `PreSend` lifecycle stage.
2. It parses the request body as JSON and recursively walks the structure.
3. Any object key ending with `{{json-stringify:wrap}}` (or `{{json-stringify:wrap indent="false"}}`) is processed:
   - The suffix is stripped to obtain the base key.
   - The key's value is marshaled to a JSON string.
   - The base key is replaced with the stringified value.
4. The modified body is sent to the endpoint.

### Example Transformation

**Before:**
```json
{
  "favorite": false,
  "serviceData{{json-stringify:wrap}}": {
    "serviceId": "12345",
    "name": "Example Service",
    "logoList": ["logo.png"],
    "formList": []
  }
}
```

**After:**
```json
{
  "favorite": false,
  "serviceData": "{\n\t\"serviceId\": \"12345\",\n\t\"name\": \"Example Service\",\n\t\"logoList\": [\n\t\t\"logo.png\"\n\t],\n\t\"formList\": []\n}"
}
```

## Building

```bash
CGO_ENABLED=0 go build -o json-stringify .
```

## Installation

1. Build the plugin (see above).
2. Copy the binary to your Petitorium plugins directory:

```bash
cp json-stringify ~/.config/petitorium/plugins/available/
```

3. Enable it in your Petitorium configuration (`~/.config/petitorium/config.yaml`):

```yaml
plugins:
  enabled:
    - json-stringify
```

## Tag Editor Support

This plugin implements the `TagEditorCapable` interface, allowing Petitorium to render a dynamic form when editing `json-stringify` tags.

### Form Fields

| Field                | Type     | Description                        |
| -------------------- | -------- | ---------------------------------- |
| Pretty Print (Indent)| checkbox | Enable pretty-printed JSON output  |

## Plugin Hooks

| Hook      | Purpose                                           |
| --------- | ------------------------------------------------- |
| `PreSend` | Stringifies tagged JSON values before the request is sent |

## Development

```bash
# Download dependencies
go mod tidy

# Build
go build -o json-stringify .

# Run tests
go test -v ./...
```

## Troubleshooting

### Plugin Not Loading

1. Verify the executable exists: `ls -la ~/.config/petitorium/plugins/available/json-stringify`
2. Ensure execution permissions: `chmod +x ~/.config/petitorium/plugins/available/json-stringify`
3. Confirm the plugin is listed under `plugins.enabled` in `config.yaml`

### Tag Not Being Processed

1. Ensure the tag is placed **inside** the JSON key name, e.g. `"key{{json-stringify:wrap}}": value`
2. Check that the request body is valid JSON; malformed JSON is left untouched
3. Verify the plugin is loaded and enabled (check Petitorium startup logs)

## License

MIT License
