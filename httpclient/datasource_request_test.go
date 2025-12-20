package httpclient

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRequestDataSource_basic(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "httpclient_request" "test" {
  url = "%s"
}
`, ts.URL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.httpclient_request.test", "response_code", "200"),
					resource.TestCheckResourceAttr("data.httpclient_request.test", "response_body", `{"status":"ok"}`),
				),
			},
		},
	})
}

func TestAccRequestDataSource_postMethod(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(201)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "httpclient_request" "test" {
  url = "%s"
  request_method = "POST"
  request_body   = "{\"hello\":\"world\"}"
}
`, ts.URL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.httpclient_request.test", "response_code", "201"),
					resource.TestCheckResourceAttr("data.httpclient_request.test", "response_body", `{"ok":true}`),
				),
			},
		},
	})
}

func TestAccRequestDataSource_headers(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Test") != "yes" {
			t.Fatalf("missing header X-Test")
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"header_ok":true}`))
	}))
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "httpclient_request" "test" {
  url = "%s"
  request_headers = {
    "X-Test" = "yes"
  }
}
`, ts.URL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.httpclient_request.test", "response_code", "200"),
					resource.TestCheckResourceAttr("data.httpclient_request.test", "response_body", `{"header_ok":true}`),
				),
			},
		},
	})
}

func TestAccRequestDataSource_basicAuth(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "user" || pass != "pass" {
			w.WriteHeader(401)
			w.Write([]byte(`{"authenticated":false}`))
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"authenticated":true}`))
	}))
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "httpclient_request" "test" {
  url      = "%s"
  username = "user"
  password = "pass"
}
`, ts.URL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.httpclient_request.test", "response_code", "200"),
					resource.TestCheckResourceAttr("data.httpclient_request.test", "response_body", `{"authenticated":true}`),
				),
			},
		},
	})
}
