package main

import (
	"os"
	"time"
)

var (
	configHTTPRetryLimit = 10

	configMongoUrl            = os.Getenv("MONGO_URL")
	configMongoDatabaseName   = "phishing"
	configMongoCollectionName = "phishing_records"

	configPhishTankUrl     = "http://data.phishtank.com/data/online-valid.json"
	configRequestUserAgent = "phishtank/xyz"
	configFetchPeriod      = time.Hour
)
