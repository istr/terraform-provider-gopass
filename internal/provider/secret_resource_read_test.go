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

func TestSecretResource_Read(t *testing.T) {
	r := &SecretResource{}
	mockStore := newMockStore()
	client := NewGopassClient("")
	client.store = mockStore
	r.client = client

	// Add a secret to the mock store
	mockStore.secrets["test/secret"] = newMockSecret("test-password")
	mockStore.revisions["test/secret"] = []string{"1", "2"}

	ctx := context.Background()
	schemaReq := resource.SchemaRequest{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, schemaReq, schemaResp)

	stateValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":               tftypes.String,
			"path":             tftypes.String,
			"value_wo":         tftypes.String,
			"value_wo_version": tftypes.Number,
			"delete_on_remove": tftypes.Bool,
			"revision_count":   tftypes.Number,
		},
	}, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "test/secret"),
		"path":             tftypes.NewValue(tftypes.String, "test/secret"),
		"value_wo":         tftypes.NewValue(tftypes.String, nil),
		"value_wo_version": tftypes.NewValue(tftypes.Number, nil),
		"delete_on_remove": tftypes.NewValue(tftypes.Bool, true),
		"revision_count":   tftypes.NewValue(tftypes.Number, 1),
	})

	req := resource.ReadRequest{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    stateValue,
		},
	}
	resp := &resource.ReadResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    stateValue,
		},
	}

	r.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %v", resp.Diagnostics)
	}

	// Verify state was updated with new revision count
	var state SecretResourceModel
	resp.Diagnostics.Append(resp.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		t.Errorf("failed to get state: %v", resp.Diagnostics)
	}

	if state.RevisionCount.ValueInt64() != 2 {
		t.Errorf("expected revision count 2, got %d", state.RevisionCount.ValueInt64())
	}
}

func TestSecretResource_Read_ExistsError(t *testing.T) {
	r := &SecretResource{}
	mockStore := newMockStore()
	mockStore.shouldFail = true
	mockStore.failMsg = "exists check error"
	client := NewGopassClient("")
	client.store = mockStore
	r.client = client

	ctx := context.Background()
	schemaReq := resource.SchemaRequest{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, schemaReq, schemaResp)

	stateValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":               tftypes.String,
			"path":             tftypes.String,
			"value_wo":         tftypes.String,
			"value_wo_version": tftypes.Number,
			"delete_on_remove": tftypes.Bool,
			"revision_count":   tftypes.Number,
		},
	}, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "test/secret"),
		"path":             tftypes.NewValue(tftypes.String, "test/secret"),
		"value_wo":         tftypes.NewValue(tftypes.String, nil),
		"value_wo_version": tftypes.NewValue(tftypes.Number, nil),
		"delete_on_remove": tftypes.NewValue(tftypes.Bool, true),
		"revision_count":   tftypes.NewValue(tftypes.Number, 1),
	})

	req := resource.ReadRequest{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    stateValue,
		},
	}
	resp := &resource.ReadResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    stateValue,
		},
	}

	r.Read(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for SecretExists failure")
	}
}

func TestSecretResource_Read_NotFound(t *testing.T) {
	r := &SecretResource{}
	mockStore := newMockStore()
	client := NewGopassClient("")
	client.store = mockStore
	r.client = client

	ctx := context.Background()
	schemaReq := resource.SchemaRequest{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, schemaReq, schemaResp)

	stateValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":               tftypes.String,
			"path":             tftypes.String,
			"value_wo":         tftypes.String,
			"value_wo_version": tftypes.Number,
			"delete_on_remove": tftypes.Bool,
			"revision_count":   tftypes.Number,
		},
	}, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "nonexistent"),
		"path":             tftypes.NewValue(tftypes.String, "nonexistent"),
		"value_wo":         tftypes.NewValue(tftypes.String, nil),
		"value_wo_version": tftypes.NewValue(tftypes.Number, nil),
		"delete_on_remove": tftypes.NewValue(tftypes.Bool, true),
		"revision_count":   tftypes.NewValue(tftypes.Number, 1),
	})

	req := resource.ReadRequest{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    stateValue,
		},
	}
	resp := &resource.ReadResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    stateValue,
		},
	}

	r.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %v", resp.Diagnostics)
	}

	// State should be removed when secret doesn't exist
	if !resp.State.Raw.IsNull() {
		t.Error("expected state to be removed for non-existent secret")
	}
}
