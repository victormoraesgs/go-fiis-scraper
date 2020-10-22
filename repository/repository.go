package repository

import (
	"context"
	"log"
	"os"
	"time"

	"example.com/go-fiis-scraper/mongodb"

	"go.mongodb.org/mongo-driver/mongo"
)

// MongoDBCollection is ...
type MongoDBCollection struct {
	Name       string
	Collection *mongo.Collection
	Client     *mongodb.Client
	AddQueue   []interface{}
}

// ConsumeAddQueue is ...
func (c *MongoDBCollection) ConsumeAddQueue() {
	if len(c.AddQueue) > 0 {
		items := c.AddQueue
		c.AddQueue = nil

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := c.Collection.InsertMany(ctx, items)
		if err != nil {
			log.Fatal("Error while writting to the database: ", err)
			panic(err.Error())
		}
	}
}

// Add is ...
func (c *MongoDBCollection) Add(data interface{}) {
	c.AddQueue = append(c.AddQueue, data)

	c.ConsumeAddQueue()
}

// GetCollectionFromClient returns collection
func GetCollectionFromClient(name string, client *mongodb.Client) MongoDBCollection {
	if client == nil {
		client = mongodb.GetClient()
	}

	coll := client.MongoDBClient.Database(os.Getenv("MONGODB_DB")).Collection(name)

	collection := MongoDBCollection{Name: name, Collection: coll, Client: client}

	return collection
}

// GetCollection returns collection
func GetCollection(name string) MongoDBCollection {
	client := mongodb.GetClient()

	client.Connect()

	coll := client.MongoDBClient.Database(os.Getenv("MONGODB_DB")).Collection(name)

	collection := MongoDBCollection{Name: name, Collection: coll, Client: client}

	return collection
}
