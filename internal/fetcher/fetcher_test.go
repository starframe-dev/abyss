package fetcher

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponseReadError(t *testing.T) {
	cause := errors.New("read failed")
	err := &ResponseReadError{BytesRead: 3, Err: cause}
	assert.Equal(t, "read failed", err.Error())
	assert.ErrorIs(t, err, cause)
	assert.EqualValues(t, 3, err.BytesRead)
}

func TestFetchContextAndFetch(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, UserAgent, r.Header.Get("User-Agent"))
		assert.Contains(t, r.Header.Get("Accept"), "text/html")
		w.Header().Set("Content-Type", "text/html")
		_, _ = io.WriteString(w, "<html>ok</html>")
	}))
	defer ts.Close()
	body, err := FetchContext(context.Background(), ts.URL)
	require.NoError(t, err)
	assert.Equal(t, "<html>ok</html>", body)
	body, err = Fetch(ts.URL)
	require.NoError(t, err)
	assert.Equal(t, "<html>ok</html>", body)
}

func TestFetchBytesVariants(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.Header.Get("Accept"), "image/")
		w.Header().Set("Content-Type", "image/test")
		_, _ = w.Write([]byte{0, 1, 2})
	}))
	defer ts.Close()
	for _, fetch := range []func(context.Context, string) ([]byte, string, error){FetchBytesContext} {
		body, typ, err := fetch(context.Background(), ts.URL)
		require.NoError(t, err)
		assert.Equal(t, []byte{0, 1, 2}, body)
		assert.Equal(t, "image/test", typ)
	}
	body, typ, err := FetchBytes(ts.URL)
	require.NoError(t, err)
	assert.Equal(t, []byte{0, 1, 2}, body)
	assert.Equal(t, "image/test", typ)
}

func TestFetchCSSAndJS(t *testing.T) {
	for _, tc := range []struct {
		name, accept, want string
		fetch              func(context.Context, string) (string, error)
	}{{"css", "text/css", "body{}", FetchCSS}, {"js", "application/javascript", "return;", FetchJS}} {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Contains(t, r.Header.Get("Accept"), tc.accept)
				_, _ = io.WriteString(w, tc.want)
			}))
			defer ts.Close()
			got, err := tc.fetch(context.Background(), ts.URL)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestGetErrorsAndLimits(t *testing.T) {
	_, _, err := get(context.Background(), "://bad", 10, "*")
	require.Error(t, err)
	for _, tc := range []struct {
		name    string
		handler http.HandlerFunc
		want    error
	}{{"status", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNotFound) }, nil},
		{"oversize", func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, strings.Repeat("x", 11)) }, ErrResponseTooLarge}} {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(tc.handler)
			defer ts.Close()
			_, _, err := get(context.Background(), ts.URL, 10, "*")
			require.Error(t, err)
			if tc.want != nil {
				assert.ErrorIs(t, err, tc.want)
			} else {
				assert.Contains(t, err.Error(), fmt.Sprint(http.StatusNotFound))
			}
		})
	}
}

func TestGetContentLengthAndReadError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "11")
		_, _ = io.WriteString(w, "01234567890")
	}))
	defer ts.Close()
	_, _, err := get(context.Background(), ts.URL, 10, "*")
	assert.ErrorIs(t, err, ErrResponseTooLarge)
}

func TestFetchContextCancellation(t *testing.T) {
	entered := make(chan struct{})
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		<-r.Context().Done()
	}))
	defer ts.Close()
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() { _, err := FetchContext(ctx, ts.URL); result <- err }()
	<-entered
	cancel()
	select {
	case err := <-result:
		assert.ErrorIs(t, err, context.Canceled)
	case <-time.After(2 * time.Second):
		t.Fatal("request did not cancel")
	}
}
