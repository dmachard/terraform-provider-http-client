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
		},
	}
}

// Read
func (d *RequestDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Define model
	var data struct {
		URL            types.String `tfsdk:"url"`
		Username       types.String `tfsdk:"username"`
		Password       types.String `tfsdk:"password"`
		Insecure       types.Bool   `tfsdk:"insecure"`
		RequestHeaders types.Map    `tfsdk:"request_headers"`
		RequestMethod  types.String `tfsdk:"request_method"`
		RequestBody    types.String `tfsdk:"request_body"`
		Timeout        types.Int64  `tfsdk:"timeout"`

		ResponseHeaders types.Map    `tfsdk:"response_headers"`
		ResponseCode    types.Int64  `tfsdk:"response_code"`
		ResponseBody    types.String `tfsdk:"response_body"`
	}

	// Read data from Terraform configuration
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare HTTP request
	method := "GET"
	if !data.RequestMethod.IsNull() {
		method = data.RequestMethod.ValueString()
	}
	body := []byte{}
	if !data.RequestBody.IsNull() {
		body = []byte(data.RequestBody.ValueString())
	}

	reqHTTP, err := http.NewRequest(method, data.URL.ValueString(), bytes.NewBuffer(body))
	if err != nil {
		resp.Diagnostics.AddError("Error creating HTTP request", err.Error())
		return
	}

	// Basic auth
	if !data.Username.IsNull() && data.Username.ValueString() != "" {
		reqHTTP.SetBasicAuth(data.Username.ValueString(), data.Password.ValueString())
	}

	// Headers
	if !data.RequestHeaders.IsNull() {
		headers := data.RequestHeaders.Elements()
		for k, v := range headers {
			reqHTTP.Header.Set(k, v.(types.String).ValueString())
		}
	}

	// set timeout
	timeout := 10 * time.Second
	if !data.Timeout.IsNull() {
		timeout = time.Duration(data.Timeout.ValueInt64()) * time.Second
	}

	// Send HTTP request
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: data.Insecure.ValueBool()},
		},
	}

	r, err := client.Do(reqHTTP)
	if err != nil {
		resp.Diagnostics.AddError("Error sending HTTP request", err.Error())
		return
	}
	defer r.Body.Close()

	respBody, err := io.ReadAll(r.Body)
	if err != nil {
		resp.Diagnostics.AddError("Error reading response body", err.Error())
		return
	}

	// Response headers
	rspHeaders := make(map[string]attr.Value)
	for k, v := range r.Header {
		rspHeaders[k] = types.StringValue(strings.Join(v, ", "))
	}

	data.ResponseHeaders = types.MapValueMust(types.StringType, rspHeaders)
	data.ResponseBody = types.StringValue(string(respBody))
	data.ResponseCode = types.Int64Value(int64(r.StatusCode))

	// Set ID
	resp.State.Set(ctx, &data)
}
