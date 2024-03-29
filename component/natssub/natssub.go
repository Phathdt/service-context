package natssub

import (
	"flag"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	sctx "github.com/phathdt/service-context"
)

type HandlerFunc func() (jetstream.ConsumeContext, error)

type Component interface {
	GetJs() jetstream.JetStream
	HandleFunc(key string, function HandlerFunc)
}

type natsComp struct {
	id           string
	natsUrl      string
	nc           *nats.Conn
	js           jetstream.JetStream
	logger       sctx.Logger
	consumerFunc map[string]HandlerFunc
	consumerDic  map[string]jetstream.ConsumeContext
}

func (n *natsComp) HandleFunc(key string, function HandlerFunc) {
	n.consumerFunc[key] = function
}

func (n *natsComp) GetJs() jetstream.JetStream {
	return n.js
}

func New(id string) *natsComp {
	return &natsComp{
		id:           id,
		consumerFunc: make(map[string]HandlerFunc),
		consumerDic:  map[string]jetstream.ConsumeContext{},
	}
}

func (n *natsComp) ID() string {
	return n.id
}

func (n *natsComp) InitFlags() {
	flag.StringVar(&n.natsUrl, "nats-sub-url", nats.DefaultURL, "nats subscribe connection")
}

func (n *natsComp) Activate(sc sctx.ServiceContext) error {
	n.logger = sctx.GlobalLogger().GetLogger(n.id)
	n.logger.Info("connecting")

	nc, err := nats.Connect(n.natsUrl)
	if err != nil {
		return err
	}

	js, err := jetstream.New(nc)
	if err != nil {
		return err
	}

	n.nc = nc
	n.js = js

	for key, handlerFunc := range n.consumerFunc {
		consumer, err := handlerFunc()

		if err != nil {
			return err
		}

		n.consumerDic[key] = consumer
	}

	return nil
}

func (n *natsComp) Stop() error {
	for _, c := range n.consumerDic {
		c.Stop()
	}

	return nil
}
