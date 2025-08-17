package asynqc

import (
	"fmt"

	"github.com/hibiken/asynq"
	sctx "github.com/phathdt/service-context"
)

type AsynqClientComponent interface {
	GetClient() *asynq.Client
}

type asynqClientEngine struct {
	id       string
	client   *asynq.Client
	logger   sctx.Logger
	redisURI string
}

func New(id string, redisURI string) *asynqClientEngine {
	return &asynqClientEngine{id: id, redisURI: redisURI}
}

func (a *asynqClientEngine) ID() string {
	return a.id
}

func (a *asynqClientEngine) Activate(sc sctx.ServiceContext) error {
	a.logger = sc.Logger(a.id)
	a.logger.Info("Connecting to Asynq Redis at ", a.redisURI, "...")

	opt, err := asynq.ParseRedisURI(a.redisURI)
	if err != nil {
		a.logger.Error("Cannot parse Asynq Redis URI", err.Error())
		return err
	}

	client := asynq.NewClient(opt)
	a.client = client

	a.logger.Info("Asynq client connected successfully")
	return nil
}

func (a *asynqClientEngine) Stop() error {
	if a.client != nil {
		if err := a.client.Close(); err != nil {
			return fmt.Errorf("failed to close asynq client: %w", err)
		}
	}
	return nil
}

func (a *asynqClientEngine) GetClient() *asynq.Client {
	return a.client
}
