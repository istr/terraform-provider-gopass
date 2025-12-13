// Copyright (c) Ingo Struck
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestSecretResource_Configure(t *testing.T) {
	r := &SecretResource{}
	client := NewGopassClient("")

	req := resource.ConfigureRequest{
		ProviderData: client,
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %v", resp.Diagnostics)
	}

	if r.client != client {
		t.Error("expected client to be set")
	}
}

func TestSecretResource_Configure_InvalidType(t *testing.T) {
	r := &SecretResource{}

	req := resource.ConfigureRequest{
		ProviderData: "invalid",
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for invalid provider data type")
	}
}

func TestSecretResource_Configure_NilData(t *testing.T) {
	r := &SecretResource{}

	req := resource.ConfigureRequest{
		ProviderData: nil,
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %v", resp.Diagnostics)
	}
}
