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

func TestSecretResource_Delete(t *testing.T) {
	r := &SecretResource{}
	mockStore := newMockStore()
	client := NewGopassClient("")
	client.store = mockStore
	r.client = client

	// Add a secret to the mock store
	mockStore.secrets["test/secret"] = newMockSecret("test-password")

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

	req := resource.DeleteRequest{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    stateValue,
		},
	}
	resp := &resource.DeleteResponse{}

	r.Delete(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %v", resp.Diagnostics)
	}

	// Verify secret was removed from mock store
	if _, exists := mockStore.secrets["test/secret"]; exists {
		t.Error("expected secret to be removed from mock store")
	}
}

func TestSecretResource_Delete_KeepSecret(t *testing.T) {
	r := &SecretResource{}
	mockStore := newMockStore()
	client := NewGopassClient("")
	client.store = mockStore
	r.client = client

	// Add a secret to the mock store
	mockStore.secrets["test/secret"] = newMockSecret("test-password")

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
		"delete_on_remove": tftypes.NewValue(tftypes.Bool, false), // Keep secret
		"revision_count":   tftypes.NewValue(tftypes.Number, 1),
	})

	req := resource.DeleteRequest{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    stateValue,
		},
	}
	resp := &resource.DeleteResponse{}

	r.Delete(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %v", resp.Diagnostics)
	}

	// Verify secret still exists in mock store
	if _, exists := mockStore.secrets["test/secret"]; !exists {
		t.Error("expected secret to still exist in mock store")
	}
}

func TestSecretResource_Delete_AlreadyDeleted(t *testing.T) {
	r := &SecretResource{}
	mockStore := newMockStore()
	client := NewGopassClient("")
	client.store = mockStore
	r.client = client

	// Do NOT add the secret - simulate it was already deleted externally

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

	req := resource.DeleteRequest{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    stateValue,
		},
	}
	resp := &resource.DeleteResponse{}

	r.Delete(ctx, req, resp)

	// Should succeed without error - gracefully handles already-deleted secret
	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error for already-deleted secret: %v", resp.Diagnostics)
	}
}

func TestSecretResource_Delete_RemoveError(t *testing.T) {
	r := &SecretResource{}
	mockStore := newMockStore()
	mockStore.shouldFail = true
	mockStore.failMsg = "permission denied"
	client := NewGopassClient("")
	client.store = mockStore
	r.client = client

	// Add a secret to the mock store
	mockStore.secrets["test/secret"] = newMockSecret("test-password")

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

	req := resource.DeleteRequest{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    stateValue,
		},
	}
	resp := &resource.DeleteResponse{}

	r.Delete(ctx, req, resp)

	// Should have error for non-"not found" errors
	if !resp.Diagnostics.HasError() {
		t.Error("expected error for remove failure")
	}
}
