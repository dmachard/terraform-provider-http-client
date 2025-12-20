package httpclient

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	ephemeralschema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ ephemeral.EphemeralResource = &EphemeralRequest{}

type EphemeralRequest struct{}

func NewEphemeralRequest() ephemeral.EphemeralResource {
	return &EphemeralRequest{}
}

func (r *EphemeralRequest) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_request"
}

func (r *EphemeralRequest) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = ephemeralschema.Schema{
		Attributes: map[string]ephemeralschema.Attribute{
			"url": ephemeralschema.StringAttribute{
				Required: true,
			},
			"request_method": ephemeralschema.StringAttribute{
				Optional: true,
			},
			"request_headers": ephemeralschema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
			"request_body": ephemeralschema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
			"username": ephemeralschema.StringAttribute{
				Optional: true,
			},
			"password": ephemeralschema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
			"timeout": ephemeralschema.Int64Attribute{
				Optional: true,
			},
			"response_code": ephemeralschema.Int64Attribute{
				Computed: true,
			},
			"response_body": ephemeralschema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
			"response_headers": ephemeralschema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Sensitive:   true,
			},
			"insecure": ephemeralschema.BoolAttribute{
				Optional: true,
			},
			"tls_min_version": ephemeralschema.StringAttribute{
				Optional: true,
			},
			"http_version": ephemeralschema.StringAttribute{
				Optional: true,
			},
			"client_cert": ephemeralschema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
			"client_key": ephemeralschema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
			"ca_cert": ephemeralschema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
			"expected_status_codes": ephemeralschema.ListAttribute{
				Optional:    true,
				ElementType: types.Int64Type,
			},
			"fail_on_http_error": ephemeralschema.BoolAttribute{
				Optional: true,
			},
			"follow_redirects": ephemeralschema.BoolAttribute{
				Optional: true,
			},
			"max_redirects": ephemeralschema.Int64Attribute{
				Optional: true,
			},
		},
	}
}

func (r *EphemeralRequest) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
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
		ResponseBody    types.String `tfsdk:"response_body"`
		ResponseHeaders types.Map    `tfsdk:"response_headers"`
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
		Headers:             hdrs,
		Body:                []byte(data.Body.ValueString()),
		Username:            data.Username.ValueString(),
		Password:            data.Password.ValueString(),
		Timeout:             timeout,
		Insecure:            data.Insecure.ValueBool(),
		HTTPVersion:         data.HTTPVersion.ValueString(),
		ClientCertPEM:       data.ClientCert.ValueString(),
		ClientKeyPEM:        data.ClientKey.ValueString(),
		CACertPEM:           data.CACert.ValueString(),
		TLSMinVersion:       data.TLSMinVersion.ValueString(),
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

	data.ResponseCode = types.Int64Value(int64(result.ResponseCode))
	data.ResponseBody = types.StringValue(string(result.ResponseBody))
	data.ResponseHeaders = types.MapValueMust(types.StringType, headersAttr)

	resp.Result.Set(ctx, &data)
}
