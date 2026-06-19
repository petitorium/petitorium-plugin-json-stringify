package main

import (
	"encoding/json"
	"testing"

	"github.com/petitorium/petitorium-plugin-sdk/types"
)

func TestExecuteHook_ObjectRoot(t *testing.T) {
	p := &JSONStringify{}
	ctx := &types.HookContext{
		Request: &types.RequestData{
			Body: `{"favorite":false,"serviceData{{json-stringify:wrap}}":{"serviceId":"133","name":"Teletón Donación"}}`,
		},
	}

	result, err := p.ExecuteHook(types.PreSend, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var body map[string]interface{}
	if err := json.Unmarshal([]byte(result.Request.Body), &body); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	serviceData, ok := body["serviceData"].(string)
	if !ok {
		t.Fatalf("expected serviceData to be string, got %T", body["serviceData"])
	}

	var inner map[string]interface{}
	if err := json.Unmarshal([]byte(serviceData), &inner); err != nil {
		t.Fatalf("serviceData is not valid JSON: %v", err)
	}
	if inner["serviceId"] != "133" {
		t.Errorf("expected serviceId=133, got %v", inner["serviceId"])
	}
	if inner["name"] != "Teletón Donación" {
		t.Errorf("expected name='Teletón Donación', got %v", inner["name"])
	}
}

func TestExecuteHook_ArrayRoot(t *testing.T) {
	p := &JSONStringify{}
	ctx := &types.HookContext{
		Request: &types.RequestData{
			Body: `[{"item{{json-stringify:wrap}}":{"id":1}}]`,
		},
	}

	result, err := p.ExecuteHook(types.PreSend, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var body []map[string]interface{}
	if err := json.Unmarshal([]byte(result.Request.Body), &body); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if len(body) != 1 {
		t.Fatalf("expected 1 element, got %d", len(body))
	}

	item, ok := body[0]["item"].(string)
	if !ok {
		t.Fatalf("expected item to be string, got %T", body[0]["item"])
	}

	var inner map[string]interface{}
	if err := json.Unmarshal([]byte(item), &inner); err != nil {
		t.Fatalf("item is not valid JSON: %v", err)
	}
	if inner["id"] != float64(1) {
		t.Errorf("expected id=1, got %v", inner["id"])
	}
}

func TestExecuteHook_NestedObject(t *testing.T) {
	p := &JSONStringify{}
	ctx := &types.HookContext{
		Request: &types.RequestData{
			Body: `{"outer":{"inner{{json-stringify:wrap}}":{"deep":true}}}`,
		},
	}

	result, err := p.ExecuteHook(types.PreSend, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var body map[string]interface{}
	if err := json.Unmarshal([]byte(result.Request.Body), &body); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	outer, ok := body["outer"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected outer to be object, got %T", body["outer"])
	}

	inner, ok := outer["inner"].(string)
	if !ok {
		t.Fatalf("expected inner to be string, got %T", outer["inner"])
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(inner), &parsed); err != nil {
		t.Fatalf("inner is not valid JSON: %v", err)
	}
	if parsed["deep"] != true {
		t.Errorf("expected deep=true, got %v", parsed["deep"])
	}
}

func TestExecuteHook_NoTag(t *testing.T) {
	p := &JSONStringify{}
	original := `{"favorite":false,"serviceData":{"serviceId":"133"}}`
	ctx := &types.HookContext{
		Request: &types.RequestData{
			Body: original,
		},
	}

	result, err := p.ExecuteHook(types.PreSend, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Request.Body != original {
		t.Errorf("body should remain unchanged.\ngot:  %s\nwant: %s", result.Request.Body, original)
	}
}

func TestExecuteHook_WhitespaceVariants(t *testing.T) {
	variants := []string{
		`{"data{{json-stringify:wrap}}":{"a":1}}`,
		`{"data{{ json-stringify:wrap }}":{"a":1}}`,
		`{"data{{json-stringify:wrap  }}":{"a":1}}`,
	}

	for i, body := range variants {
		p := &JSONStringify{}
		ctx := &types.HookContext{
			Request: &types.RequestData{Body: body},
		}

		result, err := p.ExecuteHook(types.PreSend, ctx)
		if err != nil {
			t.Fatalf("variant %d: unexpected error: %v", i, err)
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(result.Request.Body), &parsed); err != nil {
			t.Fatalf("variant %d: invalid JSON output: %v", i, err)
		}

		data, ok := parsed["data"].(string)
		if !ok {
			t.Fatalf("variant %d: expected data to be string, got %T", i, parsed["data"])
		}

		var inner map[string]interface{}
		if err := json.Unmarshal([]byte(data), &inner); err != nil {
			t.Fatalf("variant %d: data is not valid JSON: %v", i, err)
		}
		if inner["a"] != float64(1) {
			t.Errorf("variant %d: expected a=1, got %v", i, inner["a"])
		}
	}
}

func TestExecuteHook_CompactOutput(t *testing.T) {
	p := &JSONStringify{}
	ctx := &types.HookContext{
		Request: &types.RequestData{
			Body: `{"data{{json-stringify:wrap indent=\"false\"}}":{"a":1,"b":2}}`,
		},
	}

	result, err := p.ExecuteHook(types.PreSend, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var body map[string]interface{}
	if err := json.Unmarshal([]byte(result.Request.Body), &body); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	data, ok := body["data"].(string)
	if !ok {
		t.Fatalf("expected data to be string, got %T", body["data"])
	}

	// Compact JSON should not contain whitespace.
	if data != `{"a":1,"b":2}` {
		t.Errorf("unexpected compact output: %s", data)
	}
}

func TestExecuteHook_NestedArray(t *testing.T) {
	p := &JSONStringify{}
	ctx := &types.HookContext{
		Request: &types.RequestData{
			Body: `{"items":[{"detail{{json-stringify:wrap}}":{"x":"y"}}]}`,
		},
	}

	result, err := p.ExecuteHook(types.PreSend, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var body map[string]interface{}
	if err := json.Unmarshal([]byte(result.Request.Body), &body); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	items, ok := body["items"].([]interface{})
	if !ok {
		t.Fatalf("expected items to be array, got %T", body["items"])
	}

	item, ok := items[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected item to be object, got %T", items[0])
	}

	detail, ok := item["detail"].(string)
	if !ok {
		t.Fatalf("expected detail to be string, got %T", item["detail"])
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(detail), &parsed); err != nil {
		t.Fatalf("detail is not valid JSON: %v", err)
	}
	if parsed["x"] != "y" {
		t.Errorf("expected x=y, got %v", parsed["x"])
	}
}

func TestExecuteHook_PreservesOtherKeys(t *testing.T) {
	p := &JSONStringify{}
	ctx := &types.HookContext{
		Request: &types.RequestData{
			Body: `{"a":1,"b{{json-stringify:wrap}}":{"c":3},"d":4}`,
		},
	}

	result, err := p.ExecuteHook(types.PreSend, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var body map[string]interface{}
	if err := json.Unmarshal([]byte(result.Request.Body), &body); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if body["a"] != float64(1) {
		t.Errorf("expected a=1, got %v", body["a"])
	}
	if body["d"] != float64(4) {
		t.Errorf("expected d=4, got %v", body["d"])
	}
}

func TestGetTagDetails(t *testing.T) {
	p := &JSONStringify{}
	res, err := p.GetTagDetails(`{{json-stringify:wrap}}`, "body", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.DisplayLabel != "JSON Stringify" {
		t.Errorf("unexpected label: %s", res.DisplayLabel)
	}
	if !res.Editable {
		t.Error("expected editable")
	}
	if len(res.Schema.Fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(res.Schema.Fields))
	}
	if res.Schema.Fields[0].DefaultValue != "true" {
		t.Errorf("expected default indent=true, got %s", res.Schema.Fields[0].DefaultValue)
	}
}

func TestUpdateTag(t *testing.T) {
	p := &JSONStringify{}
	res, err := p.UpdateTag(`{{json-stringify:wrap}}`, map[string]string{
		"indent": "false",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := `{{json-stringify:wrap indent="false"}}`
	if res.NewRawTag != want {
		t.Errorf("unexpected tag: got %s, want %s", res.NewRawTag, want)
	}
}
