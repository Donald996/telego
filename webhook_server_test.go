package telego

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fasthttp/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func TestFastHTTPWebhookServer_RegisterHandler(t *testing.T) {
	addr := testAddress(t)

	s := FastHTTPWebhookServer{
		Logger: testLoggerType{},
		Server: &fasthttp.Server{},
		Router: router.New(),
	}

	go func() {
		err := s.Start(addr)
		require.NoError(t, err)
	}()

	err := s.RegisterHandler("/", func(data []byte) error {
		if len(data) == 0 {
			return nil
		}

		return errTest
	})
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		ctx.Request.SetRequestURI("/")
		ctx.Request.Header.SetMethod(fasthttp.MethodPost)
		s.Server.Handler(ctx)

		assert.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode())
	})

	t.Run("error_method", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		ctx.Request.SetRequestURI("/")
		ctx.Request.Header.SetMethod(fasthttp.MethodGet)
		s.Server.Handler(ctx)

		assert.Equal(t, fasthttp.StatusMethodNotAllowed, ctx.Response.StatusCode())
	})

	t.Run("error_handler", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		ctx.Request.SetRequestURI("/")
		ctx.Request.Header.SetMethod(fasthttp.MethodPost)
		ctx.Request.SetBody([]byte("err"))
		s.Server.Handler(ctx)

		assert.Equal(t, fasthttp.StatusInternalServerError, ctx.Response.StatusCode())
	})

	err = s.Stop(context.Background())
	assert.NoError(t, err)
}

func TestHTTPWebhookServer_RegisterHandler(t *testing.T) {
	addr := testAddress(t)

	s := HTTPWebhookServer{
		Logger:   testLoggerType{},
		Server:   &http.Server{}, //nolint:gosec
		ServeMux: http.NewServeMux(),
	}

	go func() {
		err := s.Start(addr)
		require.NoError(t, err)
	}()

	time.Sleep(time.Microsecond * 10)

	err := s.Start(addr)
	require.Error(t, err)

	err = s.RegisterHandler("/", func(data []byte) error {
		if len(data) == 0 {
			return nil
		}

		return errTest
	})
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		rc := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", nil)

		s.Server.Handler.ServeHTTP(rc, req)

		assert.Equal(t, http.StatusOK, rc.Code)
	})

	t.Run("error_method", func(t *testing.T) {
		rc := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		s.Server.Handler.ServeHTTP(rc, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rc.Code)
	})

	t.Run("error_handler", func(t *testing.T) {
		rc := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("err"))

		s.Server.Handler.ServeHTTP(rc, req)

		assert.Equal(t, http.StatusInternalServerError, rc.Code)
	})

	t.Run("error_read", func(t *testing.T) {
		rc := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", errReader{})

		s.Server.Handler.ServeHTTP(rc, req)

		assert.Equal(t, http.StatusInternalServerError, rc.Code)
	})

	err = s.Stop(context.Background())
	assert.NoError(t, err)
}

type errReader struct{}

func (e errReader) Read(_ []byte) (n int, err error) {
	return 0, errTest
}

func TestMultiBotWebhookServer_RegisterHandler(t *testing.T) {
	ts := &testServer{}
	s := &MultiBotWebhookServer{
		Server: ts,
	}

	assert.Equal(t, 0, ts.started)
	assert.Equal(t, 0, ts.stopped)
	assert.Equal(t, 0, ts.registered)

	err := s.Start("")
	assert.NoError(t, err)
	assert.Equal(t, 1, ts.started)

	err = s.Start("")
	assert.NoError(t, err)
	assert.Equal(t, 1, ts.started)

	err = s.RegisterHandler("", nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, ts.registered)

	err = s.RegisterHandler("", nil)
	assert.NoError(t, err)
	assert.Equal(t, 2, ts.registered)

	err = s.Stop(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, ts.stopped)

	err = s.Stop(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, ts.stopped)
}

type testServer struct {
	started    int
	stopped    int
	registered int
}

func (t *testServer) Start(_ string) error {
	t.started++
	return nil
}

func (t *testServer) Stop(_ context.Context) error {
	t.stopped++
	return nil
}

func (t *testServer) RegisterHandler(_ string, _ func(data []byte) error) error {
	t.registered++
	return nil
}
