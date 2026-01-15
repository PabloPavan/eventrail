package sse

import (
	"errors"
	"net/http"
)

type Server struct {
	opts Options

	broker    Broker
	publisher *Publisher
	hubs      *hubManager
	handler   http.Handler
}

func NewServer(broker Broker, options Options) (*Server, error) {
	if broker == nil {
		return nil, errors.New("broker cannot be nil")
	}

	if options.Resolver == nil {
		return nil, errors.New("principal resolver cannot be nil")
	}

	if options.Router == nil {
		return nil, errors.New("channel router cannot be nil")
	}

	applyDefaultOptions(&options)

	s := &Server{
		broker:    broker,
		opts:      options,
		publisher: NewPublisher(broker),
	}

	s.hubs = newHubManager(broker, options)
	s.handler = newHandler(s.hubs, options)

	return s, nil
}

func (s *Server) Handler() http.Handler {
	return s.handler
}

func (s *Server) Publisher() *Publisher {
	return s.publisher
}
