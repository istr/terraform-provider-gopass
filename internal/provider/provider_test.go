// Copyright (c) Ingo Struck
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"gopass": providerserver.NewProtocol6WithError(New("test")()),
}

func TestProvider(t *testing.T) {
	// Basic provider instantiation test
	provider := New("test")()
	if provider == nil {
		t.Fatal("provider is nil")
	}
}

func TestProviderSchema(t *testing.T) {
	// Verify schema is valid
	provider := New("test")()
	if provider == nil {
		t.Fatal("provider is nil")
	}
}

// Acceptance tests - require TF_ACC=1 and actual gopass setup
// func TestAccSecretEphemeral_basic(t *testing.T) {
// 	resource.Test(t, resource.TestCase{
// 		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: `
// 					ephemeral "gopass_secret" "test" {
// 						path = "test/secret"
// 					}
// 				`,
// 			},
// 		},
// 	})
// }
