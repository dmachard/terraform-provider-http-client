terraform {
  required_providers {
    httpclient = {
      version = "1.0.0"
      source  = "dmachard/http-client"
    }
  }
}

ephemeral "httpclient_request" "token" {
  url            = "https://httpbingo.org/post"
  request_method = "POST"
  request_headers = { Content-Type = "application/json" }
  request_body   = jsonencode({ access_token = "faketoken" })
}

ephemeral "httpclient_request" "check" {
  url = "https://httpbingo.org/bearer"
  request_headers = {
    Authorization = "Bearer ${jsondecode(ephemeral.httpclient_request.token.response_body).json.access_token}"
  }
}

data "httpclient_request" "req" {
  url = "http://httpbingo.org/basic-auth/user/passwd"
  username = "user"
  password = "password"
  request_headers = {
    Content-Type: "application/x-www-form-urlencoded",
  }
  request_body = "access=token"
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
