package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	ce "github.com/cloudevents/sdk-go/v2"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	httpTimeout = time.Second * 5
	eventSource = "application-b"
	eventType   = "app-b.MyCloudEvent"
)

type myCloudEvent struct {
	Data    string `json:"data"`
	Counter int    `json:"counter"`
}

// Server is an HTTP CloudEvents example implementation
type Server struct {
	log      *zap.Logger
	protocol *cehttp.Protocol
	client   ce.Client
	sink     string

	address string // http bind
	http    *http.Server

	mu      sync.RWMutex
	counter int
}

// New creates a new server
func New(ctx context.Context, opts ...Option) (*Server, error) {
	var srv Server

	for _, do := range defaultOptions {
		if err := do(&srv); err != nil {
			return nil, fmt.Errorf("apply default option: %w", err)
		}
	}

	// custom options
	for _, o := range opts {
		if err := o(&srv); err != nil {
			return nil, fmt.Errorf("apply option: %w", err)
		}
	}

	h, err := srv.consumeCloudEventHandler(ctx)
	if err != nil {
		return nil, fmt.Errorf("create ce consumeCloudEventHandler: %w", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", h)
	mux.HandleFunc("/produce", srv.sendCloudEventHandler)

	httpSrv := http.Server{
		Addr:         srv.address,
		Handler:      mux,
		ReadTimeout:  httpTimeout,
		WriteTimeout: httpTimeout,
	}
	srv.http = &httpSrv

	return &srv, nil
}

func (s *Server) Run(ctx context.Context) error {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		s.log.Info("context cancelled, attempting graceful shutdown")
		if err := s.http.Shutdown(context.Background()); err != nil {
			s.log.Error("could not shutdown gracefully", zap.Error(err))
		}
	}()

	s.log.Info("starting http listener", zap.String("address", s.address))
	return s.http.ListenAndServe()
}

func (s *Server) sendCloudEventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id := uuid.New().String()
	s.log = s.log.With(zap.String("eventID", id))
	s.log.Info("creating cloudevent", zap.String("sink", s.sink))
	event := ce.NewEvent()
	event.SetID(id)
	event.SetSource(eventSource)
	event.SetType(eventType)

	s.mu.Lock()
	defer s.mu.Unlock()
	data := myCloudEvent{
		Data:    "hello from Go",
		Counter: s.counter,
	}

	if err := event.SetData(ce.ApplicationJSON, &data); err != nil {
		s.log.Error("set CloudEvent data", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	const retries = 3
	s.log.Info("sending cloudevent with data", zap.Any("data", data), zap.String("sink", s.sink), zap.Int("maxAttempts", retries))
	ctx := ce.ContextWithRetriesExponentialBackoff(r.Context(), time.Second, retries)
	ctx = ce.ContextWithTarget(ctx, s.sink)

	if result := s.client.Send(ctx, event); !ce.IsACK(result) {
		s.log.Error("failed to send cloudevent", zap.Error(result))
		http.Error(w, "delivery to sink unsuccessful", http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	_, _ = w.Write([]byte("OK"))
	s.counter++
}

func (s *Server) consumeCloudEventHandler(ctx context.Context) (http.Handler, error) {
	return ce.NewHTTPReceiveHandler(ctx, s.protocol, func(_ context.Context, event ce.Event) {
		s.log = s.log.With(zap.String("eventID", event.ID()))
		s.log.Info("received event", zap.String("event", event.String()))

		var e myCloudEvent
		if err := event.DataAs(&e); err != nil {
			s.log.Warn("invalid event e", zap.Error(err))
			return
		}

		s.log.Info("myCloudEvent content", zap.Any("data", e.Data), zap.Int("counter", e.Counter))
	})
}
