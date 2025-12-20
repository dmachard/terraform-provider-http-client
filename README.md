# HTTP client provider for terraform 

A terraform HTTP provider for interacting with HTTP servers.
It's an alternative to the hashicorp http provider.

## Requirements

-	[Terraform](https://www.terraform.io/downloads.html) > 1.11
-	[Go](https://golang.org/doc/install) >= 1.22

## Features

- Data Source: Traditional HTTP requests with state persistence
- Ephemeral request support
- mTLS support
- Basic authentication support
- Custom headers and request methods
- Custom CA certificates
- Configurable TLS versions from TLS 1.0 to 1.3
- Configurable HTTP versions from HTTP/1.1, HTTP/2 to HTTP/3 over QUIC

## Using the Provider

```hcl
terraform {
  required_providers {
    httpclient = {
      version = "1.0.0"
      source  = "github.com/dmachard/http-client"
    }
  }
}

data "httpclient_request" "req" {
  url = "http://httpbin.org/hidden-basic-auth/user/passwd"
  username = "user"
  password = "passwd"
  request_headers = {
    Content-Type: "application/x-www-form-urlencoded",
  }
  request_body = "scope=access"
}

output "response_body" {
  value = jsondecode(data.httpclient_request.req.response_body).authenticated
}

output "response_code" {
  value = data.httpclient_request.req.response_code
}
```

For detailed usage see [provider's documentation page](https://registry.terraform.io/providers/dmachard/http-client/latest/docs)


## Testing

Local test with terraform

```bash
VERSION="1.0.0"
PLUGIN_DIR="$HOME/.terraform.d/plugins/registry.terraform.io/dmachard/http-client/${VERSION}/linux_amd64"

# Build the provider binary
go build -o terraform-provider-http-client

# Create the local plugin directory
mkdir -p "${PLUGIN_DIR}"

# Copy the provider binary
cp terraform-provider-http-client "${PLUGIN_DIR}"

cd examples/
terraform init

terraform plan
```