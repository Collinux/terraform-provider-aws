---
subcategory: "API Gateway"
layout: "aws"
page_title: "AWS: aws_api_gateway_deployment"
description: |-
  Manages an API Gateway REST Deployment.
---

# Resource: aws_api_gateway_deployment

Manages an API Gateway REST Deployment. A deployment is a snapshot of the REST API configuration. The deployment can then be published to callable endpoints via the [`aws_api_gateway_stage` resource](api_gateway_stage.html) and optionally managed further with the [`aws_api_gateway_base_path_mapping` resource](api_gateway_base_path_mapping.html), [`aws_api_gateway_domain_name` resource](api_gateway_domain_name.html), and [`aws_api_method_settings` resource](api_gateway_method_settings.html). For more information, see the [API Gateway Developer Guide](https://docs.aws.amazon.com/apigateway/latest/developerguide/how-to-deploy-api.html).

To properly capture all REST API configuration in a deployment, this resource must have dependencies on all prior Terraform resources that manage resources/paths, methods, integrations, etc.

* For REST APIs that are configured via OpenAPI specification ([`aws_api_gateway_rest_api` resource](api_gateway_rest_api.html) `body` argument), no special dependency setup is needed beyond referencing the  `id` attribute of that resource unless additional Terraform resources have further customized the REST API.
* When the REST API configuration involves other Terraform resources ([`aws_api_gateway_integration` resource](api_gateway_integration.html), etc.), the dependency setup can be done with implicit resource references in the `triggers` argument or explicit resource references using the [resource `depends_on` meta-argument](https://www.terraform.io/docs/configuration/meta-arguments/depends_on.html). The `triggers` argument should be preferred over `depends_on`, since `depends_on` can only capture dependency ordering and will not cause the resource to recreate (redeploy the REST API) with upstream configuration changes.

~> **NOTE:** Enable the [resource `lifecycle` configuration block `create_before_destroy` argument](https://www.terraform.io/language/meta-arguments/lifecycle#create_before_destroy) in this resource configuration to properly order redeployments in Terraform. Without enabling `create_before_destroy`, API Gateway can return errors such as `BadRequestException: Active stages pointing to this deployment must be moved or deleted` on recreation.

## Example Usage

### OpenAPI Specification

An end-to-end example of a REST API configured with OpenAPI can be found in the [`/examples/api-gateway-rest-api-openapi` directory within the GitHub repository](https://github.com/hashicorp/terraform-provider-aws/tree/main/examples/api-gateway-rest-api-openapi).

```terraform
resource "aws_api_gateway_rest_api" "example" {
  body = jsonencode({
    openapi = "3.0.1"
    info = {
      title   = "example"
      version = "1.0"
    }
    paths = {
      "/path1" = {
        get = {
          x-amazon-apigateway-integration = {
            httpMethod           = "GET"
            payloadFormatVersion = "1.0"
            type                 = "HTTP_PROXY"
            uri                  = "https://ip-ranges.amazonaws.com/ip-ranges.json"
          }
        }
      }
    }
  })

  name = "example"
}

resource "aws_api_gateway_deployment" "example" {
  rest_api_id = aws_api_gateway_rest_api.example.id

  triggers = {
    redeployment = sha1(jsonencode(aws_api_gateway_rest_api.example.body))
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_api_gateway_stage" "example" {
  deployment_id = aws_api_gateway_deployment.example.id
  rest_api_id   = aws_api_gateway_rest_api.example.id
  stage_name    = "example"
}
```

### Terraform Resources

```terraform
resource "aws_api_gateway_rest_api" "example" {
  name = "example"
}

resource "aws_api_gateway_resource" "example" {
  parent_id   = aws_api_gateway_rest_api.example.root_resource_id
  path_part   = "example"
  rest_api_id = aws_api_gateway_rest_api.example.id
}

resource "aws_api_gateway_method" "example" {
  authorization = "NONE"
  http_method   = "GET"
  resource_id   = aws_api_gateway_resource.example.id
  rest_api_id   = aws_api_gateway_rest_api.example.id
}

resource "aws_api_gateway_integration" "example" {
  http_method = aws_api_gateway_method.example.http_method
  resource_id = aws_api_gateway_resource.example.id
  rest_api_id = aws_api_gateway_rest_api.example.id
  type        = "MOCK"
}

resource "aws_api_gateway_deployment" "example" {
  rest_api_id = aws_api_gateway_rest_api.example.id

  triggers = {
    # NOTE: The configuration below will satisfy ordering considerations,
    #       but not pick up all future REST API changes. More advanced patterns
    #       are possible, such as using the filesha1() function against the
    #       Terraform configuration file(s) or removing the .id references to
    #       calculate a hash against whole resources. Be aware that using whole
    #       resources will show a difference after the initial implementation.
    #       It will stabilize to only change when resources change afterwards.
    redeployment = sha1(jsonencode([
      aws_api_gateway_resource.example.id,
      aws_api_gateway_method.example.id,
      aws_api_gateway_integration.example.id,
    ]))
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_api_gateway_stage" "example" {
  deployment_id = aws_api_gateway_deployment.example.id
  rest_api_id   = aws_api_gateway_rest_api.example.id
  stage_name    = "example"
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `description` - (Optional) Description of the deployment.
* `rest_api_id` - (Required) REST API identifier.
* `triggers` - (Optional) Map of arbitrary keys and values that, when changed, will trigger a redeployment. To force a redeployment without changing these keys/values, use the [`-replace` option](https://developer.hashicorp.com/terraform/cli/commands/plan#replace-address) with `terraform plan` or `terraform apply`.
* `variables` - (Optional) Map to set on the related stage.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the deployment
* `created_date` - Creation date of the deployment

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_api_gateway_deployment` using `REST-API-ID/DEPLOYMENT-ID`. For example:

```terraform
import {
  to = aws_api_gateway_deployment.example
  id = "aabbccddee/1122334"
}
```

Using `terraform import`, import `aws_api_gateway_deployment` using `REST-API-ID/DEPLOYMENT-ID`. For example:

```console
% terraform import aws_api_gateway_deployment.example aabbccddee/1122334
```

The `variables` arguments cannot be imported. Use the [`aws_api_gateway_stage` resource](api_gateway_stage.html) to import and manage stages.

The `triggers` argument cannot be imported.
