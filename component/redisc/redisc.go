package redisc

import (
	"context"

	sctx "github.com/phathdt/service-context"

	"github.com/redis/go-redis/v9"
)

var (
	defaultRedisMaxActive = 0 // 0 is unlimited max active connection
	defaultRedisMaxIdle   = 10
)

type RedisComponent interface {
	GetClient() *redis.Client
}

type redisEngine struct {
	id        string
	client    *redis.Client
	logger    sctx.Logger
	redisUri  string
	maxActive int
	maxIde    int
}

func New(id string, redisURI string) *redisEngine {
	return &redisEngine{
		id:        id,
		redisUri:  redisURI,
		maxActive: defaultRedisMaxActive,
		maxIde:    defaultRedisMaxIdle,
	}
}

func (r *redisEngine) ID() string {
	return r.id
}

func (r *redisEngine) Activate(sc sctx.ServiceContext) error {
	r.logger = sctx.GlobalLogger().GetLogger(r.id)
	r.logger.Info("Connecting to Redis at ", r.redisUri, "...")

	opt, err := redis.ParseURL(r.redisUri)

	if err != nil {
		r.logger.Error("Cannot parse Redis ", err.Error())
		return err
	}

	opt.PoolSize = r.maxActive
	opt.MinIdleConns = r.maxIde

	client := redis.NewClient(opt)

	// Ping to test Redis connection
	if err = client.Ping(context.Background()).Err(); err != nil {
		r.logger.Error("Cannot connect Redis. ", err.Error())
		return err
	}

	// Connect successfully, assign client to goRedisDB
	r.client = client
	return nil
}

func (r *redisEngine) Stop() error {
	if err := r.client.Close(); err != nil {
		return err
	}

	return nil
}

func (r *redisEngine) GetClient() *redis.Client {
	return r.client
}
