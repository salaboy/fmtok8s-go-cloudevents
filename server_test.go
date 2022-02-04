package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/client/test"
	"gotest.tools/v3/assert"
)

func TestServer_Run(t *testing.T) {
	t.Run("stops server on context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		srv, err := New(ctx)
		assert.NilError(t, err, "create server")

		cancel()
		err = srv.Run(ctx)
		assert.Assert(t, errors.Is(err, http.ErrServerClosed))
	})
}

func TestServer_sendCloudEventHandler(t *testing.T) {
	t.Run("fails when http method is not POST", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		srv, err := New(ctx)
		assert.NilError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/produce", nil)
		srv.sendCloudEventHandler(rec, req)

		assert.Equal(t, rec.Code, http.StatusMethodNotAllowed)
	})

	t.Run("sends cloudevent on produce", func(t *testing.T) {
		sender, recvCh := test.NewMockSenderClient(t, 1, ce.WithUUIDs(), ce.WithTimeNow())

		p, err := ce.NewHTTP()
		assert.NilError(t, err)

		srv, err := New(context.Background(), WithCEClient(sender, p))
		assert.NilError(t, err)

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()

			e := <-recvCh
			t.Logf("received event: %v", e)
		}()

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/produce", nil)
		srv.sendCloudEventHandler(rec, req)
		wg.Wait()

		assert.Equal(t, rec.Code, http.StatusOK)
		assert.Equal(t, srv.counter, 1)

		b, err := io.ReadAll(rec.Result().Body)
		assert.NilError(t, err)
		assert.Equal(t, string(b), "OK")
	})
}

func TestServer_consumeCloudEventHandler(t *testing.T) {
	t.Run("consumes cloudevent", func(t *testing.T) {
		ctx := context.Background()

		srv, err := New(ctx)
		assert.NilError(t, err)

		h, err := srv.consumeCloudEventHandler(ctx)
		assert.NilError(t, err)

		rec := httptest.NewRecorder()

		e := ce.NewEvent()
		e.SetID("1")
		e.SetType("test.event")
		e.SetSource("http://test.source")

		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		err = enc.Encode(e)
		assert.NilError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/", &buf)
		req.Header.Add("content-type", ce.ApplicationCloudEventsJSON)

		h.ServeHTTP(rec, req)
		assert.Equal(t, rec.Code, http.StatusOK)
	})
}
