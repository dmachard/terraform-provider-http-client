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
			"insecure": ephemeralschema.BoolAttribute{
				Optional:    true,
				Description: "Skip TLS certificate verification",
			},
			"tls_min_version": ephemeralschema.StringAttribute{
				Optional:    true,
				Description: "Minimum TLS version (TLS10, TLS11, TLS12, TLS13). Default: TLS12",
			},
			"http_version": ephemeralschema.StringAttribute{
				Optional:    true,
				Description: "HTTP version to use (HTTP1.1, HTTP2). Default: HTTP1.1",
			},
			// mTLS attributes
			"client_cert": ephemeralschema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Client certificate in PEM format for mTLS authentication",
			},
			"client_key": ephemeralschema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Client private key in PEM format for mTLS authentication",
			},
			"ca_cert": ephemeralschema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "CA certificate in PEM format to verify server certificate",
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

		Insecure types.Bool `tfsdk:"insecure"`

		TLSMinVersion types.String `tfsdk:"tls_min_version"`
		HTTPVersion   types.String `tfsdk:"http_version"`

		// mTLS fields
		ClientCert types.String `tfsdk:"client_cert"`
		ClientKey  types.String `tfsdk:"client_key"`
		CACert     types.String `tfsdk:"ca_cert"`

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

	// Configure TLS
	minVersion := uint16(tls.VersionTLS12)

	// Set TLS minimum version
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
