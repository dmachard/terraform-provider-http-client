package httpclient

import (
	"context"
	"io"
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

func TestExecuteRequest_ExpectedStatusCodes_OK(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
	}))
	defer ts.Close()

	_, err := ExecuteRequest(RequestConfig{
		Ctx:                 context.Background(),
		URL:                 ts.URL,
		Method:              "GET",
		Timeout:             5 * time.Second,
		ExpectedStatusCodes: []int{200, 201},
		FailOnHTTPError:     true,
	})

	require.NoError(t, err)
}

func TestExecuteRequest_ExpectedStatusCodes_Fail(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
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

func TestExecuteRequest_Headers(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))
		require.Equal(t, "Bearer token123", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	_, err := ExecuteRequest(RequestConfig{
		Ctx:    context.Background(),
		URL:    ts.URL,
		Method: "GET",
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer token123",
		},
		Timeout: 5 * time.Second,
	})

	require.NoError(t, err)
}

func TestExecuteRequest_Body(t *testing.T) {
	expected := `{"name":"john"}`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.JSONEq(t, expected, string(body))
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	_, err := ExecuteRequest(RequestConfig{
		Ctx:     context.Background(),
		URL:     ts.URL,
		Method:  "POST",
		Body:    []byte(expected),
		Timeout: 5 * time.Second,
	})

	require.NoError(t, err)
}

func TestExecuteRequest_BasicAuth(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		require.True(t, ok)
		require.Equal(t, "user", user)
		require.Equal(t, "pass", pass)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	_, err := ExecuteRequest(RequestConfig{
		Ctx:      context.Background(),
		URL:      ts.URL,
		Method:   "GET",
		Username: "user",
		Password: "pass",
		Timeout:  5 * time.Second,
	})

	require.NoError(t, err)
}

func TestExecuteRequest_Timeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	_, err := ExecuteRequest(RequestConfig{
		Ctx:     context.Background(),
		URL:     ts.URL,
		Method:  "GET",
		Timeout: 50 * time.Millisecond,
	})

	require.Error(t, err)
}

func TestExecuteRequest_FollowRedirects(t *testing.T) {
	redirectTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/final" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"ok":true}`))
			return
		}
		http.Redirect(w, r, "/final", http.StatusFound)
	}))
	defer redirectTS.Close()

	result, err := ExecuteRequest(RequestConfig{
		Ctx:             context.Background(),
		URL:             redirectTS.URL,
		Method:          "GET",
		Timeout:         5 * time.Second,
		FollowRedirects: true,
		MaxRedirects:    5,
	})

	require.NoError(t, err)
	require.Equal(t, 200, result.ResponseCode)
	require.JSONEq(t, `{"ok":true}`, string(result.ResponseBody))
}
