// Copyright (c) Ingo Struck
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

func TestSecretResource_Create_SetSecretError(t *testing.T) {
	r := &SecretResource{}
	mockStore := newMockStore()
	mockStore.shouldFail = true
	mockStore.failMsg = "failed to write to store"
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

	// Should have error from SetSecret
	if !resp.Diagnostics.HasError() {
		t.Error("expected error from SetSecret but got none")
	}

	// Check error message
	hasExpectedError := false
	for _, diag := range resp.Diagnostics {
		if diag.Summary() == "Failed to create secret" {
			hasExpectedError = true
			break
		}
	}
	if !hasExpectedError {
		t.Error("expected 'Failed to create secret' error")
	}
}

// Local mock store wrapper to force Get failure
type failGetStore struct {
	*mockStore
}

func (m *failGetStore) Get(ctx context.Context, name, revision string) (gopass.Secret, error) {
	if name == "test/secret-error" {
		return nil, fmt.Errorf("forced failure for %s", name)
	}
	return m.mockStore.Get(ctx, name, revision)
}

func TestSecretResource_Create_GetRevisionCountError(t *testing.T) {
	r := &SecretResource{}
	// Use local wrapper to fail Get (used by GetRevisionCount)
	// but allow Set (used by Create)
	baseStore := newMockStore()
	mockStore := &failGetStore{mockStore: baseStore}

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
		"path":             tftypes.NewValue(tftypes.String, "test/secret-error"),
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
		"path":             tftypes.NewValue(tftypes.String, "test/secret-error"),
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

	// Should succeed (GetRevisionCount error is non-fatal)
	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %v", resp.Diagnostics)
	}

	// Verify state was set with revision count 0 (because we removed the fallback to 1)
	var state SecretResourceModel
	resp.Diagnostics.Append(resp.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		t.Errorf("failed to get state: %v", resp.Diagnostics)
	}

	if state.RevisionCount.ValueInt64() != 0 {
		t.Errorf("expected revision count 0 (error), got %d", state.RevisionCount.ValueInt64())
	}
}

func TestSecretResource_Create_PlanGetError(t *testing.T) {
	r := &SecretResource{}
	client := NewGopassClient("")
	r.client = client
	ctx := context.Background()

	// Define a schema that is incompatible with SecretResourceModel
	// SecretResourceModel expects 'path' to be a String, but we define it as Int64
	incompatibleSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"path": schema.Int64Attribute{
				Required: true,
			},
			"id":               schema.StringAttribute{Computed: true},
			"value_wo":         schema.StringAttribute{Optional: true},
			"value_wo_version": schema.Int64Attribute{Optional: true},
			"delete_on_remove": schema.BoolAttribute{Optional: true},
			"revision_count":   schema.Int64Attribute{Computed: true},
		},
	}

	// Create a plan value that matches the schema (int64)
	planValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"path":             tftypes.Number,
			"id":               tftypes.String,
			"value_wo":         tftypes.String,
			"value_wo_version": tftypes.Number,
			"delete_on_remove": tftypes.Bool,
			"revision_count":   tftypes.Number,
		},
	}, map[string]tftypes.Value{
		"path":             tftypes.NewValue(tftypes.Number, 123),
		"id":               tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"value_wo":         tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"value_wo_version": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		"delete_on_remove": tftypes.NewValue(tftypes.Bool, true),
		"revision_count":   tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
	})

	req := resource.CreateRequest{
		Plan: tfsdk.Plan{
			Schema: incompatibleSchema,
			Raw:    planValue,
		},
		Config: tfsdk.Config{
			Schema: incompatibleSchema,
			Raw:    planValue,
		},
	}
	resp := &resource.CreateResponse{
		State: tfsdk.State{
			Schema: incompatibleSchema,
		},
	}

	r.Create(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error from Plan.Get but got none")
	}
}

func TestSecretResource_Create_ConfigGetError(t *testing.T) {
	r := &SecretResource{}
	client := NewGopassClient("")
	r.client = client
	ctx := context.Background()

	// 1. Create a VALID schema and value for Plan (so Plan.Get succeeds)
	validSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"path":             schema.StringAttribute{Required: true},
			"id":               schema.StringAttribute{Computed: true},
			"value_wo":         schema.StringAttribute{Optional: true},
			"value_wo_version": schema.Int64Attribute{Optional: true},
			"delete_on_remove": schema.BoolAttribute{Optional: true},
			"revision_count":   schema.Int64Attribute{Computed: true},
		},
	}

	validPlanValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"path":             tftypes.String,
			"id":               tftypes.String,
			"value_wo":         tftypes.String,
			"value_wo_version": tftypes.Number,
			"delete_on_remove": tftypes.Bool,
			"revision_count":   tftypes.Number,
		},
	}, map[string]tftypes.Value{
		"path":             tftypes.NewValue(tftypes.String, "some/path"),
		"id":               tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"value_wo":         tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"value_wo_version": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		"delete_on_remove": tftypes.NewValue(tftypes.Bool, true),
		"revision_count":   tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
	})

	// 2. Create an INCOMPATIBLE schema and value for Config (so Config.Get fails)
	incompatibleSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"path":             schema.Int64Attribute{Required: true},
			"id":               schema.StringAttribute{Computed: true},
			"value_wo":         schema.StringAttribute{Optional: true},
			"value_wo_version": schema.Int64Attribute{Optional: true},
			"delete_on_remove": schema.BoolAttribute{Optional: true},
			"revision_count":   schema.Int64Attribute{Computed: true},
		},
	}

	incompatibleConfigValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"path":             tftypes.Number,
			"id":               tftypes.String,
			"value_wo":         tftypes.String,
			"value_wo_version": tftypes.Number,
			"delete_on_remove": tftypes.Bool,
			"revision_count":   tftypes.Number,
		},
	}, map[string]tftypes.Value{
		"path":             tftypes.NewValue(tftypes.Number, 123),
		"id":               tftypes.NewValue(tftypes.String, nil),
		"value_wo":         tftypes.NewValue(tftypes.String, nil),
		"value_wo_version": tftypes.NewValue(tftypes.Number, nil),
		"delete_on_remove": tftypes.NewValue(tftypes.Bool, true),
		"revision_count":   tftypes.NewValue(tftypes.Number, nil),
	})

	req := resource.CreateRequest{
		Plan: tfsdk.Plan{
			Schema: validSchema,
			Raw:    validPlanValue,
		},
		Config: tfsdk.Config{
			Schema: incompatibleSchema,
			Raw:    incompatibleConfigValue,
		},
	}
	resp := &resource.CreateResponse{
		State: tfsdk.State{
			Schema: validSchema,
		},
	}

	r.Create(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error from Config.Get but got none")
	}
}
