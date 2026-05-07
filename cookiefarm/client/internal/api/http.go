package api

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"client/config"
)

const (
	AUTHED    = true
	NOTAUTHED = false
)

type Client struct {
	baseURL string
	http    *http.Client
}

var (
	instance *Client
	once     sync.Once
)

func getClient() *Client {
	once.Do(func() {
		cm := config.GetInstance()
		baseURL := fmt.Sprintf("http://%s:%d", cm.GetHost(), cm.GetPort())

		instance = &Client{
			baseURL: baseURL,
			http: &http.Client{
				Timeout: 3 * time.Second,
			},
		}
	})

	return instance
}

func (*Client) setToken(token string) {
	cm := config.GetInstance()
	cm.SetToken(token)
}

func (c *Client) doRequest(method, endpoint string, body []byte, authed bool, contentType string) (*http.Response, []byte, error) {
	fullURL := c.baseURL + endpoint

	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, fullURL, reader)
	if err != nil {
		return nil, nil, fmt.Errorf("create request: %w", err)
	}

	cfg := config.GetInstance()

	if authed {
		if cfg.GetHost() == "" {
			return nil, nil, errors.New("missing auth token")
		}
		req.Header.Set("Cookie", "token="+cfg.GetToken())
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, nil, fmt.Errorf("read body: %w", err)
	}

	return resp, respBody, nil
}

func (c *Client) get(endpoint string, authed bool) (*http.Response, []byte, error) {
	return c.doRequest(http.MethodGet, endpoint, nil, authed, "")
}

func (c *Client) postJSON(endpoint string, body []byte, authed bool) (*http.Response, []byte, error) {
	return c.doRequest(http.MethodPost, endpoint, body, authed, "application/json")
}

func (c *Client) uploadFile(endpoint string, filePath string, authed bool) (*http.Response, []byte, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create form file: %v", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return nil, nil, fmt.Errorf("failed to copy file content: %v", err)
	}

	if err := writer.Close(); err != nil {
		return nil, nil, fmt.Errorf("failed to close writer: %v", err)
	}

	return c.doRequest(http.MethodPost, endpoint, body.Bytes(), authed, writer.FormDataContentType())
}

func (c *Client) postForm(endpoint string, data url.Values, authed bool) (*http.Response, []byte, error) {
	return c.doRequest(
		http.MethodPost,
		endpoint,
		[]byte(data.Encode()),
		authed,
		"application/x-www-form-urlencoded",
	)
}

func getCookie(resp *http.Response, name string) (string, error) {
	for _, c := range resp.Cookies() {
		if c.Name == name {
			return c.Value, nil
		}
	}
	return "", fmt.Errorf("%s not found in cookies", name)
}
