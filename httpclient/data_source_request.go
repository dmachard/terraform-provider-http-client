package httpclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceRequest() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRequestRead,
		Schema: map[string]*schema.Schema{
			"url": {
				Type:     schema.TypeString,
				Required: true,
			},
			"username": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"password": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"insecure": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"request_headers": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"request_method": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "GET",
			},
			"request_body": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  nil,
			},
			"response_headers": {
				Type:     schema.TypeMap,
				Computed: true,
			},
			"response_code": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"response_body": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceRequestRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// get vars
	url := d.Get("url").(string)
	method := d.Get("request_method").(string)
	body := []byte(d.Get("request_body").(string))
	req_headers := d.Get("request_headers").(map[string]interface{})
	username := d.Get("username").(string)
	password := d.Get("password").(string)

	insecure := d.Get("insecure").(bool)

	// warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// init http request
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return diag.FromErr(err)
	}

	// set basic auth ?
	if len(username) > 0 {
		req.SetBasicAuth(username, password)
	}

	// add headers
	for name, value := range req_headers {
		req.Header.Set(name, value.(string))
	}

	// init go client and send request
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure,
		},
	}

	client := &http.Client{Transport: tr, Timeout: 10 * time.Second}
	r, err := client.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer r.Body.Close()

	// read response body
	rsp_body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return diag.FromErr(err)
	}

	// get headers from response
	rsp_headers := make(map[string]string)
	for k, v := range r.Header {
		rsp_headers[k] = strings.Join(v, ", ")
	}

	// set data resource
	d.Set("response_code", r.StatusCode)
	d.Set("response_body", string(rsp_body))
	d.Set("response_headers", rsp_headers)
	d.SetId(url)

	return diags
}
