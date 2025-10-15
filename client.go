package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Client is a client for the ZenGRC API.
type Client struct {
	apiURL     string
	token      string
	httpClient *http.Client
}

// NewClient creates a new ZenGRC API client.
func NewClient(apiURL, token string) *Client {
	return &Client{
		apiURL: apiURL,
		token:  token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Request represents a ZenGRC request object.
type Request struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	// Add other fields as needed
}

// Attachment represents a file attachment.
type Attachment struct {
	DocumentID int    `json:"document_id"`
	Name       string `json:"name"`
}

// RequestListResponse is the response from the API when listing requests.
type RequestListResponse struct {
	Data  []Request `json:"data"`
	Links struct {
		Next struct {
			Href string `json:"href"`
		} `json:"next"`
	} `json:"links"`
}

// AttachmentListResponse is the response from the API when listing attachments.
type AttachmentListResponse struct {
	Data struct {
		Files []Attachment `json:"files"`
	} `json:"data"`
}

func (c *Client) newRequest(method, path string, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s%s", c.apiURL, path)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", basicAuth(c.token))
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (c *Client) do(req *http.Request, v interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status: %s, body: %s", resp.Status, string(bodyBytes))
	}

	if v != nil {
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
			return err
		}
	}
	return nil
}

// GetRequestDetails retrieves the details of a single request.
func (c *Client) GetRequestDetails(requestID int) (*Request, error) {
	path := fmt.Sprintf("/api/v2/requests/%d", requestID)
	req, err := c.newRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var request Request
	if err := c.do(req, &request); err != nil {
		return nil, err
	}

	return &request, nil
}

// GetRequests retrieves a list of requests.
func (c *Client) GetRequests(cursor string) (*RequestListResponse, error) {
	path := "/api/v2/requests"
	if cursor != "" {
		path = cursor
	}

	req, err := c.newRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var resp RequestListResponse
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetAttachments retrieves the attachments for a given request.
func (c *Client) GetAttachments(requestID int) ([]Attachment, error) {
	path := fmt.Sprintf("/api/v2/requests/%d/attachments", requestID)
	req, err := c.newRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var resp AttachmentListResponse
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}

	return resp.Data.Files, nil
}

// DownloadAttachment downloads an attachment.
func (c *Client) DownloadAttachment(requestID int, attachment Attachment, outputDir string) error {
	path := fmt.Sprintf("/api/v2/requests/%d/files/%d", requestID, attachment.DocumentID)
	req, err := c.newRequest("GET", path, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status: %s, body: %s", resp.Status, string(bodyBytes))
	}

	// Create the output file
	out, err := os.Create(filepath.Join(outputDir, attachment.Name))
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func basicAuth(token string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(token))
}