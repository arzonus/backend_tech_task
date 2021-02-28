package mongoutil

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/mongo"
)

type Config struct {
	URL      string `envconfig:"url"`
	Database string `envconfig:"database"`
	User     string `envconfig:"user"`
	Password string `envconfig:"password"`
}

func NewDB(cfg Config) (*mongo.Database, error) {
	clientOptions := options.Client().ApplyURI(cfg.URL)
	clientOptions.SetAuth(options.Credential{Username: cfg.User, Password: cfg.Password})

	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}

	if cfg.Database == "" {
		return nil, fmt.Errorf("mongo database cannot be empty")
	}

	return client.Database(cfg.Database), nil
}
