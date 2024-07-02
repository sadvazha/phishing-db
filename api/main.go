package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
)

var dbClient *DBClient

func main() {
	var err error
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: configLogLevel}))
	ctx := context.Background()
	dbClient, err = ConnectMongoDB(ctx, configMongoUrl, configMongoDatabaseName, configMongoCollectionName)
	if err != nil {
		logger.Error("failed to connect to mongo db", "error", err.Error())
	}
	defer dbClient.Disconnect(context.Background())

	r := mux.NewRouter()
	r.HandleFunc("/download_report", DownloadReportHandler).Methods("GET")
	r.HandleFunc("/search_domain", SearchDomainHandler).Methods("GET")

	http.Handle("/", r)
	logger.Info("Server will be running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func DownloadReportHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: level should be configurable
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: configLogLevel})).With("handler", "DownloadReportHandler")
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	if fromStr == "" {
		http.Error(w, "Missing 'from' parameter", http.StatusBadRequest)
		return
	}

	fromSec, err := strconv.ParseInt(fromStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid 'from' parameter", http.StatusBadRequest)
		return
	}
	from := time.Unix(fromSec, 0).Unix()

	to := time.Now().Unix()
	if toStr != "" {
		toSec, err := strconv.ParseInt(toStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid 'to' parameter", http.StatusBadRequest)
			return
		}
		to = time.Unix(toSec, 0).Unix()
	}

	filter := bson.M{
		"submission_time_unix": bson.M{
			"$gte": from,
			"$lte": to,
		},
	}
	logger.Debug("filter params", "from", from, "to", to)
	cursor, err := dbClient.Collection.Find(r.Context(), filter)
	if err != nil {
		http.Error(w, "failed to fetch records", http.StatusInternalServerError)
		logger.Error("failed to fetch records", "error", err.Error())
		return
	}
	defer cursor.Close(r.Context())

	type TLDRecord struct {
		TLD         string
		Occurrences int
	}
	type Report struct {
		URLs         []string
		TotalResults int
		SortedTLDs   []TLDRecord
	}

	tlds := make(map[string]int)
	report := Report{
		URLs: make([]string, 0),
	}

	// Processing items one by one, and append to the report
	for cursor.Next(r.Context()) {
		var record PhishingRecord
		if err := cursor.Decode(&record); err != nil {
			http.Error(w, "failed to decode record", http.StatusInternalServerError)
			logger.Error("failed to decode record", "error", err.Error())
			return
		}
		logger.Debug("found record", "url", record.URL)
		report.URLs = append(report.URLs, record.URL)
		tld, err := extractTLD(record.URL)
		if err != nil {
			logger.Warn("failed to extract TLD", "error", err.Error())
			continue
		}
		tlds[tld]++
	}

	report.TotalResults = len(report.URLs)

	for tld, occurrences := range tlds {
		report.SortedTLDs = append(report.SortedTLDs, TLDRecord{TLD: tld, Occurrences: occurrences})
	}

	sort.Slice(report.SortedTLDs, func(i, j int) bool {
		return report.SortedTLDs[i].Occurrences > report.SortedTLDs[j].Occurrences
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func SearchDomainHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: level should be configurable
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: configLogLevel})).With("handler", "SearchDomainHandler")
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		http.Error(w, "Missing 'domain' parameter", http.StatusBadRequest)
		return
	}
	// Do not use unsanitized input in regex
	domain = regexp.QuoteMeta(domain)

	ctx := r.Context()
	logger.Debug("filter params", "url_to_search", domain)
	filter := bson.M{"url": bson.M{"$regex": domain, "$options": "i"}}
	cursor, err := dbClient.Collection.Find(ctx, filter)
	if err != nil {
		http.Error(w, "failed to fetch records", http.StatusInternalServerError)
		logger.Error("failed to fetch records", "error", err.Error())
		return
	}
	defer cursor.Close(ctx)

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	for cursor.Next(ctx) {
		var record PhishingRecord
		if err := cursor.Decode(&record); err != nil {
			http.Error(w, "failed to decode record", http.StatusInternalServerError)
			logger.Error("failed to decode record", "error", err.Error())
			return
		}
		logger.Debug("processing record", "url", record.URL, "domain", domain)
		if err := encoder.Encode(record); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("failed to encode record", "error", err.Error())
			return
		}
	}

	if err := cursor.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error("cursor error", "error", err.Error())
		return
	}
}

func extractTLD(inputURL string) (string, error) {
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		return "", err
	}
	hostname := parsedURL.Hostname()

	// Split the hostname into parts
	parts := strings.Split(hostname, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid hostname: %s", hostname)
	}

	tld := parts[len(parts)-1]
	// TODO: we can move this out of this function and compile it once for performance sake
	re := regexp.MustCompile(`^[a-zA-Z]+$`)
	// Verify that the TLD is a word (domain can be an ipv4 or ipv6 address)
	if !re.MatchString(tld) {
		return "", fmt.Errorf("invalid TLD: %s", tld)
	}

	return tld, nil
}
