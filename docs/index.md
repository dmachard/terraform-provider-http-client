# HTTP Provider

A terraform HTTP provider for interacting with HTTP servers. It's an alternative to the hashicorp http provider.

## Features

- Data Source: Traditional HTTP requests with state persistence
- Ephemeral request support
- mTLS support
- Basic authentication support
- Custom headers and request methods
- Custom CA certificates
- Configurable TLS versions from TLS 1.0 to 1.3
- Configurable HTTP versions from HTTP/1.1, HTTP/2 to HTTP/3 over QUIC


## Example Usage

```terraform
terraform {
  required_providers {
    httpclient = {
      version = "1.0.0"
      source  = "dmachard/http-client"
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
