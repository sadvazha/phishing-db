package db

import (
	"context"
	"fmt"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

type DBClient struct {
	logger     *slog.Logger
	client     *mongo.Client
	collection *mongo.Collection
}

func NewClient(logger *slog.Logger, client *mongo.Client, databaseName string, collectionName string) *DBClient {
	wcMajority := writeconcern.Majority()
	wcMajorityCollectionOpts := options.Collection().SetWriteConcern(wcMajority)
	collection := client.Database(databaseName).Collection(collectionName, wcMajorityCollectionOpts)

	return &DBClient{
		logger:     logger,
		client:     client,
		collection: collection,
	}
}

func (c *DBClient) Close(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}

func (c *DBClient) Write(ctx context.Context, recordC <-chan *PhishingRecord, done chan<- error) {
	c.logger.Info("starting write")
	session, err := c.client.StartSession()
	if err != nil {
		done <- fmt.Errorf("failed to start session: %w", err)
		return
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessionContext mongo.SessionContext) (interface{}, error) {
		c.logger.Debug("cleaning collection...")
		_, err := c.collection.DeleteMany(sessionContext, bson.D{})
		if err != nil {
			c.logger.Error("failed to clean collection", "error", err.Error())
			return nil, fmt.Errorf("failed to drop collection: %w", err)
		}
		c.logger.Debug("cleaned collection")
		for record := range recordC {
			c.logger.Debug("trying to insert a record", "phish_id", record.PhishID)
			_, err := c.collection.InsertOne(sessionContext, record)
			if err != nil {
				c.logger.Error("failed to insert record", "err", err.Error())
				return nil, fmt.Errorf("failed to insert record: %w", err)
			}
			c.logger.Debug("inserted record", "phish_id", record.PhishID)
		}

		return nil, nil
	})
	if err != nil {
		done <- fmt.Errorf("failed to execute transaction: %w", err)
		return
	}
	c.logger.Info("write completed")

	done <- nil
}
