package mongoc

import (
	"context"
	"flag"
	sctx "github.com/phathdt/service-context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoClient interface {
	GetClient() *mongo.Client
}

type mongoClient struct {
	id       string
	mongoUri string
	client   *mongo.Client
	logger   sctx.Logger
}

func (m *mongoClient) GetClient() *mongo.Client {
	return m.client
}

func (m *mongoClient) ID() string {
	return m.id
}

func (m *mongoClient) InitFlags() {
	flag.StringVar(&m.mongoUri, "mongo-uri", "mongodb://localhost:27017", "mongo uri")
}

func (m *mongoClient) Activate(sc sctx.ServiceContext) error {
	m.logger = sctx.GlobalLogger().GetLogger(m.id)
	ctx := context.Background()

	clientOptions := options.Client().ApplyURI(m.mongoUri)
	m.client, _ = mongo.Connect(ctx, clientOptions)

	return nil
}

func (m *mongoClient) Stop() error {
	return m.client.Disconnect(context.Background())
}

func New(id string) *mongoClient {
	return &mongoClient{id: id}
}
