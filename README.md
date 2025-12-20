# HTTP client provider for terraform 

A terraform HTTP provider for interacting with HTTP servers. It's an alternative to the hashicorp http provider.

## Requirements

-	[Terraform](https://www.terraform.io/downloads.html) > 1.11
-	[Go](https://golang.org/doc/install) >= 1.22

## Features

- **Data Source**: Traditional HTTP requests with state persistence
- **Ephemeral Request**: HTTP requests without state persistence (Terraform 1.11+)
- **mTLS** support
- Basic authentication support
- Custom headers and request methods
- Custom CA certificates
- Configurable TLS versions

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
# Build the provider binary
go build -o terraform-provider-http-client

# Create the local plugin directory
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/dmachard/http-client/1.0.0/linux_amd64/

# Copy the provider binary
cp terraform-provider-http-client ~/.terraform.d/plugins/registry.terraform.io/dmachard/http-client/1.0.0/linux_amd64/

cd examples/
terraform init

terraform plan
```