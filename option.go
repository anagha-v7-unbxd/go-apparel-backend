package app

import (
	"github.com/unbxd/go-base/kit/transport/http"
	"github.com/unbxd/go-base/utils/log"
)

// Option configures the app at initialisation.
type Option func(*App) (err error)

func WithCustomLogger(logger log.Logger) Option {
	return func(s *App) (err error) {
		s.logger = logger
		return
	}
}

// WithHTTPTransport configures the HTTP server.
func WithHTTPTransport(
	host, port string,
	monitor []string,
	opts ...http.TransportOption,
) Option {
	return func(s *App) (err error) {
		options := append(
			[]http.TransportOption{
				http.WithLogger(s.logger),
				http.WithFullDefaults(),
				http.WithMonitors(monitor),
			}, opts...)

		tr, err := http.NewTransport(
			host,
			port,
			options...,
		)

		if err != nil {
			return err
		}

		s.httpTransport = tr
		return
	}
}
