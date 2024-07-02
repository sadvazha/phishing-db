package main

import (
	"log/slog"
	"os"
)

var (
	configLogLevel = slog.LevelDebug

	configMongoUrl            = os.Getenv("MONGO_URL")
	configMongoDatabaseName   = "phishing"
	configMongoCollectionName = "phishing_records"
)
