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

// main is the entry point of the application. It parses command-line flags,
// sets up a worker pool for concurrent processing, fetches all records from the
// ZenGRC API, and distributes them to the workers for processing.
func main() {
	// Define and parse command-line flags for configuration.
	apiURL := flag.String("api-url", "", "The URL of your ZenGRC API instance (e.g., https://acme.api.zengrc.com).")
	token := flag.String("token", "", "Your ZenGRC API authentication token (key_id:key_secret).")
	outputDir := flag.String("output-dir", "./zengrc_attachments", "The directory where the attachments and metadata will be saved.")
	numWorkers := flag.Int("workers", 5, "The number of concurrent workers to use.")
	overwrite := flag.Bool("overwrite", false, "Overwrite existing files.")
	flag.Parse()

	// Validate that required flags are provided.
	if *apiURL == "" || *token == "" {
		fmt.Println("Error: -api-url and -token flags are required.")
		flag.Usage()
		os.Exit(1)
	}

	// Initialize the ZenGRC API client.
	client := NewClient(*apiURL, *token)

	// Create channels for distributing requests and collecting errors.
	requestsChan := make(chan Request)
	errChan := make(chan error, *numWorkers)
	var wg sync.WaitGroup

	// Start the worker pool. Each worker will process requests from the requestsChan.
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

	// Start a goroutine to fetch all requests from the API and send them to the workers.
	// This runs concurrently with the workers, allowing processing to start as soon as
	// the first page of requests is fetched.
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

			// Handle pagination.
			if resp.Links.Next.Href == "" {
				break
			}
			cursor = resp.Links.Next.Href
		}
		close(requestsChan)
	}()

	// Wait for all workers to finish their jobs, then close the error channel.
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Collect and log any errors that occurred during processing.
	for err := range errChan {
		log.Println(err)
	}
}

// processRequest handles the processing of a single ZenGRC request. It creates a
// directory for the record, saves its metadata, and downloads all associated attachments.
func processRequest(client *Client, request Request, outputDir string, overwrite bool) error {
	fmt.Printf("Processing request: %d - %s\n", request.ID, request.Title)

	// Create a dedicated directory for the record.
	recordDir := filepath.Join(outputDir, fmt.Sprintf("record_%d", request.ID))
	if err := os.MkdirAll(recordDir, 0755); err != nil {
		return fmt.Errorf("error creating directory for record %d: %w", request.ID, err)
	}

	// Fetch and save the full metadata for the record.
	if err := saveMetadata(client, request.ID, recordDir); err != nil {
		return fmt.Errorf("error saving metadata for record %d: %w", request.ID, err)
	}

	// Fetch the list of attachments for the record.
	attachments, err := client.GetAttachments(request.ID)
	if err != nil {
		return fmt.Errorf("error getting attachments for record %d: %w", request.ID, err)
	}

	// Download each attachment.
	for _, attachment := range attachments {
		fmt.Printf("Downloading attachment: %s\n", attachment.Name)
		if err := client.DownloadAttachment(request.ID, attachment, recordDir, overwrite); err != nil {
			log.Printf("Error downloading attachment %s for record %d: %v", attachment.Name, request.ID, err)
		}
	}
	return nil
}

// saveMetadata fetches the full details of a request and saves it as a
// metadata.json file in the specified directory.
func saveMetadata(client *Client, requestID int, dir string) error {
	req, err := client.GetRequestDetails(requestID)
	if err != nil {
		return err
	}

	// Marshal the request details into a nicely formatted JSON string.
	data, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return err
	}

	// Write the metadata to the file.
	return os.WriteFile(filepath.Join(dir, "metadata.json"), data, 0644)
}