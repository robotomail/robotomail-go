package robotomail

import (
	"bufio"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"
)

var testURL string

func TestMain(m *testing.M) {
	cmd := exec.Command("python3", "test/server.py")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(stdout)
	if !scanner.Scan() {
		panic("fixture did not start")
	}
	testURL = scanner.Text()
	code := m.Run()
	_ = cmd.Process.Kill()
	_ = cmd.Wait()
	os.Exit(code)
}
func testClient(t *testing.T, key string) *Client {
	t.Helper()
	c, err := NewClient(Options{APIKey: key, BaseURL: testURL})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestAllOperations(t *testing.T) {
	raw, err := os.ReadFile("test/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		ID       string          `json:"id"`
		Body     json.RawMessage `json:"body"`
		Params   json.RawMessage `json:"params"`
		Response any             `json:"response"`
	}
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 40 {
		t.Fatal("missing operations")
	}
	for _, fixture := range cases {
		t.Run(fixture.ID, func(t *testing.T) {
			result, err := callCase(testClient(t, "test-key"), fixture.ID, fixture.Body, fixture.Params)
			if err != nil {
				t.Fatal(err)
			}
			if fixture.ID == "streamEvents" {
				stream := result.(*EventStream)
				defer stream.Close()
				var events []EventFrame
				for stream.Next() {
					events = append(events, stream.Event())
				}
				if stream.Err() != nil {
					t.Fatal(stream.Err())
				}
				expected := []EventFrame{{"evt-1", "message.received", "{\"text\":\n\"héllo\"}"}, {"evt-1", "message", "second"}, {"evt-2", "reconnect", "{}"}}
				if !reflect.DeepEqual(events, expected) {
					t.Fatalf("events: %#v", events)
				}
			} else {
				encoded, err := json.Marshal(result)
				if err != nil {
					t.Fatal(err)
				}
				var actual any
				_ = json.Unmarshal(encoded, &actual)
				if !reflect.DeepEqual(actual, fixture.Response) {
					t.Fatalf("response differs: %s", encoded)
				}
			}
		})
	}
}

func TestHTTPErrors(t *testing.T) {
	for _, status := range []int{401, 402, 403, 404, 429, 500} {
		t.Run(queryValue(status), func(t *testing.T) {
			_, err := testClient(t, "error-"+queryValue(status)).ListMailboxes(context.Background())
			var api *APIError
			if !errors.As(err, &api) || api.Status != status || api.Code != "FIXTURE_ERROR" || api.Headers.Get("Retry-After") != "7" {
				t.Fatalf("error: %#v", err)
			}
		})
	}
}
func TestRedirectAndTimeout(t *testing.T) {
	_, err := testClient(t, "redirect").SendMessage(context.Background(), "box", SendMessageRequest{To: []string{"fixture@example.com"}, Subject: "fixture"})
	var api *APIError
	if !errors.As(err, &api) || api.Status != 307 {
		t.Fatalf("redirect: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Millisecond)
	defer cancel()
	_, err = testClient(t, "slow").ListMailboxes(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("timeout: %v", err)
	}
	response, err := http.Get(strings.TrimSuffix(testURL, "/v1") + "/__requests")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	raw, _ := io.ReadAll(response.Body)
	var requests []struct{ Auth, Path string }
	_ = json.Unmarshal(raw, &requests)
	count := 0
	for _, r := range requests {
		if r.Auth == "Bearer redirect" {
			count++
		}
		if r.Path == "/__unexpected" {
			t.Fatal("followed redirect")
		}
	}
	if count != 1 {
		t.Fatal("retried send")
	}
}
func TestNullSerialization(t *testing.T) {
	empty, _ := json.Marshal(UpdateWebhookRequest{})
	if string(empty) != "{}" {
		t.Fatal(string(empty))
	}
	explicit, _ := json.Marshal(UpdateWebhookRequest{Headers: Ptr[*WebhookHeaders](nil)})
	if string(explicit) != `{"headers":null}` {
		t.Fatal(string(explicit))
	}
	zero, _ := json.Marshal(ListMessagesParams{Offset: Ptr[int64](0)})
	if string(zero) != `{"offset":0}` {
		t.Fatal(string(zero))
	}
}
func TestWebhookVerification(t *testing.T) {
	raw := []byte("{\"text\":\"héllo\"}\n")
	secret := "fixture-secret"
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(raw)
	signature := hex.EncodeToString(mac.Sum(nil))
	if !VerifyWebhook(raw, signature, secret) {
		t.Fatal("valid signature rejected")
	}
	for _, bad := range []string{"", strings.ToUpper(signature), strings.Repeat("g", 64), signature + "0"} {
		if VerifyWebhook(raw, bad, secret) {
			t.Fatal("accepted invalid signature")
		}
	}
	if VerifyWebhook(bytesTrim(raw), signature, secret) || VerifyWebhook(raw, signature, "wrong") {
		t.Fatal("accepted tampering")
	}
}
func bytesTrim(b []byte) []byte { return []byte(strings.TrimSpace(string(b))) }
func TestValidation(t *testing.T) {
	for _, base := range []string{"http://example.com/v1", "https://user:pass@example.com/v1", "https://example.com/v1?key=x"} {
		if _, err := NewClient(Options{BaseURL: base}); err == nil {
			t.Fatal("accepted unsafe URL")
		}
	}
	for _, id := range []string{"", ".", ".."} {
		if _, err := testClient(t, "test-key").GetMailbox(context.Background(), id); err == nil {
			t.Fatal("accepted dot segment")
		}
	}
}
