package natspub

import (
	"encoding/json"
	"flag"

	"github.com/nats-io/nats.go"
	sctx "github.com/phathdt/service-context"
)

type Component interface {
	Publish(topic string, data interface{}) error
	PublishRaw(topic string, data []byte) error
}

type natsComponent struct {
	id      string
	natsUrl string
	nc      *nats.Conn
	logger  sctx.Logger
}

func (n *natsComponent) Publish(topic string, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if err = n.nc.Publish(topic, payload); err != nil {
		return err
	}

	n.logger.Infof("natsPub %s message = %+v\n", topic, string(payload))

	return nil
}

func (n *natsComponent) PublishRaw(topic string, data []byte) error {
	if err := n.nc.Publish(topic, data); err != nil {
		return err
	}

	n.logger.Infof("natsPub %s message = %+v\n", topic, string(data))

	return nil
}

func (n *natsComponent) ID() string {
	return n.id
}

func (n *natsComponent) InitFlags() {
	flag.StringVar(&n.natsUrl, "nats-pub-url", nats.DefaultURL, "nats publish connection")
}

func (n *natsComponent) Activate(context sctx.ServiceContext) error {
	n.logger = sctx.GlobalLogger().GetLogger(n.id)
	n.logger.Info("connecting")

	nc, err := nats.Connect(n.natsUrl)
	if err != nil {
		return err
	}

	n.nc = nc

	return nil
}

func (n *natsComponent) Stop() error {
	n.nc.Close()

	return nil
}

func New(id string) *natsComponent {
	return &natsComponent{id: id}
}
