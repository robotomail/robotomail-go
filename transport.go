// Package robotomail provides the official Robotomail REST API client.
package robotomail

import (
	"bufio"
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"strings"
	"time"
)

type Options struct {
	APIKey     string
	BaseURL    string
	Timeout    time.Duration
	HTTPClient *http.Client
}
type Client struct {
	apiKey  string
	baseURL string
	timeout time.Duration
	http    *http.Client
}
type Upload struct {
	Data        []byte
	Filename    string
	ContentType string
}
type APIError struct {
	Status  int
	Body    json.RawMessage
	Headers http.Header
	Code    string
	Message string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("Robotomail HTTP %d", e.Status)
}

// Ptr constructs an optional request field. For nullable optional fields use Ptr[map[string]string](nil) to send null.
func Ptr[T any](value T) *T { return &value }

func NewClient(options Options) (*Client, error) {
	if options.BaseURL == "" {
		options.BaseURL = "https://api.robotomail.com/v1"
	}
	u, err := url.Parse(options.BaseURL)
	if err != nil {
		return nil, errors.New("invalid base URL")
	}
	loopback := u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1" || u.Hostname() == "::1"
	if u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || !(u.Scheme == "https" || u.Scheme == "http" && loopback) {
		return nil, errors.New("base URL must be HTTPS, or HTTP on localhost, without credentials, query or fragment")
	}
	if options.Timeout == 0 {
		options.Timeout = 30 * time.Second
	}
	if options.Timeout < 0 {
		return nil, errors.New("timeout must be positive")
	}
	if options.APIKey == "" {
		options.APIKey = os.Getenv("ROBOTOMAIL_API_KEY")
	}
	client := &http.Client{}
	if options.HTTPClient != nil {
		copy := *options.HTTPClient
		client = &copy
	}
	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }
	return &Client{options.APIKey, strings.TrimRight(options.BaseURL, "/"), options.Timeout, client}, nil
}

func validateSegment(value string) error {
	if value == "" || value == "." || value == ".." {
		return errors.New("resource IDs must be nonempty path segments")
	}
	return nil
}
func escapeSegment(value string) string { return url.PathEscape(value) }
func queryValue(value any) string       { return fmt.Sprint(value) }

// VerifyWebhook verifies the original body bytes, before parsing or reserializing JSON.
func VerifyWebhook(payload []byte, signature, secret string) bool {
	if secret == "" || len(signature) != 64 || strings.ToLower(signature) != signature {
		return false
	}
	actual, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(payload)
	return hmac.Equal(mac.Sum(nil), actual)
}

func (c *Client) newRequest(ctx context.Context, method, path string, query url.Values, headers map[string]string, body io.Reader) (*http.Request, error) {
	target := c.baseURL + path
	if len(query) > 0 {
		target += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, target, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "robotomail-go/0.2.0")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return req, nil
}

func decode(response *http.Response, result any) error {
	raw, err := io.ReadAll(io.LimitReader(response.Body, 64*1024*1024+1))
	if err != nil {
		return err
	}
	if len(raw) > 64*1024*1024 {
		return errors.New("Robotomail response exceeds 64 MB")
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		envelope := struct {
			Error string `json:"error"`
			Code  string `json:"code"`
		}{}
		_ = json.Unmarshal(raw, &envelope)
		if !json.Valid(raw) {
			raw, _ = json.Marshal(string(raw))
		}
		return &APIError{response.StatusCode, raw, response.Header.Clone(), envelope.Code, envelope.Error}
	}
	if len(raw) == 0 {
		return nil
	}
	if err := json.Unmarshal(raw, result); err != nil {
		return fmt.Errorf("invalid Robotomail response: %w", err)
	}
	return nil
}

func (c *Client) request(ctx context.Context, method, path string, query url.Values, headers map[string]string, body any, file *Upload, result any) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	var payload bytes.Buffer
	if file != nil {
		if len(file.Data) > 25*1024*1024 {
			return errors.New("attachments must be at most 25 MB")
		}
		if strings.ContainsAny(file.Filename+file.ContentType, "\r\n") {
			return errors.New("invalid upload filename or content type")
		}
		writer := multipart.NewWriter(&payload)
		escaped := strings.NewReplacer("\\", "\\\\", "\"", "\\\"").Replace(file.Filename)
		mime := file.ContentType
		if mime == "" {
			mime = "application/octet-stream"
		}
		part, err := writer.CreatePart(textproto.MIMEHeader{"Content-Disposition": {`form-data; name="file"; filename="` + escaped + `"`}, "Content-Type": {mime}})
		if err != nil {
			return err
		}
		if _, err = part.Write(file.Data); err != nil {
			return err
		}
		if err = writer.Close(); err != nil {
			return err
		}
		headers["Content-Type"] = writer.FormDataContentType()
	} else if body != nil {
		if err := json.NewEncoder(&payload).Encode(body); err != nil {
			return err
		}
		headers["Content-Type"] = "application/json"
	}
	var reader io.Reader
	if payload.Len() > 0 {
		reader = &payload
	}
	req, err := c.newRequest(ctx, method, path, query, headers, reader)
	if err != nil {
		return err
	}
	// No SDK retries. In particular, an ambiguous send failure is returned to the caller.
	response, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	return decode(response, result)
}

type EventFrame struct {
	ID    string
	Event string
	Data  string
}
type EventStream struct {
	body    io.ReadCloser
	scanner *bufio.Scanner
	cancel  context.CancelFunc
	frame   EventFrame
	err     error
	id      string
}

func (s *EventStream) Event() EventFrame { return s.frame }
func (s *EventStream) Err() error        { return s.err }
func (s *EventStream) Close() error      { s.cancel(); return s.body.Close() }

func splitSSELine(data []byte, atEOF bool) (int, []byte, error) {
	for i, b := range data {
		if b == '\n' {
			return i + 1, data[:i], nil
		}
		if b == '\r' {
			if i+1 == len(data) && !atEOF {
				return 0, nil, nil
			}
			if i+1 < len(data) && data[i+1] == '\n' {
				return i + 2, data[:i], nil
			}
			return i + 1, data[:i], nil
		}
	}
	if atEOF {
		return len(data), nil, nil
	}
	return 0, nil, nil
}

// Next reads one complete SSE event. Close the stream when stopping early.
func (s *EventStream) Next() bool {
	data := []string{}
	event := "message"
	size := 0
	for s.scanner.Scan() {
		line := s.scanner.Text()
		size += len(line)
		if size > 1024*1024 {
			s.err = errors.New("SSE event exceeds 1 MB")
			_ = s.Close()
			return false
		}
		if line == "" {
			if len(data) > 0 {
				s.frame = EventFrame{s.id, event, strings.Join(data, "\n")}
				return true
			}
			event = "message"
			size = 0
			continue
		}
		key, value, _ := strings.Cut(line, ":")
		value = strings.TrimPrefix(value, " ")
		switch key {
		case "data":
			data = append(data, value)
		case "event":
			event = value
		case "id":
			if !strings.ContainsRune(value, 0) {
				s.id = value
			}
		}
	}
	s.err = s.scanner.Err()
	_ = s.Close()
	return false
}

func (c *Client) openStream(ctx context.Context, query url.Values, headers map[string]string) (*EventStream, error) {
	// Server reconnects after about 4.5 minutes; give each connection a bounded lifetime.
	timeout := c.timeout
	if timeout < 6*time.Minute {
		timeout = 6 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	headers["Accept"] = "text/event-stream"
	req, err := c.newRequest(ctx, "GET", "/events", query, headers, nil)
	if err != nil {
		cancel()
		return nil, err
	}
	response, err := c.http.Do(req)
	if err != nil {
		cancel()
		return nil, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		defer response.Body.Close()
		defer cancel()
		return nil, decode(response, nil)
	}
	if !strings.HasPrefix(response.Header.Get("Content-Type"), "text/event-stream") {
		response.Body.Close()
		cancel()
		return nil, errors.New("Robotomail returned a non-SSE response")
	}
	scanner := bufio.NewScanner(response.Body)
	scanner.Split(splitSSELine)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	return &EventStream{body: response.Body, scanner: scanner, cancel: cancel}, nil
}
