// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_elasticache_serverless_cache", name="Serverless Cache")
// @Tags(identifierAttribute="arn")
func newServerlessCacheResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &serverlessCacheResource{}

	r.SetDefaultCreateTimeout(40 * time.Minute)
	r.SetDefaultUpdateTimeout(80 * time.Minute)
	r.SetDefaultDeleteTimeout(40 * time.Minute)

	return r, nil
}

type serverlessCacheResource struct {
	framework.ResourceWithModel[serverlessCacheResourceModel]
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *serverlessCacheResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreateTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"daily_snapshot_time": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrEndpoint: schema.ListAttribute{
				CustomType:  fwtypes.NewListNestedObjectTypeOf[endpointModel](ctx),
				ElementType: fwtypes.NewObjectTypeOf[endpointModel](ctx),
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrEngine: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(
						func(ctx context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
							// In-place updates are only supported for redis -> valkey
							if req.StateValue.Equal(types.StringValue(engineRedis)) && req.PlanValue.Equal(types.StringValue(engineValkey)) {
								return
							}

							// Any other change will force a replacement
							resp.RequiresReplace = true
						},
						"Engine modifications other than redis to valkey require a replacement",
						"Engine modifications other than redis to valkey require a replacement",
					),
				},
			},
			"full_engine_version": schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrKMSKeyID: schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"major_engine_version": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"reader_endpoint": schema.ListAttribute{
				CustomType:  fwtypes.NewListNestedObjectTypeOf[endpointModel](ctx),
				ElementType: fwtypes.NewObjectTypeOf[endpointModel](ctx),
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrSecurityGroupIDs: schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"snapshot_arns_to_restore": schema.ListAttribute{
				CustomType:  fwtypes.ListOfARNType,
				ElementType: fwtypes.ARNType,
				Optional:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"snapshot_retention_limit": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Validators: []validator.Int64{
					int64validator.AtMost(35),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrSubnetIDs: schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
					setplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"user_group_id": schema.StringAttribute{
				Optional: true,
			},
		},
		Blocks: map[string]schema.Block{
			"cache_usage_limits": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[cacheUsageLimitsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"data_storage": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[dataStorageModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"maximum": schema.Int64Attribute{
										Optional: true,
									},
									"minimum": schema.Int64Attribute{
										Optional: true,
									},
									names.AttrUnit: schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.DataStorageUnit](),
										Required:   true,
									},
								},
							},
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
						},
						"ecpu_per_second": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[ecpuPerSecondModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"maximum": schema.Int64Attribute{
										Optional: true,
										Validators: []validator.Int64{
											int64validator.Between(1000, 15000000),
										},
									},
									"minimum": schema.Int64Attribute{
										Optional: true,
										Validators: []validator.Int64{
											int64validator.Between(1000, 15000000),
										},
									},
								},
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *serverlessCacheResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data serverlessCacheResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ElastiCacheClient(ctx)

	input := &elasticache.CreateServerlessCacheInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	_, err := conn.CreateServerlessCache(ctx, input)
	if err != nil {
		response.Diagnostics.AddError("creating ElastiCache Serverless Cache", err.Error())

		return
	}

	// Set values for unknowns.
	data.setID()

	output, err := waitServerlessCacheAvailable(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for ElastiCache Serverless Cache (%s) create", data.ID.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *serverlessCacheResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data serverlessCacheResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())
		return
	}

	conn := r.Meta().ElastiCacheClient(ctx)

	output, err := findServerlessCacheByID(ctx, conn, data.ID.ValueString())

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading ElastiCache Serverless Cache (%s)", data.ID.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *serverlessCacheResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new serverlessCacheResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ElastiCacheClient(ctx)

	diff, d := fwflex.Diff(ctx, new, old)
	response.Diagnostics.Append(d...)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		input := elasticache.ModifyServerlessCacheInput{
			ServerlessCacheName: new.ServerlessCacheName.ValueStringPointer(),
		}
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input, diff.IgnoredFieldNamesOpts()...)...)
		if response.Diagnostics.HasError() {
			return
		}

		// If no engine changes are made, unset related fields to prevent the following error:
		// This API supports only cross-engine upgrades to Valkey engine currently.
		if new.Engine.Equal(old.Engine) && new.MajorEngineVersion.Equal(old.MajorEngineVersion) {
			input.Engine = nil
			input.MajorEngineVersion = nil
		}

		// If engine is changed but major_engine_version is omitted in configuration, explicitly
		// include it in the request to prevent the following error:
		// InvalidParameterCombination: No modifications were requested
		if !new.Engine.Equal(old.Engine) && input.MajorEngineVersion == nil {
			input.MajorEngineVersion = old.MajorEngineVersion.ValueStringPointer()
		}

		if _, err := conn.ModifyServerlessCache(ctx, &input); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating ElastiCache Serverless Cache (%s)", new.ID.ValueString()), err.Error())
			return
		}

		if _, err := waitServerlessCacheAvailable(ctx, conn, old.ServerlessCacheName.ValueString(), r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for ElastiCache Serverless Cache (%s) update", new.ID.ValueString()), err.Error())
			return
		}
	}

	// AWS returns null values for certain values that are available on redis/valkey only.
	// always set these values to the state value to avoid unnecessary diff failures on computed values.
	output, err := findServerlessCacheByID(ctx, conn, old.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading ElastiCache Serverless Cache (%s)", old.ID.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *serverlessCacheResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data serverlessCacheResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ElastiCacheClient(ctx)

	tflog.Debug(ctx, "deleting ElastiCache Serverless Cache", map[string]any{
		names.AttrID: data.ID.ValueString(),
	})

	input := &elasticache.DeleteServerlessCacheInput{
		ServerlessCacheName: fwflex.StringFromFramework(ctx, data.ID),
		FinalSnapshotName:   nil,
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 5*time.Minute, func() (any, error) {
		return conn.DeleteServerlessCache(ctx, input)
	}, errCodeDependencyViolation)

	if errs.IsA[*awstypes.ServerlessCacheNotFoundFault](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting ElastiCache Serverless Cache (%s)", data.ID.ValueString()), err.Error())
		return
	}

	if _, err := waitServerlessCacheDeleted(ctx, conn, data.ID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for ElastiCache Serverless Cache (%s) delete", data.ID.ValueString()), err.Error())
		return
	}
}

func findServerlessCache(ctx context.Context, conn *elasticache.Client, input *elasticache.DescribeServerlessCachesInput) (*awstypes.ServerlessCache, error) {
	output, err := findServerlessCaches(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findServerlessCaches(ctx context.Context, conn *elasticache.Client, input *elasticache.DescribeServerlessCachesInput) ([]awstypes.ServerlessCache, error) {
	var output []awstypes.ServerlessCache

	pages := elasticache.NewDescribeServerlessCachesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ServerlessCacheNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.ServerlessCaches...)
	}

	return output, nil
}

func findServerlessCacheByID(ctx context.Context, conn *elasticache.Client, id string) (*awstypes.ServerlessCache, error) {
	input := &elasticache.DescribeServerlessCachesInput{
		ServerlessCacheName: aws.String(id),
	}

	return findServerlessCache(ctx, conn, input)
}

func statusServerlessCache(conn *elasticache.Client, cacheClusterID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findServerlessCacheByID(ctx, conn, cacheClusterID)

		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.Status), nil
	}
}

const (
	serverlessCacheStatusAvailable = "available"
	serverlessCacheStatusCreating  = "creating"
	serverlessCacheStatusDeleting  = "deleting"
	serverlessCacheStatusModifying = "modifying"
)

func waitServerlessCacheAvailable(ctx context.Context, conn *elasticache.Client, cacheClusterID string, timeout time.Duration) (*awstypes.ServerlessCache, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			serverlessCacheStatusCreating,
			serverlessCacheStatusDeleting,
			serverlessCacheStatusModifying,
		},
		Target:     []string{serverlessCacheStatusAvailable},
		Refresh:    statusServerlessCache(conn, cacheClusterID),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ServerlessCache); ok {
		return output, err
	}

	return nil, err
}

func waitServerlessCacheDeleted(ctx context.Context, conn *elasticache.Client, cacheClusterID string, timeout time.Duration) (*awstypes.ServerlessCache, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			serverlessCacheStatusCreating,
			serverlessCacheStatusDeleting,
			serverlessCacheStatusModifying,
		},
		Target:     []string{},
		Refresh:    statusServerlessCache(conn, cacheClusterID),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ServerlessCache); ok {
		return output, err
	}

	return nil, err
}

type serverlessCacheResourceModel struct {
	framework.WithRegionModel
	ARN                    types.String                                           `tfsdk:"arn"`
	CacheUsageLimits       fwtypes.ListNestedObjectValueOf[cacheUsageLimitsModel] `tfsdk:"cache_usage_limits"`
	CreateTime             timetypes.RFC3339                                      `tfsdk:"create_time"`
	DailySnapshotTime      types.String                                           `tfsdk:"daily_snapshot_time"`
	Description            types.String                                           `tfsdk:"description"`
	Endpoint               fwtypes.ListNestedObjectValueOf[endpointModel]         `tfsdk:"endpoint"`
	Engine                 types.String                                           `tfsdk:"engine"`
	FullEngineVersion      types.String                                           `tfsdk:"full_engine_version"`
	ID                     types.String                                           `tfsdk:"id"`
	KmsKeyID               types.String                                           `tfsdk:"kms_key_id"`
	MajorEngineVersion     types.String                                           `tfsdk:"major_engine_version"`
	ReaderEndpoint         fwtypes.ListNestedObjectValueOf[endpointModel]         `tfsdk:"reader_endpoint"`
	SecurityGroupIDs       fwtypes.SetValueOf[types.String]                       `tfsdk:"security_group_ids"`
	ServerlessCacheName    types.String                                           `tfsdk:"name"`
	SnapshotARNsToRestore  fwtypes.ListValueOf[fwtypes.ARN]                       `tfsdk:"snapshot_arns_to_restore"`
	SnapshotRetentionLimit types.Int64                                            `tfsdk:"snapshot_retention_limit"`
	Status                 types.String                                           `tfsdk:"status"`
	SubnetIDs              fwtypes.SetValueOf[types.String]                       `tfsdk:"subnet_ids"`
	Tags                   tftags.Map                                             `tfsdk:"tags"`
	TagsAll                tftags.Map                                             `tfsdk:"tags_all"`
	Timeouts               timeouts.Value                                         `tfsdk:"timeouts"`
	UserGroupID            types.String                                           `tfsdk:"user_group_id"`
}

func (data *serverlessCacheResourceModel) setID() {
	data.ID = data.ServerlessCacheName
}

func (data *serverlessCacheResourceModel) InitFromID() error {
	data.ServerlessCacheName = data.ID

	return nil
}

type cacheUsageLimitsModel struct {
	DataStorage   fwtypes.ListNestedObjectValueOf[dataStorageModel]   `tfsdk:"data_storage"`
	ECPUPerSecond fwtypes.ListNestedObjectValueOf[ecpuPerSecondModel] `tfsdk:"ecpu_per_second"`
}

type dataStorageModel struct {
	Maximum types.Int64                                  `tfsdk:"maximum"`
	Minimum types.Int64                                  `tfsdk:"minimum"`
	Unit    fwtypes.StringEnum[awstypes.DataStorageUnit] `tfsdk:"unit"`
}

type ecpuPerSecondModel struct {
	Maximum types.Int64 `tfsdk:"maximum"`
	Minimum types.Int64 `tfsdk:"minimum"`
}

type endpointModel struct {
	Address types.String `tfsdk:"address"`
	Port    types.Int64  `tfsdk:"port"`
}
