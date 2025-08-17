package pgxc

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	sctx "github.com/phathdt/service-context"
)

type PgxComp interface {
	GetConn() *pgxpool.Pool
}

type pgxComp struct {
	id     string
	prefix string
	dsn    string
	logger sctx.Logger
	pool   *pgxpool.Pool
}

func New(id string, prefix string, dsn string) *pgxComp {
	return &pgxComp{id: id, prefix: prefix, dsn: dsn}
}

func (p *pgxComp) ID() string {
	return p.id
}


func (p *pgxComp) Activate(_ sctx.ServiceContext) error {
	p.logger = sctx.GlobalLogger().GetLogger(p.id)

	p.logger.Info("Connecting to database...")

	config, err := pgxpool.ParseConfig(p.dsn)
	if err != nil {
		p.logger.Error("Cannot parse dsn", err.Error())
		return err
	}

	config.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   &PgxLogAdapter{logger: p.logger},
		LogLevel: tracelog.LogLevelDebug,
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		p.logger.Error("Unable to connect to database", err.Error())
		return err
	}

	if err = pool.Ping(context.Background()); err != nil {
		p.logger.Error("Unable to connect to database", err.Error())
		return err
	}

	p.pool = pool

	return nil
}

func (p *pgxComp) Stop() error {
	p.pool.Close()
	return nil
}

func (p *pgxComp) GetConn() *pgxpool.Pool {
	return p.pool
}
