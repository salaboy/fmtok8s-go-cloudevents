package server

import (
	"errors"
	"fmt"

	ce "github.com/cloudevents/sdk-go/v2"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"go.uber.org/zap"
)

const (
	defaultAddress = ":8080"
	defaultSink    = "http://localhost:8081"
)

// Option configures the server
type Option func(s *Server) error

var defaultOptions = []Option{
	WithCEClient(nil, nil),
	WithListenAddress(defaultAddress),
	WithLogger(nil),
	WithSink(defaultSink),
}

// WithCEClient uses the specified cloudevents client and cloudevents http
// protocol. If both are nil a default cloudevents http client and protocol are
// used.
func WithCEClient(client ce.Client, p *cehttp.Protocol) Option {
	return func(s *Server) error {
		if client == nil && p == nil {
			proto, err := ce.NewHTTP()
			if err != nil {
				return fmt.Errorf("create cloudevent protocol: %w", err)
			}

			c, err := ce.NewClient(proto, ce.WithTimeNow(), ce.WithUUIDs())
			if err != nil {
				return fmt.Errorf("create cloudevent client: %w", err)
			}
			s.protocol = proto
			s.client = c
			return nil
		}

		if (client != nil && p == nil) || (client == nil && p != nil) {
			return errors.New("one parameter is nil")
		}

		s.protocol = p
		s.client = client
		return nil
	}
}

// WithListenAddress sets the http bind address
func WithListenAddress(address string) Option {
	return func(s *Server) error {
		if address == "" {
			return errors.New("address empty")
		}
		s.address = address
		return nil
	}
}

// WithLogger uses the specified zap logger. If l is nil a default zap
// production logger is used.
func WithLogger(l *zap.Logger) Option {
	return func(s *Server) error {
		if l == nil {
			logger, err := zap.NewProduction()
			if err != nil {
				panic("create zap logger: " + err.Error())
			}
			s.log = logger
			return nil
		}

		s.log = l
		return nil
	}
}

// WithSink sets the sink used to send cloudevents to
func WithSink(sink string) Option {
	return func(s *Server) error {
		if sink == "" {
			return errors.New("sink empty")
		}
		s.sink = sink
		return nil
	}
}
