package sctx

import (
	"fmt"
)

const (
	DevEnv = "dev"
	StgEnv = "stg"
	PrdEnv = "prd"
)

type Component interface {
	ID() string
	Activate(ServiceContext) error
	Stop() error
}

type ServiceContext interface {
	Load() error
	MustGet(id string) any
	Get(id string) (any, bool)
	Logger(prefix string) Logger
	Stop() error
}

type serviceCtx struct {
	name       string
	env        string
	components []Component
	store      map[string]Component
	logger     Logger
}

func NewServiceContext(opts ...Option) ServiceContext {
	sv := &serviceCtx{
		store: make(map[string]Component),
		env:   DevEnv, // Default environment
	}

	sv.components = []Component{defaultLogger}

	for _, opt := range opts {
		opt(sv)
	}

	sv.logger = defaultLogger.GetLogger(sv.name)

	return sv
}

func (s *serviceCtx) Get(id string) (any, bool) {
	c, ok := s.store[id]

	if !ok {
		return nil, false
	}

	return c, true
}

func (s *serviceCtx) MustGet(id string) any {
	c, ok := s.Get(id)

	if !ok {
		panic(fmt.Sprintf("can not get %s\n", id))
	}

	return c
}

func (s *serviceCtx) Load() error {
	s.logger.Info("Service context is loading...")

	for _, c := range s.components {
		if err := c.Activate(s); err != nil {
			return err
		}
	}

	return nil
}

func (s *serviceCtx) Logger(prefix string) Logger {
	return defaultLogger.GetLogger(prefix)
}

func (s *serviceCtx) Stop() error {
	s.logger.Info("Stopping service context")
	for i := range s.components {
		if err := s.components[i].Stop(); err != nil {
			return err
		}
	}

	s.logger.Info("Service context stopped")

	return nil
}

type Option func(*serviceCtx)

func WithName(name string) Option {
	return func(s *serviceCtx) { s.name = name }
}

func WithComponent(c Component) Option {
	return func(s *serviceCtx) {
		if _, ok := s.store[c.ID()]; !ok {
			s.components = append(s.components, c)
			s.store[c.ID()] = c
		}
	}
}
