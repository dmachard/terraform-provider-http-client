# HTTP client provider for terraform 

A terraform HTTP provider for interacting with HTTP servers. It's an alternative to the hashicorp http provider.

## Requirements

-	[Terraform](https://www.terraform.io/downloads.html) > 0.12
-	[Go](https://golang.org/doc/install) >= 1.15

## Using the Provider

```hcl
terraform {
  required_providers {
    httpclient = {
      version = "0.0.3"
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
