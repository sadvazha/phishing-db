package fetcher_service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/sadvazha/phishing-db/fetcher/db"
)

type FetcherService struct {
	logger    *slog.Logger
	client    *db.DBClient
	url       string
	userAgent string
}

func NewFetcherService(
	logger *slog.Logger,
	client *db.DBClient,
	url string,
	userAgent string,
) *FetcherService {
	return &FetcherService{
		logger:    logger,
		client:    client,
		url:       url,
		userAgent: userAgent,
	}
}

func (s *FetcherService) FetchAndProcess(ctx context.Context) error {
	recordsC := make(chan *db.PhishingRecord)

	client := &http.Client{}
	req, err := http.NewRequest("GET", s.url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set custom User-Agent header
	req.Header.Set("User-Agent", s.userAgent)

	// Perform the request
	resp, err := client.Do(req)
	if err != nil {
		s.logger.Error("error fetching URL", "url", s.url, "error", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return s.processResponse(ctx, resp.Body, recordsC)
}

func (s *FetcherService) processResponse(ctx context.Context, body io.Reader, recordsC chan *db.PhishingRecord) error {
	decoder := json.NewDecoder(body)
	_, err := decoder.Token() // Read the opening bracket of the array
	if err != nil {
		return err
	}

	done := make(chan error, 1)
	go s.client.Write(ctx, recordsC, done)

	for decoder.More() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-done:
			return err
		default:
			var item db.PhishingRecord
			err := decoder.Decode(&item)
			if err != nil {
				return err
			}
			// Parse the time string using time.RFC3339 layout
			submissionTime, err := time.Parse(time.RFC3339, item.SubmissionTime)
			if err != nil {
				return fmt.Errorf("failed to parse submission time: %w", err)
			}
			// Convert to UNIX seconds
			item.SubmissionTimeUnix = submissionTime.Unix()

			// Process the item here
			recordsC <- &item
			// For example, logging the item or further processing
			s.logger.Info("Processed phishing item", "phishID", item.PhishID)
		}
	}
	close(recordsC)
	s.logger.Info("Pushed all phishing items for processing")

	// Read the closing bracket (end of the array)
	token, err := decoder.Token()
	if err != nil {
		s.logger.Warn("malformed JSON, expected closing brackets", "error", err.Error())
	}

	// Check that the token is the end of an array
	if delim, ok := token.(json.Delim); !ok || delim != ']' {
		s.logger.Warn("malformed JSON, expected end of array in response body", "token", token)
	}

	err = <-done
	if err != nil {
		return fmt.Errorf("worker encountered an error: %v", err)
	}

	return nil
}
