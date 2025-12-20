package httpclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestExecuteRequest_OK(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer ts.Close()

	result, err := ExecuteRequest(RequestConfig{
		Ctx:     context.Background(),
		URL:     ts.URL,
		Method:  "GET",
		Timeout: 5 * time.Second,
	})

	require.NoError(t, err)
	require.Equal(t, 200, result.ResponseCode)
	require.JSONEq(t, `{"ok":true}`, string(result.ResponseBody))
}

func TestExecuteRequest_UnexpectedStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	_, err := ExecuteRequest(RequestConfig{
		Ctx:                 context.Background(),
		URL:                 ts.URL,
		Method:              "GET",
		Timeout:             5 * time.Second,
		ExpectedStatusCodes: []int{200},
		FailOnHTTPError:     true,
	})

	require.Error(t, err)
}

func TestExecuteRequest_Method(t *testing.T) {
	expectedMethod := "POST"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != expectedMethod {
			t.Fatalf("expected method %s, got %s", expectedMethod, r.Method)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	_, err := ExecuteRequest(RequestConfig{
		Ctx:     context.Background(),
		URL:     ts.URL,
		Method:  expectedMethod,
		Body:    []byte(`{"test":true}`),
		Timeout: 5 * time.Second,
	})

	require.NoError(t, err)
}
