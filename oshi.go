package oshi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

const (
	defaultEndpoint = "https://oshi.at"

	expire      = "expire"
	autodestroy = "autodestroy"
	randomizefn = "randomizefn"
	filename    = "filename"
	shorturl    = "shorturl"

	onion   = "onion"
	hashsum = "hashsum"

	hashsumMatches = 3
)

var uploadResponseRegex = regexp.MustCompile(`(https?://[^\s]+)\s+\[([^\]]+)\]`)

var hashsumResponseRegex = regexp.MustCompile(`([0-9a-zA-Z]+)\s+\(([^\]]+)\)`)

var ErrWrongResponse = errors.New("wrong response")

type Error struct {
	StatusCode int
	Message    string
}

func (e *Error) Error() string {
	return fmt.Sprintf("Status %d: %s", e.StatusCode, e.Message)
}

type Image struct {
	file        io.Reader
	filename    string
	expire      uint64
	autodestroy bool
	randomizefn bool
	shorturl    bool
}

func NewImage(
	file []byte,
	filename string,
	expire uint64,
	autodestroy bool,
	randomizefn bool,
	shorturl bool,
) *Image {
	return &Image{
		file:        bytes.NewReader(file),
		filename:    filename,
		expire:      expire,
		autodestroy: autodestroy,
		randomizefn: randomizefn,
		shorturl:    shorturl,
	}
}

type UploadResponse struct {
	Admin       string
	Download    string
	TorDownload string
}

type GetHashsumResponse struct {
	Algorithm string
	Hashsum   string
}

// Option helps to confgurate oshi client.
type Option func(client *Client)

// WithEndpoint sets new custom endpoint to oshi client.
func WithEndpoint(endpoint string) Option {
	return func(client *Client) {
		client.endpoint = endpoint
	}
}

// Client is an oshi.at api client.
type Client struct {
	endpoint string

	httpClient *http.Client
}

// NewClient create a new oshi.at api client.
func NewClient(httpClient *http.Client, opts ...Option) *Client {
	client := &Client{
		endpoint:   defaultEndpoint,
		httpClient: httpClient,
	}

	for _, o := range opts {
		o(client)
	}

	return client
}

// GetTorEndpoint gets tor endpoint.
func (c *Client) GetTorEndpoint(ctx context.Context) (string, error) {
	u, err := url.JoinPath(c.endpoint, onion)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", &Error{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	return string(body), nil
}

// GetHashsum gets file hashsum and algorithm.
func (c *Client) GetHashsum(ctx context.Context, file string) (GetHashsumResponse, error) {
	u, err := url.JoinPath(c.endpoint, hashsum, file)
	if err != nil {
		return GetHashsumResponse{}, fmt.Errorf("%w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return GetHashsumResponse{}, fmt.Errorf("%w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return GetHashsumResponse{}, fmt.Errorf("%w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return GetHashsumResponse{}, fmt.Errorf("%w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return GetHashsumResponse{}, &Error{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	return c.parseHashsumResponse(body)
}

func (c *Client) parseHashsumResponse(body []byte) (GetHashsumResponse, error) {
	matches := hashsumResponseRegex.FindStringSubmatch(string(body))

	if len(matches) < hashsumMatches {
		return GetHashsumResponse{}, fmt.Errorf("%w: %s", ErrWrongResponse, string(body))
	}

	return GetHashsumResponse{
		Algorithm: matches[2],
		Hashsum:   matches[1],
	}, nil
}

// Delete delete file.
func (c *Client) Delete(ctx context.Context, adminURL string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, adminURL, nil)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		return &Error{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	return nil
}

// Upload uploads file.
func (c *Client) Upload(ctx context.Context, img *Image) (UploadResponse, error) {
	url, err := c.prepareUploadURL(img)
	if err != nil {
		return UploadResponse{}, fmt.Errorf("%w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, img.file)
	if err != nil {
		return UploadResponse{}, fmt.Errorf("%w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return UploadResponse{}, fmt.Errorf("%w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return UploadResponse{}, fmt.Errorf("%w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return UploadResponse{}, &Error{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	return c.parseUploadResponse(body), nil
}

func (c *Client) prepareUploadURL(img *Image) (string, error) {
	url, err := url.Parse(c.endpoint)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}

	query := url.Query()

	if img.filename != filename {
		query.Set(filename, img.filename)
	}

	if img.expire > 0 {
		query.Set(expire, strconv.FormatUint(img.expire, 10))
	}

	if img.autodestroy {
		query.Set(autodestroy, "1")
	}

	if img.randomizefn {
		query.Set(randomizefn, "1")
	}

	if img.shorturl {
		query.Set(shorturl, "1")
	}

	url.RawQuery = query.Encode()

	return url.String(), nil
}

func (c *Client) parseUploadResponse(body []byte) UploadResponse {
	matches := uploadResponseRegex.FindAllStringSubmatch(string(body), -1)

	response := UploadResponse{}

	for _, match := range matches {
		url := match[1]
		label := strings.ToLower(match[2])

		switch label {
		case "admin":
			response.Admin = url
		case "download":
			response.Download = url
		case "tor download":
			response.TorDownload = url
		}
	}

	return response
}
