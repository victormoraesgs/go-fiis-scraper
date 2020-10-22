package mongodb

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client is ...
type Client struct {
	MongoDBClient *mongo.Client
	URI           string
}

// Connect returns nil
func (c Client) Connect() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := c.MongoDBClient.Connect(ctx)
	if err != nil {
		panic(err.Error())
	}
}

// GetClient returns &client
func GetClient() *Client {
	URI := os.Getenv("MONGODB_URI")
	if len(URI) == 0 {
		log.Fatal("MONGODB_URI environment variable should have a value")
		panic("MONGODB_URI environment variable should have a value")
	}
	mongoClient, err := mongo.NewClient(options.Client().ApplyURI(URI))
	if err != nil {
		panic(err.Error())
	}

	client := Client{MongoDBClient: mongoClient, URI: URI}

	return &client
}

func createClient() *mongo.Client {
	client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		panic(err.Error())
	}

	return client
}

// Connect returns client
func Connect() *mongo.Client {
	client := createClient()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		panic(err.Error())
	}

	return client
}
