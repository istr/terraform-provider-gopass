// Copyright (c) Ingo Struck
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestSecretResource_Create(t *testing.T) {
	r := &SecretResource{}
	mockStore := newMockStore()
	client := NewGopassClient("")
	client.store = mockStore
	r.client = client

	// Create the schema for our test
	ctx := context.Background()
	schemaReq := resource.SchemaRequest{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, schemaReq, schemaResp)

	// Create plan and config values
	planValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":               tftypes.String,
			"path":             tftypes.String,
			"value_wo":         tftypes.String,
			"value_wo_version": tftypes.Number,
			"delete_on_remove": tftypes.Bool,
			"revision_count":   tftypes.Number,
		},
	}, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"path":             tftypes.NewValue(tftypes.String, "test/secret"),
		"value_wo":         tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"value_wo_version": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		"delete_on_remove": tftypes.NewValue(tftypes.Bool, true),
		"revision_count":   tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
	})

	configValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":               tftypes.String,
			"path":             tftypes.String,
			"value_wo":         tftypes.String,
			"value_wo_version": tftypes.Number,
			"delete_on_remove": tftypes.Bool,
			"revision_count":   tftypes.Number,
		},
	}, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, nil),
		"path":             tftypes.NewValue(tftypes.String, "test/secret"),
		"value_wo":         tftypes.NewValue(tftypes.String, "test-password"),
		"value_wo_version": tftypes.NewValue(tftypes.Number, 1),
		"delete_on_remove": tftypes.NewValue(tftypes.Bool, true),
		"revision_count":   tftypes.NewValue(tftypes.Number, nil),
	})

	req := resource.CreateRequest{
		Plan: tfsdk.Plan{
			Schema: schemaResp.Schema,
			Raw:    planValue,
		},
		Config: tfsdk.Config{
			Schema: schemaResp.Schema,
			Raw:    configValue,
		},
	}
	resp := &resource.CreateResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
		},
	}

	r.Create(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %v", resp.Diagnostics)
	}

	// Verify secret was stored
	if _, exists := mockStore.secrets["test/secret"]; !exists {
		t.Error("expected secret to be stored in mock store")
	}

	// Verify state was set
	var state SecretResourceModel
	resp.Diagnostics.Append(resp.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		t.Errorf("failed to get state: %v", resp.Diagnostics)
	}

	if state.Path.ValueString() != "test/secret" {
		t.Errorf("expected path 'test/secret', got %q", state.Path.ValueString())
	}
	if state.ID.ValueString() != "test/secret" {
		t.Errorf("expected ID 'test/secret', got %q", state.ID.ValueString())
	}
}

func TestSecretResource_Create_NoValueWO(t *testing.T) {
	r := &SecretResource{}
	mockStore := newMockStore()
	client := NewGopassClient("")
	client.store = mockStore
	r.client = client

	ctx := context.Background()
	schemaReq := resource.SchemaRequest{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, schemaReq, schemaResp)

	planValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":               tftypes.String,
			"path":             tftypes.String,
			"value_wo":         tftypes.String,
			"value_wo_version": tftypes.Number,
			"delete_on_remove": tftypes.Bool,
			"revision_count":   tftypes.Number,
		},
	}, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"path":             tftypes.NewValue(tftypes.String, "test/secret"),
		"value_wo":         tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"value_wo_version": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		"delete_on_remove": tftypes.NewValue(tftypes.Bool, true),
		"revision_count":   tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
	})

	configValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":               tftypes.String,
			"path":             tftypes.String,
			"value_wo":         tftypes.String,
			"value_wo_version": tftypes.Number,
			"delete_on_remove": tftypes.Bool,
			"revision_count":   tftypes.Number,
		},
	}, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, nil),
		"path":             tftypes.NewValue(tftypes.String, "test/secret"),
		"value_wo":         tftypes.NewValue(tftypes.String, nil), // No value provided
		"value_wo_version": tftypes.NewValue(tftypes.Number, nil),
		"delete_on_remove": tftypes.NewValue(tftypes.Bool, true),
		"revision_count":   tftypes.NewValue(tftypes.Number, nil),
	})

	req := resource.CreateRequest{
		Plan: tfsdk.Plan{
			Schema: schemaResp.Schema,
			Raw:    planValue,
		},
		Config: tfsdk.Config{
			Schema: schemaResp.Schema,
			Raw:    configValue,
		},
	}
	resp := &resource.CreateResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
		},
	}

	r.Create(ctx, req, resp)

	// Should succeed but with a warning
	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %v", resp.Diagnostics)
	}

	// Should have warning about no value provided
	hasWarning := false
	for _, diag := range resp.Diagnostics {
		if diag.Summary() == "No value provided" {
			hasWarning = true
			break
		}
	}
	if !hasWarning {
		t.Error("expected warning about no value provided")
	}
}
