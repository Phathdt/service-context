package asynqw

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	sctx "github.com/phathdt/service-context"
)

type AsynqWorkerComponent interface {
	GetWorker() *asynq.Server
	RegisterHandler(pattern string, handler func(context.Context, *asynq.Task) error)
}

type asynqWorkerEngine struct {
	id       string
	server   *asynq.Server
	mux      *asynq.ServeMux
	logger   sctx.Logger
	redisURI string
}

func New(id string, redisURI string) *asynqWorkerEngine {
	return &asynqWorkerEngine{
		id:       id,
		redisURI: redisURI,
		mux:      asynq.NewServeMux(),
	}
}

func (a *asynqWorkerEngine) ID() string {
	return a.id
}

func (a *asynqWorkerEngine) InitFlags() {
	// Configuration is passed via constructor, no flags needed
}

func (a *asynqWorkerEngine) Activate(sc sctx.ServiceContext) error {
	a.logger = sc.Logger(a.id)
	a.logger.Info("Starting Asynq worker with Redis at ", a.redisURI, "...")

	opt, err := asynq.ParseRedisURI(a.redisURI)
	if err != nil {
		a.logger.Error("Cannot parse Asynq Redis URI", err.Error())
		return err
	}

	server := asynq.NewServer(
		opt,
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			Logger: &AsynqLogAdapter{logger: a.logger},
		},
	)

	a.server = server

	// Start the server in a goroutine
	go func() {
		if err := a.server.Run(a.mux); err != nil {
			a.logger.Error("Asynq server failed to start", err.Error())
		}
	}()

	a.logger.Info("Asynq worker started successfully")
	return nil
}

func (a *asynqWorkerEngine) Stop() error {
	if a.server != nil {
		a.logger.Info("Shutting down Asynq worker...")
		a.server.Shutdown()
	}
	return nil
}

func (a *asynqWorkerEngine) GetWorker() *asynq.Server {
	return a.server
}

func (a *asynqWorkerEngine) RegisterHandler(pattern string, handler func(context.Context, *asynq.Task) error) {
	a.mux.HandleFunc(pattern, handler)
}

// AsynqLogAdapter adapts service context logger to asynq logger interface
type AsynqLogAdapter struct {
	logger sctx.Logger
}

func (a *AsynqLogAdapter) Debug(args ...interface{}) {
	a.logger.Debug(fmt.Sprint(args...))
}

func (a *AsynqLogAdapter) Info(args ...interface{}) {
	a.logger.Info(fmt.Sprint(args...))
}

func (a *AsynqLogAdapter) Warn(args ...interface{}) {
	a.logger.Warn(fmt.Sprint(args...))
}

func (a *AsynqLogAdapter) Error(args ...interface{}) {
	a.logger.Error(fmt.Sprint(args...))
}

func (a *AsynqLogAdapter) Fatal(args ...interface{}) {
	a.logger.Fatal(fmt.Sprint(args...))
}