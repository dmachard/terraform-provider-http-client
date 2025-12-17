package httpclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"strings"
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

/*
Metadata
*/
func (r *EphemeralRequest) Metadata(
	ctx context.Context,
	req ephemeral.MetadataRequest,
	resp *ephemeral.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_request"
}

/*
Schema
*/
func (r *EphemeralRequest) Schema(
	ctx context.Context,
	req ephemeral.SchemaRequest,
	resp *ephemeral.SchemaResponse,
) {
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
		},
	}
}

func (r *EphemeralRequest) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data struct {
		URL      types.String `tfsdk:"url"`
		Method   types.String `tfsdk:"request_method"`
		Headers  types.Map    `tfsdk:"request_headers"`
		Body     types.String `tfsdk:"request_body"`
		Username types.String `tfsdk:"username"`
		Password types.String `tfsdk:"password"`
		Timeout  types.Int64  `tfsdk:"timeout"`

		ResponseCode    types.Int64  `tfsdk:"response_code"`
		ResponseBody    types.String `tfsdk:"response_body"`
		ResponseHeaders types.Map    `tfsdk:"response_headers"`
	}

	// read config
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	method := "GET"
	if !data.Method.IsNull() {
		method = data.Method.ValueString()
	}

	timeout := 10 * time.Second
	if !data.Timeout.IsNull() {
		timeout = time.Duration(data.Timeout.ValueInt64()) * time.Second
	}

	body := []byte{}
	if !data.Body.IsNull() {
		body = []byte(data.Body.ValueString())
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		method,
		data.URL.ValueString(),
		bytes.NewBuffer(body),
	)
	if err != nil {
		resp.Diagnostics.AddError("HTTP request error", err.Error())
		return
	}

	// Headers
	if !data.Headers.IsNull() {
		for k, v := range data.Headers.Elements() {
			httpReq.Header.Set(k, v.(types.String).ValueString())
		}
	}

	// Basic auth
	if !data.Username.IsNull() && data.Username.ValueString() != "" {
		httpReq.SetBasicAuth(
			data.Username.ValueString(),
			data.Password.ValueString(),
		)
	}

	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	rsp, err := client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("HTTP request failed", err.Error())
		return
	}
	defer rsp.Body.Close()

	rspBody, err := io.ReadAll(rsp.Body)
	if err != nil {
		resp.Diagnostics.AddError("Read response failed", err.Error())
		return
	}

	headers := make(map[string]attr.Value)
	for k, v := range rsp.Header {
		headers[k] = types.StringValue(strings.Join(v, ", "))
	}

	data.ResponseCode = types.Int64Value(int64(rsp.StatusCode))
	data.ResponseBody = types.StringValue(string(rspBody))
	data.ResponseHeaders = types.MapValueMust(types.StringType, headers)

	resp.Result.Set(ctx, &data)
}
