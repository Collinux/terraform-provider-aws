// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apprunner_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapprunner "github.com/hashicorp/terraform-provider-aws/internal/service/apprunner"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppRunnerAutoScalingConfigurationVersion_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_auto_scaling_configuration_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAutoScalingConfigurationVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAutoScalingConfigurationVersionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoScalingConfigurationVersionExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "apprunner", regexache.MustCompile(fmt.Sprintf(`autoscalingconfiguration/%s/1/.+`, rName))),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_configuration_name", rName),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_configuration_revision", "1"),
					resource.TestCheckResourceAttr(resourceName, "has_associated_service", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "is_default", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "latest", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "max_concurrency", "100"),
					resource.TestCheckResourceAttr(resourceName, "max_size", "25"),
					resource.TestCheckResourceAttr(resourceName, "min_size", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "active"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppRunnerAutoScalingConfigurationVersion_complex(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_auto_scaling_configuration_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAutoScalingConfigurationVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAutoScalingConfigurationVersionConfig_nonDefaults(rName, 50, 10, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoScalingConfigurationVersionExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "apprunner", regexache.MustCompile(fmt.Sprintf(`autoscalingconfiguration/%s/1/.+`, rName))),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_configuration_name", rName),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_configuration_revision", "1"),
					resource.TestCheckResourceAttr(resourceName, "latest", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "max_concurrency", "50"),
					resource.TestCheckResourceAttr(resourceName, "max_size", "10"),
					resource.TestCheckResourceAttr(resourceName, "min_size", "2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "active"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Test resource recreation such that the revision number is still 1
				Config: testAccAutoScalingConfigurationVersionConfig_nonDefaults(rName, 150, 20, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoScalingConfigurationVersionExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "apprunner", regexache.MustCompile(fmt.Sprintf(`autoscalingconfiguration/%s/1/.+`, rName))),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_configuration_name", rName),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_configuration_revision", "1"),
					resource.TestCheckResourceAttr(resourceName, "latest", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "max_concurrency", "150"),
					resource.TestCheckResourceAttr(resourceName, "max_size", "20"),
					resource.TestCheckResourceAttr(resourceName, "min_size", "5"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "active"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Test resource recreation such that the revision number is still 1
				Config: testAccAutoScalingConfigurationVersionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoScalingConfigurationVersionExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "apprunner", regexache.MustCompile(fmt.Sprintf(`autoscalingconfiguration/%s/1/.+`, rName))),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_configuration_name", rName),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_configuration_revision", "1"),
					resource.TestCheckResourceAttr(resourceName, "latest", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "max_concurrency", "100"),
					resource.TestCheckResourceAttr(resourceName, "max_size", "25"),
					resource.TestCheckResourceAttr(resourceName, "min_size", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "active"),
				),
			},
		},
	})
}

func TestAccAppRunnerAutoScalingConfigurationVersion_multipleVersions(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_auto_scaling_configuration_version.test"
	otherResourceName := "aws_apprunner_auto_scaling_configuration_version.other"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAutoScalingConfigurationVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAutoScalingConfigurationVersionConfig_multipleVersions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoScalingConfigurationVersionExists(ctx, resourceName),
					testAccCheckAutoScalingConfigurationVersionExists(ctx, otherResourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "apprunner", regexache.MustCompile(fmt.Sprintf(`autoscalingconfiguration/%s/1/.+`, rName))),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_configuration_name", rName),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_configuration_revision", "1"),
					resource.TestCheckResourceAttr(resourceName, "latest", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "max_concurrency", "100"),
					resource.TestCheckResourceAttr(resourceName, "max_size", "25"),
					resource.TestCheckResourceAttr(resourceName, "min_size", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "active"),
					acctest.MatchResourceAttrRegionalARN(ctx, otherResourceName, names.AttrARN, "apprunner", regexache.MustCompile(fmt.Sprintf(`autoscalingconfiguration/%s/2/.+`, rName))),
					resource.TestCheckResourceAttr(otherResourceName, "auto_scaling_configuration_name", rName),
					resource.TestCheckResourceAttr(otherResourceName, "auto_scaling_configuration_revision", "2"),
					resource.TestCheckResourceAttr(otherResourceName, "latest", acctest.CtTrue),
					resource.TestCheckResourceAttr(otherResourceName, "max_concurrency", "100"),
					resource.TestCheckResourceAttr(otherResourceName, "max_size", "25"),
					resource.TestCheckResourceAttr(otherResourceName, "min_size", "1"),
					resource.TestCheckResourceAttr(otherResourceName, names.AttrStatus, "active"),
				),
			},
			{
				// Test update of "latest" computed attribute after apply
				Config: testAccAutoScalingConfigurationVersionConfig_multipleVersions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoScalingConfigurationVersionExists(ctx, resourceName),
					testAccCheckAutoScalingConfigurationVersionExists(ctx, otherResourceName),
					resource.TestCheckResourceAttr(resourceName, "latest", acctest.CtFalse),
					resource.TestCheckResourceAttr(otherResourceName, "latest", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      otherResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppRunnerAutoScalingConfigurationVersion_updateMultipleVersions(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_auto_scaling_configuration_version.test"
	otherResourceName := "aws_apprunner_auto_scaling_configuration_version.other"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAutoScalingConfigurationVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAutoScalingConfigurationVersionConfig_multipleVersions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoScalingConfigurationVersionExists(ctx, resourceName),
					testAccCheckAutoScalingConfigurationVersionExists(ctx, otherResourceName),
				),
			},
			{
				Config: testAccAutoScalingConfigurationVersionConfig_updateMultipleVersions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoScalingConfigurationVersionExists(ctx, resourceName),
					testAccCheckAutoScalingConfigurationVersionExists(ctx, otherResourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "apprunner", regexache.MustCompile(fmt.Sprintf(`autoscalingconfiguration/%s/1/.+`, rName))),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_configuration_name", rName),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_configuration_revision", "1"),
					resource.TestCheckResourceAttr(resourceName, "latest", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "max_concurrency", "100"),
					resource.TestCheckResourceAttr(resourceName, "max_size", "25"),
					resource.TestCheckResourceAttr(resourceName, "min_size", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "active"),
					acctest.MatchResourceAttrRegionalARN(ctx, otherResourceName, names.AttrARN, "apprunner", regexache.MustCompile(fmt.Sprintf(`autoscalingconfiguration/%s/2/.+`, rName))),
					resource.TestCheckResourceAttr(otherResourceName, "auto_scaling_configuration_name", rName),
					resource.TestCheckResourceAttr(otherResourceName, "auto_scaling_configuration_revision", "2"),
					resource.TestCheckResourceAttr(otherResourceName, "latest", acctest.CtTrue),
					resource.TestCheckResourceAttr(otherResourceName, "max_concurrency", "125"),
					resource.TestCheckResourceAttr(otherResourceName, "max_size", "20"),
					resource.TestCheckResourceAttr(otherResourceName, "min_size", "1"),
					resource.TestCheckResourceAttr(otherResourceName, names.AttrStatus, "active"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      otherResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppRunnerAutoScalingConfigurationVersion_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_auto_scaling_configuration_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAutoScalingConfigurationVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAutoScalingConfigurationVersionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoScalingConfigurationVersionExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapprunner.ResourceAutoScalingConfigurationVersion(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAppRunnerAutoScalingConfigurationVersion_Identity_ExistingResource(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_auto_scaling_configuration_version.test"

	resource.ParallelTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_12_0),
		},
		PreCheck:     func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.AppRunnerServiceID),
		CheckDestroy: testAccCheckAutoScalingConfigurationVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.100.0",
					},
				},
				Config: testAccAutoScalingConfigurationVersionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoScalingConfigurationVersionExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectNoIdentity(resourceName),
				},
			},
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "6.0.0",
					},
				},
				Config: testAccAutoScalingConfigurationVersionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoScalingConfigurationVersionExists(ctx, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrARN: knownvalue.Null(),
					}),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccAutoScalingConfigurationVersionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoScalingConfigurationVersionExists(ctx, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrARN: tfknownvalue.RegionalARNRegexp("apprunner", regexache.MustCompile(`autoscalingconfiguration/.+`)),
					}),
				},
			},
		},
	})
}

func testAccCheckAutoScalingConfigurationVersionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_apprunner_auto_scaling_configuration_version" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).AppRunnerClient(ctx)

			_, err := tfapprunner.FindAutoScalingConfigurationByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("App Runner AutoScaling Configuration Version %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAutoScalingConfigurationVersionExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No App Runner Service ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppRunnerClient(ctx)

		_, err := tfapprunner.FindAutoScalingConfigurationByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccAutoScalingConfigurationVersionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_auto_scaling_configuration_version" "test" {
  auto_scaling_configuration_name = %[1]q
}
`, rName)
}

func testAccAutoScalingConfigurationVersionConfig_nonDefaults(rName string, maxConcurrency, maxSize, minSize int) string {
	return fmt.Sprintf(`
resource "aws_apprunner_auto_scaling_configuration_version" "test" {
  auto_scaling_configuration_name = %[1]q

  max_concurrency = %[2]d
  max_size        = %[3]d
  min_size        = %[4]d
}
`, rName, maxConcurrency, maxSize, minSize)
}

func testAccAutoScalingConfigurationVersionConfig_multipleVersions(rName string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_auto_scaling_configuration_version" "test" {
  auto_scaling_configuration_name = %[1]q
}

resource "aws_apprunner_auto_scaling_configuration_version" "other" {
  auto_scaling_configuration_name = aws_apprunner_auto_scaling_configuration_version.test.auto_scaling_configuration_name
}
`, rName)
}

func testAccAutoScalingConfigurationVersionConfig_updateMultipleVersions(rName string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_auto_scaling_configuration_version" "test" {
  auto_scaling_configuration_name = %[1]q
}

resource "aws_apprunner_auto_scaling_configuration_version" "other" {
  auto_scaling_configuration_name = aws_apprunner_auto_scaling_configuration_version.test.auto_scaling_configuration_name

  max_concurrency = 125
  max_size        = 20
}
`, rName)
}
