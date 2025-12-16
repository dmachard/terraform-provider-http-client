terraform {
  required_providers {
    httpclient = {
      version = "1.0.0"
      source  = "dmachard/http-client"
    }
  }
}

data "httpclient_request" "req" {
  url = "http://httpbingo.org/basic-auth/user/passwd"
  username = "user"
  password = "passwd"
  request_headers = {
    Content-Type: "application/x-www-form-urlencoded",
  }
  request_body = "scope=access"
}

output "response_body" {
  value = jsondecode(data.httpclient_request.req.response_body).authorized
}

output "response_code" {
  value = data.httpclient_request.req.response_code
}

output "response_headers" {
  value = data.httpclient_request.req.response_headers
}

