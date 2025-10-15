package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

func main() {
	apiURL := flag.String("api-url", "", "The URL of your ZenGRC API instance (e.g., https://acme.api.zengrc.com).")
	token := flag.String("token", "", "Your ZenGRC API authentication token (key_id:key_secret).")
	outputDir := flag.String("output-dir", "./zengrc_attachments", "The directory where the attachments and metadata will be saved.")
	numWorkers := flag.Int("workers", 5, "The number of concurrent workers to use.")
	overwrite := flag.Bool("overwrite", false, "Overwrite existing files.")

	flag.Parse()

	if *apiURL == "" || *token == "" {
		fmt.Println("Error: -api-url and -token flags are required.")
		flag.Usage()
		os.Exit(1)
	}

	client := NewClient(*apiURL, *token)

	requestsChan := make(chan Request)
	errChan := make(chan error, *numWorkers)
	var wg sync.WaitGroup

	// Start worker pool
	for i := 0; i < *numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for request := range requestsChan {
				if err := processRequest(client, request, *outputDir, *overwrite); err != nil {
					errChan <- fmt.Errorf("failed to process request %d: %w", request.ID, err)
				}
			}
		}()
	}

	// Fetch requests and send to workers
	go func() {
		var cursor string
		for {
			resp, err := client.GetRequests(cursor)
			if err != nil {
				errChan <- fmt.Errorf("failed to get requests: %w", err)
				break
			}

			for _, request := range resp.Data {
				requestsChan <- request
			}

			if resp.Links.Next.Href == "" {
				break
			}
			cursor = resp.Links.Next.Href
		}
		close(requestsChan)
	}()

	// Wait for all workers to finish and close the error channel
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Collect and log errors
	for err := range errChan {
		log.Println(err)
	}
}

func processRequest(client *Client, request Request, outputDir string, overwrite bool) error {
	fmt.Printf("Processing request: %d - %s\n", request.ID, request.Title)

	recordDir := filepath.Join(outputDir, fmt.Sprintf("record_%d", request.ID))
	if err := os.MkdirAll(recordDir, 0755); err != nil {
		return fmt.Errorf("error creating directory for record %d: %w", request.ID, err)
	}

	// Save metadata
	if err := saveMetadata(client, request.ID, recordDir); err != nil {
		return fmt.Errorf("error saving metadata for record %d: %w", request.ID, err)
	}

	// Get and download attachments
	attachments, err := client.GetAttachments(request.ID)
	if err != nil {
		return fmt.Errorf("error getting attachments for record %d: %w", request.ID, err)
	}

	for _, attachment := range attachments {
		fmt.Printf("Downloading attachment: %s\n", attachment.Name)
		if err := client.DownloadAttachment(request.ID, attachment, recordDir, overwrite); err != nil {
			log.Printf("Error downloading attachment %s for record %d: %v", attachment.Name, request.ID, err)
		}
	}
	return nil
}

func saveMetadata(client *Client, requestID int, dir string) error {
	req, err := client.GetRequestDetails(requestID)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, "metadata.json"), data, 0644)
}