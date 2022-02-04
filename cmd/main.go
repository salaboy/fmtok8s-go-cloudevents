package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/kelseyhightower/envconfig"

	"github.com/salaboy/fmtok8s-go-cloudevents"
)

type envConfig struct {
	Address string `envconfig:"BIND_ADDRESS" required:"true" default:":8080"`
	Sink    string `envconfig:"SINK" required:"true" default:"http://localhost:8081"`
	Debug   bool   `envconfig:"DEBUG" default:"false"`
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		panic("process environment variables: " + err.Error())
	}

	var l *zap.Logger
	if env.Debug {
		zapLogger, err := zap.NewDevelopment()
		if err != nil {
			panic("unable to create logger: " + err.Error())
		}
		l = zapLogger

	} else {
		zapLogger, err := zap.NewProduction()
		if err != nil {
			panic("unable to create logger: " + err.Error())
		}
		l = zapLogger
	}
	l = l.Named("go-cloudevents")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer cancel()

	opts := []server.Option{
		server.WithListenAddress(env.Address),
		server.WithLogger(l),
		server.WithSink(env.Sink),
	}
	srv, err := server.New(ctx, opts...)
	if err != nil {
		l.Fatal("could not create server", zap.Error(err))
	}

	if err = srv.Run(ctx); !errors.Is(err, http.ErrServerClosed) {
		l.Fatal("could not run server", zap.Error(err))
	}
}
