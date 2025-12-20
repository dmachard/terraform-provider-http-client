package httpclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
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
			"http_version": schema.StringAttribute{
				Optional:    true,
				Description: "HTTP version to use (HTTP1.1, HTTP2). Default: HTTP1.1",
			},
			// mTLS attributes
			"client_cert": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Client certificate in PEM format for mTLS authentication",
			},
			"client_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Client private key in PEM format for mTLS authentication",
			},
			"ca_cert": schema.StringAttribute{
				Optional:    true,
				Description: "CA certificate in PEM format to verify server certificate",
			},
			"tls_min_version": schema.StringAttribute{
				Optional:    true,
				Description: "Minimum TLS version (1.0, 1.1, 1.2, 1.3). Default: 1.2",
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
		HTTPVersion    types.String `tfsdk:"http_version"`

		// mTLS fields
		ClientCert    types.String `tfsdk:"client_cert"`
		ClientKey     types.String `tfsdk:"client_key"`
		CACert        types.String `tfsdk:"ca_cert"`
		TLSMinVersion types.String `tfsdk:"tls_min_version"`

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

	// Set timeout
	timeout := 10 * time.Second
	if !data.Timeout.IsNull() {
		timeout = time.Duration(data.Timeout.ValueInt64()) * time.Second
	}

	// Configure TLS
	minVersion := uint16(tls.VersionTLS12)
	if !data.TLSMinVersion.IsNull() {
		ver, err := getTLSVersion(data.TLSMinVersion.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Invalid TLS version", err.Error())
			return
		}
		minVersion = ver
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: data.Insecure.ValueBool(),
		MinVersion:         minVersion,
	}

	// Configure mTLS client certificate
	if !data.ClientCert.IsNull() && !data.ClientKey.IsNull() {
		cert, err := tls.X509KeyPair(
			[]byte(data.ClientCert.ValueString()),
			[]byte(data.ClientKey.ValueString()),
		)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error loading client certificate",
				fmt.Sprintf("Failed to parse client certificate and key: %s", err.Error()),
			)
			return
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	// Configure CA certificate for server verification
	if !data.CACert.IsNull() {
		caCertPool := x509.NewCertPool()
		caCertPEM := []byte(data.CACert.ValueString())

		// Parse PEM blocks
		block, _ := pem.Decode(caCertPEM)
		if block == nil {
			resp.Diagnostics.AddError(
				"Error parsing CA certificate",
				"Failed to decode PEM block from CA certificate",
			)
			return
		}

		caCert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error parsing CA certificate",
				fmt.Sprintf("Failed to parse CA certificate: %s", err.Error()),
			)
			return
		}

		caCertPool.AddCert(caCert)
		tlsConfig.RootCAs = caCertPool
	}

	// Send HTTP request
	transport, err := configureHTTPTransport(data.HTTPVersion.ValueString(), tlsConfig)
	if err != nil {
		resp.Diagnostics.AddError("Invalid HTTP version", err.Error())
		return
	}

	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
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
