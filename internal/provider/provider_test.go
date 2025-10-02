// Copyright (c) 0ch1r
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"auditlogfilter": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the appropriate environment variables being set are common to see in a pre-check
	// function.
}

// TestAccAuditLogFilterResource_basic tests basic filter creation
func TestAccAuditLogFilterResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccAuditLogFilterResourceConfig("test_filter", `{"filter":{"class":{"name":"connection"}}}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("auditlogfilter_filter.test", "name", "test_filter"),
					resource.TestCheckResourceAttr("auditlogfilter_filter.test", "id", "test_filter"),
					resource.TestCheckResourceAttrSet("auditlogfilter_filter.test", "filter_id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "auditlogfilter_filter.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccAuditLogFilterResourceConfig("test_filter", `{"filter":{"class":{"name":"general"}}}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("auditlogfilter_filter.test", "name", "test_filter"),
					resource.TestCheckResourceAttr("auditlogfilter_filter.test", "definition", `{"filter":{"class":{"name":"general"}}}`),
				),
			},
		},
	})
}

func testAccAuditLogFilterResourceConfig(name, definition string) string {
	return `
provider "auditlogfilter" {
  endpoint = "localhost:3306"
  username = "root"
  password = ""
}

resource "auditlogfilter_filter" "test" {
  name       = "` + name + `"
  definition = "` + definition + `"
}
`
}

// TestAccAuditLogUserAssignmentResource_basic tests basic user assignment
func TestAccAuditLogUserAssignmentResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create filter first
			{
				Config: testAccAuditLogUserAssignmentResourceConfig("test_user", "%", "test_assignment_filter"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("auditlogfilter_user_assignment.test", "username", "test_user"),
					resource.TestCheckResourceAttr("auditlogfilter_user_assignment.test", "userhost", "%"),
					resource.TestCheckResourceAttr("auditlogfilter_user_assignment.test", "filter_name", "test_assignment_filter"),
					resource.TestCheckResourceAttr("auditlogfilter_user_assignment.test", "id", "test_user@%"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "auditlogfilter_user_assignment.test",
				ImportState:       true,
				ImportStateId:     "test_user@%",
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testAccCheckAuditLogUserAssignmentDestroy,
	})
}

func testAccAuditLogUserAssignmentResourceConfig(username, userhost, filterName string) string {
	return `
provider "auditlogfilter" {
  endpoint = "localhost:3306"
  username = "root"
  password = ""
}

resource "auditlogfilter_filter" "test_assignment" {
  name       = "` + filterName + `"
  definition = "{\"filter\":{\"class\":{\"name\":\"connection\"}}}"
}

resource "auditlogfilter_user_assignment" "test" {
  username    = "` + username + `"
  userhost    = "` + userhost + `"
  filter_name = auditlogfilter_filter.test_assignment.name
}
`
}

func testAccCheckAuditLogUserAssignmentDestroy(s *terraform.State) error {
	// Add logic to verify the user assignment no longer exists
	return nil
}
