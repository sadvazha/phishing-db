package main

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type DBClient struct {
	client     *mongo.Client
	Collection *mongo.Collection
}

type PhishingRecord struct {
	PhishID            int    `bson:"phish_id"`
	URL                string `bson:"url"`
	PhishDetailURL     string `bson:"phish_detail_url"`
	SubmissionTime     string `bson:"submission_time"`
	SubmissionTimeUnix int64  `bson:"submission_time_unix"`
	Verified           string `bson:"verified"`
	VerificationTime   string `bson:"verification_time"`
	Online             string `bson:"online"`
	Target             string `bson:"target"`
}

func ConnectMongoDB(ctx context.Context, uri, dbName, collectionName string) (*DBClient, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	collection := client.Database(dbName).Collection(collectionName)
	return &DBClient{client: client, Collection: collection}, nil
}

func (c *DBClient) Disconnect(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}
