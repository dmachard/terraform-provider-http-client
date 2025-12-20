terraform {
  required_providers {
    httpclient = {
      version = "1.0.0"
      source  = "dmachard/http-client"
    }
  }
}

data "httpclient_request" "req_expect" {
  url = "http://httpbin.org/status/200"
  request_method = "GET"

  expected_status_codes = [200]
  fail_on_http_error   = true
}

output "response_code_valid" {
  value = data.httpclient_request.req_expect.response_code
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
  password = "passwd"
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
