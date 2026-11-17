package controllers

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	echobasicauth "github.com/etkecc/go-echo-basic-auth/v2"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"
)

type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func unescapeLog(raw string) string {
	return strings.ReplaceAll(raw, `\"`, `"`)
}

func newTestServer(t *testing.T, auth *echobasicauth.Auth, setup ...func(*echo.Echo)) (sink *syncBuffer, baseURL string) {
	t.Helper()
	e := echo.New()
	sink = &syncBuffer{}
	e.Logger = slog.New(slog.NewTextHandler(sink, &slog.HandlerOptions{Level: slog.LevelDebug}))
	ConfigureRouter(e, auth)
	for _, fn := range setup {
		fn(e)
	}

	addrs := make(chan net.Addr, 1)
	sc := echo.StartConfig{
		Address:          "127.0.0.1:0",
		HideBanner:       true,
		HidePort:         true,
		ListenerAddrFunc: func(addr net.Addr) { addrs <- addr },
	}
	ctx, cancel := context.WithCancel(t.Context())
	errCh := make(chan error, 1)
	go func() { errCh <- sc.Start(ctx, e) }()

	var addr net.Addr
	select {
	case addr = <-addrs:
	case err := <-errCh:
		cancel()
		t.Fatalf("server exited before listening: %v", err)
	case <-time.After(5 * time.Second):
		cancel()
		t.Fatal("timed out waiting for the listener")
	}

	t.Cleanup(func() {
		cancel()
		select {
		case err := <-errCh:
			require.NoError(t, err)
		case <-time.After(5 * time.Second):
			t.Error("timed out waiting for server shutdown")
		}
	})
	baseURL = "http://" + addr.String()
	return sink, baseURL
}

func doRequest(t *testing.T, method, url string, header http.Header, login, password string) (status int, respHeader http.Header, body string) {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), method, url, http.NoBody)
	require.NoError(t, err)
	for key, values := range header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	if login != "" || password != "" {
		req.SetBasicAuth(login, password)
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, resp.Header, string(data)
}

func TestRouterBasics(t *testing.T) {
	auth := &echobasicauth.Auth{Login: "admin", Password: "secret", IPs: []string{"127.0.0.1"}}
	sink, baseURL := newTestServer(t, auth)

	status, _, body := doRequest(t, http.MethodGet, baseURL+"/_health", nil, "", "")
	require.Equal(t, http.StatusOK, status)
	require.JSONEq(t, `{"status":"ok"}`, body)

	status, header, _ := doRequest(t, http.MethodGet, baseURL+"/metrics", nil, "", "")
	require.Equal(t, http.StatusUnauthorized, status)
	require.Contains(t, strings.ToLower(header.Get(echo.HeaderWWWAuthenticate)), "basic realm=")

	status, _, _ = doRequest(t, http.MethodGet, baseURL+"/metrics", nil, "admin", "secret")
	require.Equal(t, http.StatusOK, status)

	status, _, _ = doRequest(t, http.MethodGet, baseURL+"/nope", nil, "", "")
	require.Equal(t, http.StatusNotFound, status)

	status, _, _ = doRequest(t, http.MethodPost, baseURL+"/_health", nil, "", "")
	require.Equal(t, http.StatusMethodNotAllowed, status)

	require.Eventually(t, func() bool {
		return strings.Contains(unescapeLog(sink.String()), `"GET /metrics HTTP/1.1" 401`)
	}, 2*time.Second, 10*time.Millisecond)
	require.NotContains(t, unescapeLog(sink.String()), "GET /_health")
}

func TestRouterIPAllowlist(t *testing.T) {
	auth := &echobasicauth.Auth{Login: "admin", Password: "secret", IPs: []string{"10.1.2.3"}}
	_, baseURL := newTestServer(t, auth)

	status, _, _ := doRequest(t, http.MethodGet, baseURL+"/metrics", nil, "admin", "secret")
	require.Equal(t, http.StatusUnauthorized, status)

	allowlisted := http.Header{"X-Forwarded-For": []string{"10.1.2.3"}}
	status, _, _ = doRequest(t, http.MethodGet, baseURL+"/metrics", allowlisted, "admin", "secret")
	require.Equal(t, http.StatusOK, status)

	denied := http.Header{"X-Forwarded-For": []string{"10.9.9.9"}}
	status, _, _ = doRequest(t, http.MethodGet, baseURL+"/metrics", denied, "admin", "secret")
	require.Equal(t, http.StatusUnauthorized, status)
}

func TestRouterLogsServerErrors(t *testing.T) {
	auth := &echobasicauth.Auth{Login: "admin", Password: "secret"}
	sink, baseURL := newTestServer(t, auth, func(e *echo.Echo) {
		e.GET("/boom", func(_ *echo.Context) error {
			return echo.NewHTTPError(http.StatusInternalServerError, "boom")
		})
	})

	status, _, _ := doRequest(t, http.MethodGet, baseURL+"/boom", nil, "", "")
	require.Equal(t, http.StatusInternalServerError, status)

	require.Eventually(t, func() bool {
		logs := unescapeLog(sink.String())
		return strings.Contains(logs, "level=ERROR") &&
			strings.Contains(logs, "request failed") &&
			strings.Contains(logs, "status=500")
	}, 2*time.Second, 10*time.Millisecond)
}
