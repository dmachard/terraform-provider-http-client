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

For detailed usage see [provider's documentation page](https://registry.terraform.io/providers/dmachard/http-client/latest/docs)


## Testing

Test unit

```bash
export TF_ACC=1
go test -v ./...
```

Local test with terraform

```bash
export VERSION="1.0.0"
export PLUGIN_DIR="$HOME/.terraform.d/plugins/registry.terraform.io/dmachard/http-client/${VERSION}/linux_amd64"
export TF_LOG=DEBUG
export TF_LOG_PATH=./terraform.log

# Build the provider binary
go build -o terraform-provider-http-client

# Create the local plugin directory
mkdir -p "${PLUGIN_DIR}"

# Copy the provider binary
cp terraform-provider-http-client "${PLUGIN_DIR}"

cd examples/
terraform init
terraform plan
terraform apply
cd ..
```