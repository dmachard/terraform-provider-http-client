package httpclient

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure implementation
var _ datasource.DataSource = &RequestDataSource{}

type RequestDataSource struct{}

// Constructor
func NewRequestDataSource() datasource.DataSource {
	return &RequestDataSource{}
}

// Metadata
func (d *RequestDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "httpclient_request"
}

// Schema
func (d *RequestDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Required: true,
			},
			"username": schema.StringAttribute{
				Optional: true,
			},
			"password": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
			"insecure": schema.BoolAttribute{
				Optional: true,
			},
			"timeout": schema.Int64Attribute{
				Optional: true,
			},
			"request_headers": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"request_method": schema.StringAttribute{
				Optional: true,
			},
			"request_body": schema.StringAttribute{
				Optional: true,
			},
			"response_headers": schema.MapAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"response_code": schema.Int64Attribute{
				Computed: true,
			},
			"response_body": schema.StringAttribute{
				Computed: true,
			},
			"http_version": schema.StringAttribute{
				Optional: true,
			},
			"client_cert": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
			"client_key": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
			"ca_cert": schema.StringAttribute{
				Optional: true,
			},
			"tls_min_version": schema.StringAttribute{
				Optional: true,
			},
			"expected_status_codes": schema.ListAttribute{
				Optional:    true,
				ElementType: types.Int64Type,
			},
			"fail_on_http_error": schema.BoolAttribute{
				Optional: true,
			},
			"follow_redirects": schema.BoolAttribute{
				Optional: true,
			},
			"max_redirects": schema.Int64Attribute{
				Optional: true,
			},
		},
	}
}

func (d *RequestDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data struct {
		URL                 types.String  `tfsdk:"url"`
		Method              types.String  `tfsdk:"request_method"`
		Headers             types.Map     `tfsdk:"request_headers"`
		Body                types.String  `tfsdk:"request_body"`
		Username            types.String  `tfsdk:"username"`
		Password            types.String  `tfsdk:"password"`
		Timeout             types.Int64   `tfsdk:"timeout"`
		Insecure            types.Bool    `tfsdk:"insecure"`
		HTTPVersion         types.String  `tfsdk:"http_version"`
		ClientCert          types.String  `tfsdk:"client_cert"`
		ClientKey           types.String  `tfsdk:"client_key"`
		CACert              types.String  `tfsdk:"ca_cert"`
		TLSMinVersion       types.String  `tfsdk:"tls_min_version"`
		ExpectedStatusCodes []types.Int64 `tfsdk:"expected_status_codes"`
		FailOnHTTPError     types.Bool    `tfsdk:"fail_on_http_error"`
		FollowRedirects     types.Bool    `tfsdk:"follow_redirects"`
		MaxRedirects        types.Int64   `tfsdk:"max_redirects"`

		ResponseCode    types.Int64  `tfsdk:"response_code"`
		ResponseHeaders types.Map    `tfsdk:"response_headers"`
		ResponseBody    types.String `tfsdk:"response_body"`
	}

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build headers map
	hdrs := map[string]string{}
	if !data.Headers.IsNull() {
		for k, v := range data.Headers.Elements() {
			hdrs[k] = v.(types.String).ValueString()
		}
	}

	// Build expected codes slice
	expCodes := []int{}
	for _, v := range data.ExpectedStatusCodes {
		expCodes = append(expCodes, int(v.ValueInt64()))
	}

	// Default method GET
	method := "GET"
	if !data.Method.IsNull() && data.Method.ValueString() != "" {
		method = data.Method.ValueString()
	}

	// Timeout
	timeout := 10 * time.Second
	if !data.Timeout.IsNull() {
		timeout = time.Duration(data.Timeout.ValueInt64()) * time.Second
	}

	// Execute request
	result, err := ExecuteRequest(RequestConfig{
		Ctx:                 ctx,
		URL:                 data.URL.ValueString(),
		Method:              method,
		Body:                []byte(data.Body.ValueString()),
		Headers:             hdrs,
		Username:            data.Username.ValueString(),
		Password:            data.Password.ValueString(),
		Timeout:             timeout,
		Insecure:            data.Insecure.ValueBool(),
		TLSMinVersion:       data.TLSMinVersion.ValueString(),
		ClientCertPEM:       data.ClientCert.ValueString(),
		ClientKeyPEM:        data.ClientKey.ValueString(),
		CACertPEM:           data.CACert.ValueString(),
		HTTPVersion:         data.HTTPVersion.ValueString(),
		ExpectedStatusCodes: expCodes,
		FailOnHTTPError:     data.FailOnHTTPError.ValueBool(),
		FollowRedirects:     data.FollowRedirects.ValueBool(),
		MaxRedirects:        int(data.MaxRedirects.ValueInt64()),
	})
	if err != nil {
		resp.Diagnostics.AddError("HTTP request failed", err.Error())
		return
	}

	headersAttr := make(map[string]attr.Value)
	for k, v := range result.ResponseHeaders {
		headersAttr[k] = types.StringValue(v)
	}

	// Set result
	data.ResponseCode = types.Int64Value(int64(result.ResponseCode))
	data.ResponseBody = types.StringValue(string(result.ResponseBody))
	data.ResponseHeaders = types.MapValueMust(types.StringType, headersAttr)

	resp.State.Set(ctx, &data)
}
