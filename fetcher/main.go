package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/sadvazha/phishing-db/fetcher/db"
	"github.com/sadvazha/phishing-db/fetcher/fetcher_service"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ticker := time.NewTicker(configFetchPeriod)
	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(configMongoUrl))
	if err != nil {
		logger.Error("error connecting to MongoDB", "error", err)
		return
	}

	client := db.NewClient(logger, mongoClient, configMongoDatabaseName, configMongoCollectionName)
	defer client.Close(context.Background())

	fetcherService := fetcher_service.NewFetcherService(logger, client, configPhishTankUrl, configRequestUserAgent)

	logger.Info("fetcher started...")
	for ; true; <-ticker.C {
		for {
			retries := 0
			period := time.Minute
			ctx, cancel := context.WithTimeout(context.Background(), period)
			err := fetcherService.FetchAndProcess(ctx)
			cancel()
			if err != nil {
				logger.Error("error fetching and processing stream", "error", err)
				if retries >= configHTTPRetryLimit {
					logger.Error("retries limit reached", "retries", retries)
					break
				}
				time.Sleep(period)
				retries++
				continue
			}
			break
		}
	}
}
