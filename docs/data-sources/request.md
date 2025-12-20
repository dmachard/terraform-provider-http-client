---
page_title: "httpclient_request Data source - terraform-provider-http-client"
subcategory: ""
description: |-
  
---

# httpclient_request (Data Source)

The `request` data source makes an HTTP request to the given URL (http or https).

## Example Usage

### Basic HTTP Request

```terraform
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

### HTTP Request with Basic Authentication

```terraform
data "httpclient_request" "auth_example" {
  url      = "http://httpbin.org/hidden-basic-auth/user/passwd"
  username = "user"
  password = "passwd"

  # Require TLS 1.3 and use HTTP/2
  tls_min_version = "TLS13"
  http_version    = "HTTP2"

  request_headers = {
    Content-Type = "application/x-www-form-urlencoded"
  }
  request_body = "scope=access"
}

output "authenticated" {
  value = jsondecode(data.httpclient_request.auth_example.response_body).authenticated
}
```

### HTTP Request with mTLS Authentication

```terraform
data "httpclient_request" "mtls_example" {
  url            = "https://api.example.com/secure"
  request_method = "GET"
  
  # mTLS configuration
  client_cert = file("${path.module}/certs/client.pem")
  client_key  = file("${path.module}/certs/client-key.pem")
  ca_cert     = file("${path.module}/certs/ca.pem")
  
  # Require TLS 1.3
  tls_min_version = "TLS13"
  
  timeout = 30
}
```

### POST Request with JSON Body

```terraform
data "httpclient_request" "post_example" {
  url            = "https://api.example.com/users"
  request_method = "POST"
  
  request_headers = {
    Content-Type  = "application/json"
    Authorization = "Bearer ${var.api_token}"
  }
  
  request_body = jsonencode({
    name  = "John Doe"
    email = "john@example.com"
  })
}
```

### Request with Custom CA Certificate

```terraform
data "httpclient_request" "custom_ca" {
  url     = "https://internal.company.com/api"
  ca_cert = file("${path.module}/certs/internal-ca.pem")
  
  # Don't skip verification, use our custom CA
  insecure = false
}
```

## Argument Reference

### Required

- `url` (String) URL query string to request

### Optionals

- `username` (String) Username for Basic Authentication
- `password` (String) Password for Basic Authentication

- `insecure` (Boolean) Skip certificate validation. Default is `false`
- `tls_min_version` (String) - Minimum TLS version to accept. Default is `TLS12`. Valid values:
  - `TLS10` - TLS 1.0 (deprecated, use only for legacy systems)
  - `TLS11` - TLS 1.1 (deprecated, use only for legacy systems)
  - `TLS12` - TLS 1.2 (recommended minimum)
  - `TLS13` - TLS 1.3 (most secure)
- `http_version` (String) - HTTP protocol version to use. Default is `HTTP1.1`. Valid values:
  - `HTTP1.1` or `HTTP/1.1` - HTTP/1.1
  - `HTTP2` or `HTTP/2` - HTTP/2

- `request_headers` (String) A map of strings representing additional HTTP headers
- `request_method` (String) Method to use to perform request. Default is `GET`
- `request_body` (String) Body of request to send
- `timeout` (Integer) Timeout in second for HTTP connection. Default is `10` seconds

## Attributes Reference

The following attributes are exported:

- `response_code` - the HTTP status codes (200, 404, etc.)
- `response_headers` - A map of strings representing the response HTTP headers. 
- `response_body` - The raw body of the HTTP response.


## Notes

### mTLS Authentication

Mutual TLS (mTLS) requires both the client and server to authenticate each other using certificates. To use mTLS:

1. Obtain a client certificate and private key from your certificate authority
2. Provide both `client_cert` and `client_key` parameters
3. Optionally provide `ca_cert` if the server uses a private CA

> All certificates and keys must be in PEM format. Example PEM certificate:

### Timeouts

The `timeout` parameter controls the maximum duration for the entire HTTP request, including connection establishment, request sending, and response reading. If the request takes longer than the specified timeout, it will fail with a timeout error.