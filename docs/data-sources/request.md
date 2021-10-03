---
page_title: "httpclient_request Data source - terraform-provider-http-client"
subcategory: ""
description: |-
  
---

# httpclient_request (Data Source)

The `request` data source makes an HTTP request to the given URL (http or https).

## Example Usage

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

## Argument Reference

### Required

- `url` (String) URL query string to request

### Optionals

- `username` (String) Username for Basic Authentication
- `password` (String) Password for Basic Authentication
- `insecure` (Boolean) Skip certificate validation. Default is `false`
- `request_headers` (String) A map of strings representing additional HTTP headers
- `request_method` (String) Method to use to perform request. Default is `GET`
- `request_body` (String) Body of request to send


## Attributes Reference

The following attributes are exported:

- `response_code` - the HTTP status codes (200, 404, etc.)
- `response_headers` - A map of strings representing the response HTTP headers. 
- `response_body` - The raw body of the HTTP response.
