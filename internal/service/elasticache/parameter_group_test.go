// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfelasticache "github.com/hashicorp/terraform-provider-aws/internal/service/elasticache"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccElastiCacheParameterGroup_Redis_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.CacheParameterGroup
	resourceName := "aws_elasticache_parameter_group.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig_Redis_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, names.AttrFamily, "redis2.8"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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

func TestAccElastiCacheParameterGroup_Valkey_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.CacheParameterGroup
	resourceName := "aws_elasticache_parameter_group.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig_Valkey_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, names.AttrFamily, "valkey7"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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

func TestAccElastiCacheParameterGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.CacheParameterGroup
	resourceName := "aws_elasticache_parameter_group.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig_Redis_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, t, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfelasticache.ResourceParameterGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccElastiCacheParameterGroup_addParameter(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.CacheParameterGroup
	resourceName := "aws_elasticache_parameter_group.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig_1(rName, "redis2.8", "appendonly", "yes"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "appendonly",
						names.AttrValue: "yes",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccParameterGroupConfig_2(rName, "redis2.8", "appendonly", "yes", "appendfsync", "always"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "appendonly",
						names.AttrValue: "yes",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "appendfsync",
						names.AttrValue: "always",
					}),
				),
			},
		},
	})
}

// Regression for https://github.com/hashicorp/terraform-provider-aws/issues/116
func TestAccElastiCacheParameterGroup_removeAllParameters(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.CacheParameterGroup
	resourceName := "aws_elasticache_parameter_group.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig_2(rName, "redis2.8", "appendonly", "yes", "appendfsync", "always"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "appendonly",
						names.AttrValue: "yes",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "appendfsync",
						names.AttrValue: "always",
					}),
				),
			},
			{
				Config: testAccParameterGroupConfig_Redis_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "0"),
				),
			},
		},
	})
}

// The API returns errors when attempting to reset the reserved-memory parameter.
// This covers our custom logic handling for this situation.
func TestAccElastiCacheParameterGroup_RemoveReservedMemoryParameter_allParameters(t *testing.T) {
	ctx := acctest.Context(t)
	var cacheParameterGroup1 awstypes.CacheParameterGroup
	resourceName := "aws_elasticache_parameter_group.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig_1(rName, "redis3.2", "reserved-memory", "0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, t, resourceName, &cacheParameterGroup1),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "reserved-memory",
						names.AttrValue: "0",
					}),
				),
			},
			{
				Config: testAccParameterGroupConfig_Redis_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, t, resourceName, &cacheParameterGroup1),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "0"),
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

// The API returns errors when attempting to reset the reserved-memory parameter.
// This covers our custom logic handling for this situation.
func TestAccElastiCacheParameterGroup_RemoveReservedMemoryParameter_remainingParameters(t *testing.T) {
	ctx := acctest.Context(t)
	var cacheParameterGroup1 awstypes.CacheParameterGroup
	resourceName := "aws_elasticache_parameter_group.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig_2(rName, "redis3.2", "reserved-memory", "0", "tcp-keepalive", "360"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, t, resourceName, &cacheParameterGroup1),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "reserved-memory",
						names.AttrValue: "0",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "tcp-keepalive",
						names.AttrValue: "360",
					}),
				),
			},
			{
				Config: testAccParameterGroupConfig_1(rName, "redis3.2", "tcp-keepalive", "360"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, t, resourceName, &cacheParameterGroup1),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "tcp-keepalive",
						names.AttrValue: "360",
					}),
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

// The API returns errors when attempting to reset the reserved-memory parameter.
// This covers our custom logic handling for this situation.
func TestAccElastiCacheParameterGroup_switchReservedMemoryParameter(t *testing.T) {
	ctx := acctest.Context(t)
	var cacheParameterGroup1 awstypes.CacheParameterGroup
	resourceName := "aws_elasticache_parameter_group.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig_1(rName, "redis3.2", "reserved-memory", "0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, t, resourceName, &cacheParameterGroup1),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "reserved-memory",
						names.AttrValue: "0",
					}),
				),
			},
			{
				Config: testAccParameterGroupConfig_1(rName, "redis3.2", "reserved-memory-percent", "25"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, t, resourceName, &cacheParameterGroup1),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "reserved-memory-percent",
						names.AttrValue: "25",
					}),
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

// The API returns errors when attempting to reset the reserved-memory parameter.
// This covers our custom logic handling for this situation.
func TestAccElastiCacheParameterGroup_updateReservedMemoryParameter(t *testing.T) {
	ctx := acctest.Context(t)
	var cacheParameterGroup1 awstypes.CacheParameterGroup
	resourceName := "aws_elasticache_parameter_group.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig_1(rName, "redis2.8", "reserved-memory", "0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, t, resourceName, &cacheParameterGroup1),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "reserved-memory",
						names.AttrValue: "0",
					}),
				),
			},
			{
				Config: testAccParameterGroupConfig_1(rName, "redis2.8", "reserved-memory", "1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, t, resourceName, &cacheParameterGroup1),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "reserved-memory",
						names.AttrValue: "1",
					}),
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

func TestAccElastiCacheParameterGroup_uppercaseName(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.CacheParameterGroup
	resourceName := "aws_elasticache_parameter_group.test"
	rInt := acctest.RandInt(t)
	rName := fmt.Sprintf("TF-ELASTIPG-%d", rInt)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig_1(rName, "redis2.8", "appendonly", "yes"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, fmt.Sprintf("tf-elastipg-%d", rInt)),
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

func TestAccElastiCacheParameterGroup_description(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.CacheParameterGroup
	resourceName := "aws_elasticache_parameter_group.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
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

func TestAccElastiCacheParameterGroup_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var cacheParameterGroup1 awstypes.CacheParameterGroup
	resourceName := "aws_elasticache_parameter_group.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig_tags1(rName, "redis2.8", acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, t, resourceName, &cacheParameterGroup1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccParameterGroupConfig_tags2(rName, "redis2.8", acctest.CtKey1, "updatedvalue1", acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, t, resourceName, &cacheParameterGroup1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, "updatedvalue1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccParameterGroupConfig_Redis_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, t, resourceName, &cacheParameterGroup1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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

func testAccCheckParameterGroupDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ElastiCacheClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_elasticache_parameter_group" {
				continue
			}

			_, err := tfelasticache.FindCacheParameterGroupByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ElastiCache Parameter Group %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckParameterGroupExists(ctx context.Context, t *testing.T, n string, v *awstypes.CacheParameterGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ElastiCache Parameter Group ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).ElastiCacheClient(ctx)

		output, err := tfelasticache.FindCacheParameterGroupByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccParameterGroupConfig_Redis_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_parameter_group" "test" {
  family = "redis2.8"
  name   = %[1]q
}
`, rName)
}

func testAccParameterGroupConfig_Valkey_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_parameter_group" "test" {
  family = "valkey7"
  name   = %[1]q
}
`, rName)
}

func testAccParameterGroupConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_parameter_group" "test" {
  description = %[1]q
  family      = "redis2.8"
  name        = %[2]q
}
`, description, rName)
}

func testAccParameterGroupConfig_1(rName, family, parameterName1, parameterValue1 string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_parameter_group" "test" {
  family = %[1]q
  name   = %[2]q

  parameter {
    name  = %[3]q
    value = %[4]q
  }
}
`, family, rName, parameterName1, parameterValue1)
}

func testAccParameterGroupConfig_2(rName, family, parameterName1, parameterValue1, parameterName2, parameterValue2 string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_parameter_group" "test" {
  family = %[1]q
  name   = %[2]q

  parameter {
    name  = %[3]q
    value = %[4]q
  }

  parameter {
    name  = %[5]q
    value = %[6]q
  }
}
`, family, rName, parameterName1, parameterValue1, parameterName2, parameterValue2)
}

func testAccParameterGroupConfig_tags1(rName, family, tagName1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_parameter_group" "test" {
  family = %[1]q
  name   = %[2]q

  tags = {
    %[3]s = %[4]q
  }
}
`, family, rName, tagName1, tagValue1)
}

func testAccParameterGroupConfig_tags2(rName, family, tagName1, tagValue1, tagName2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_parameter_group" "test" {
  family = %[1]q
  name   = %[2]q

  tags = {
    %[3]s = %[4]q
    %[5]s = %[6]q
  }
}
`, family, rName, tagName1, tagValue1, tagName2, tagValue2)
}

func TestParameterChanges(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name                string
		Old                 *schema.Set
		New                 *schema.Set
		ExpectedRemove      []*awstypes.ParameterNameValue
		ExpectedAddOrUpdate []*awstypes.ParameterNameValue
	}{
		{
			Name:                "Empty",
			Old:                 new(schema.Set),
			New:                 new(schema.Set),
			ExpectedRemove:      []*awstypes.ParameterNameValue{},
			ExpectedAddOrUpdate: []*awstypes.ParameterNameValue{},
		},
		{
			Name: "Remove all",
			Old: schema.NewSet(tfelasticache.ParameterHash, []any{
				map[string]any{
					names.AttrName:  "reserved-memory",
					names.AttrValue: "0",
				},
			}),
			New: new(schema.Set),
			ExpectedRemove: []*awstypes.ParameterNameValue{
				{
					ParameterName:  aws.String("reserved-memory"),
					ParameterValue: aws.String("0"),
				},
			},
			ExpectedAddOrUpdate: []*awstypes.ParameterNameValue{},
		},
		{
			Name: "No change",
			Old: schema.NewSet(tfelasticache.ParameterHash, []any{
				map[string]any{
					names.AttrName:  "reserved-memory",
					names.AttrValue: "0",
				},
			}),
			New: schema.NewSet(tfelasticache.ParameterHash, []any{
				map[string]any{
					names.AttrName:  "reserved-memory",
					names.AttrValue: "0",
				},
			}),
			ExpectedRemove:      []*awstypes.ParameterNameValue{},
			ExpectedAddOrUpdate: []*awstypes.ParameterNameValue{},
		},
		{
			Name: "Remove partial",
			Old: schema.NewSet(tfelasticache.ParameterHash, []any{
				map[string]any{
					names.AttrName:  "reserved-memory",
					names.AttrValue: "0",
				},
				map[string]any{
					names.AttrName:  "appendonly",
					names.AttrValue: "yes",
				},
			}),
			New: schema.NewSet(tfelasticache.ParameterHash, []any{
				map[string]any{
					names.AttrName:  "appendonly",
					names.AttrValue: "yes",
				},
			}),
			ExpectedRemove: []*awstypes.ParameterNameValue{
				{
					ParameterName:  aws.String("reserved-memory"),
					ParameterValue: aws.String("0"),
				},
			},
			ExpectedAddOrUpdate: []*awstypes.ParameterNameValue{},
		},
		{
			Name: "Add to existing",
			Old: schema.NewSet(tfelasticache.ParameterHash, []any{
				map[string]any{
					names.AttrName:  "appendonly",
					names.AttrValue: "yes",
				},
			}),
			New: schema.NewSet(tfelasticache.ParameterHash, []any{
				map[string]any{
					names.AttrName:  "appendonly",
					names.AttrValue: "yes",
				},
				map[string]any{
					names.AttrName:  "appendfsync",
					names.AttrValue: "always",
				},
			}),
			ExpectedRemove: []*awstypes.ParameterNameValue{},
			ExpectedAddOrUpdate: []*awstypes.ParameterNameValue{
				{
					ParameterName:  aws.String("appendfsync"),
					ParameterValue: aws.String("always"),
				},
			},
		},
	}

	for _, tc := range cases {
		remove, addOrUpdate := tfelasticache.ParameterChanges(tc.Old, tc.New)
		if !reflect.DeepEqual(remove, tc.ExpectedRemove) {
			t.Errorf("Case %q: Remove did not match\n%#v\n\nGot:\n%#v", tc.Name, tc.ExpectedRemove, remove)
		}
		if !reflect.DeepEqual(addOrUpdate, tc.ExpectedAddOrUpdate) {
			t.Errorf("Case %q: AddOrUpdate did not match\n%#v\n\nGot:\n%#v", tc.Name, tc.ExpectedAddOrUpdate, addOrUpdate)
		}
	}
}
