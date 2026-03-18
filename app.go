package app

import (
	"context"
	"os"
	"os/signal"

	"github.com/pkg/errors"
	"github.com/unbxd/go-base/kit/transport/http"
	"github.com/unbxd/go-base/utils/log"
	"github.com/unbxd/go-base/utils/metrics"
	"github.com/unbxd/go-base/utils/notifier"
)

type App struct {
	logger        log.Logger        // for logging
	metrics       metrics.Metrics   // for publishing metrics to datadog
	notifier      notifier.Notifier // for publishing events on NATS
	httpTransport *http.Transport   // for serving http traffic

	binders []Binder
}

func (s *App) Listen(errch chan error) {
	err := s.httpTransport.Open()
	if err != nil {
		errch <- errors.Wrap(err, "failed to start transport")
	}
}

func (s *App) Open(cx context.Context) (err error) {
	s.logger.Info("--- Starting Service ---", log.String("addr", s.httpTransport.Addr))

	// define channels
	intch := make(chan os.Signal, 1)
	errch := make(chan error)

	go s.Listen(errch)
	go signal.Notify(intch, os.Interrupt)

	// wait the server for graceful shutdown

	for {
		select {
		case <-intch:
			s.logger.Info(
				"recieved os.Interrupt. shutting down!",
				log.String("signal", os.Interrupt.String()),
			)

			err := s.httpTransport.Close()
			if err != nil {
				panic(err)
			}

			return err
		case er := <-errch:
			s.logger.Error(
				"failed to start http transport",
				log.String("error_message", er.Error()),
				log.Error(er),
			)
			return er
		}
	}
}

func (s *App) Logger() log.Logger { return s.logger }

func NewApp(options ...Option) (*App, error) {
	var (
		err          error
		transport, _ = http.NewTransport("0.0.0.0", "6061")
		logger, _    = log.NewZapLogger()
		metricser    = metrics.NewNoopMetrics()
		notifeir     = notifier.NewNoopNotifier()

		handlerOptions = []http.HandlerOption{}
	)

	app := &App{
		logger:        logger,
		metrics:       metricser,
		notifier:      notifeir,
		httpTransport: transport,
		binders:       []Binder{},
	}

	for _, fn := range options {
		err = fn(app)
		if err != nil {
			return nil, err
		}
	}

	// execute binder
	for _, b := range app.binders {
		b.Bind(app.httpTransport, handlerOptions...)
	}

	return app, err
}
