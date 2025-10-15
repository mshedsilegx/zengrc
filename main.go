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

	flag.Parse()

	if *apiURL == "" || *token == "" {
		fmt.Println("Error: -api-url and -token flags are required.")
		flag.Usage()
		os.Exit(1)
	}

	client := NewClient(*apiURL, *token)

	requestsChan := make(chan Request)
	var wg sync.WaitGroup

	// Start worker pool
	for i := 0; i < *numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for request := range requestsChan {
				processRequest(client, request, *outputDir)
			}
		}()
	}

	// Fetch requests and send to workers
	var cursor string
	for {
		resp, err := client.GetRequests(cursor)
		if err != nil {
			log.Fatalf("Error getting requests: %v", err)
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
	wg.Wait()
}

func processRequest(client *Client, request Request, outputDir string) {
	fmt.Printf("Processing request: %d - %s\n", request.ID, request.Title)

	recordDir := filepath.Join(outputDir, fmt.Sprintf("record_%d", request.ID))
	if err := os.MkdirAll(recordDir, os.ModePerm); err != nil {
		log.Printf("Error creating directory for record %d: %v", request.ID, err)
		return
	}

	// Save metadata
	if err := saveMetadata(client, request.ID, recordDir); err != nil {
		log.Printf("Error saving metadata for record %d: %v", request.ID, err)
	}

	// Get and download attachments
	attachments, err := client.GetAttachments(request.ID)
	if err != nil {
		log.Printf("Error getting attachments for record %d: %v", request.ID, err)
		return
	}

	for _, attachment := range attachments {
		fmt.Printf("Downloading attachment: %s\n", attachment.Name)
		if err := client.DownloadAttachment(request.ID, attachment, recordDir); err != nil {
			log.Printf("Error downloading attachment %s for record %d: %v", attachment.Name, request.ID, err)
		}
	}
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