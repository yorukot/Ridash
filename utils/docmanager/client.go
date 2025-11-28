package docmanager

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	// ErrDocumentNotFound is returned when the document does not exist in the manager.
	ErrDocumentNotFound = errors.New("document not found")
)

// Client interacts with the document manager service.
type Client struct {
	baseURL    string
	apiToken   string
	httpClient *http.Client
}

// DocumentContent mirrors the document manager payload for content lookups.
type DocumentContent struct {
	DocID   int64  `json:"doc_id"`
	Content string `json:"content"`
	Seq     int64  `json:"seq"`
}

// IssueTicketResponse mirrors the document manager payload for ticket issuance.
type IssueTicketResponse struct {
	Ticket    string `json:"ticket"`
	ExpiresAt int64  `json:"expires_at"`
}

// NewClient returns a configured document manager client.
func NewClient(baseURL, apiToken string) (*Client, error) {
	if baseURL == "" {
		return nil, errors.New("doc manager base URL is empty")
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid doc manager base URL: %w", err)
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, fmt.Errorf("doc manager base URL missing scheme or host: %s", baseURL)
	}

	return &Client{
		baseURL:  strings.TrimSuffix(baseURL, "/"),
		apiToken: apiToken,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// GetDocumentContent fetches the latest document content from the manager.
func (c *Client) GetDocumentContent(ctx context.Context, docID int64) (*DocumentContent, error) {
	endpoint, err := url.JoinPath(c.baseURL, "/api/documents", strconv.FormatInt(docID, 10))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	c.applyAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var content DocumentContent
		if err := json.NewDecoder(resp.Body).Decode(&content); err != nil {
			return nil, err
		}
		return &content, nil
	case http.StatusNotFound:
		return nil, ErrDocumentNotFound
	default:
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("document manager returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
}

// DeleteDocument removes all persisted state for a document.
func (c *Client) DeleteDocument(ctx context.Context, docID int64) error {
	endpoint, err := url.JoinPath(c.baseURL, "/api/documents", strconv.FormatInt(docID, 10))
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	c.applyAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNoContent, http.StatusOK:
		return nil
	case http.StatusNotFound:
		return ErrDocumentNotFound
	default:
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("document manager returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
}

// IssueTicket requests a signed edit ticket for the document and user.
func (c *Client) IssueTicket(ctx context.Context, docID int64, userID string) (*IssueTicketResponse, error) {
	endpoint, err := url.JoinPath(c.baseURL, "/api/documents", strconv.FormatInt(docID, 10), "ticket")
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(map[string]string{"user_id": userID})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	c.applyAuth(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var ticket IssueTicketResponse
		if err := json.NewDecoder(resp.Body).Decode(&ticket); err != nil {
			return nil, err
		}
		return &ticket, nil
	case http.StatusNotFound:
		return nil, ErrDocumentNotFound
	default:
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("document manager returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
}

// EditEndpoint builds the document manager edit endpoint with the provided ticket.
func (c *Client) EditEndpoint(ticket string) (*url.URL, error) {
	base, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, err
	}

	path, err := url.Parse("/edit")
	if err != nil {
		return nil, err
	}

	target := base.ResolveReference(path)
	query := target.Query()
	query.Set("ticket", ticket)
	target.RawQuery = query.Encode()

	return target, nil
}

func (c *Client) applyAuth(req *http.Request) {
	if c.apiToken == "" {
		return
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
}
