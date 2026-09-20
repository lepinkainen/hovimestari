package output

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type testTransport func(*http.Request) (*http.Response, error)

func (f testTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// These tests are intentionally sequential because the outputters use DefaultTransport.
func setTestTransport(t *testing.T, transport http.RoundTripper) {
	t.Helper()
	original := http.DefaultTransport
	http.DefaultTransport = transport
	t.Cleanup(func() { http.DefaultTransport = original })
}

func TestDiscordSend(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusNoContent, http.StatusUnauthorized, http.StatusTooManyRequests, http.StatusInternalServerError} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			requests := make(chan map[string]string, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/webhook" || r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("unexpected request: %s %s, headers %v", r.Method, r.URL.Path, r.Header)
				}
				var payload map[string]string
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					t.Error(err)
				}
				requests <- payload
				w.WriteHeader(status)
			}))
			defer server.Close()
			outputter := NewDiscordOutputter(server.URL + "/webhook")
			if outputter.WebhookURL != server.URL+"/webhook" {
				t.Fatal("constructor lost webhook URL")
			}
			content := "**Daily brief**\nHelsinki: 10°C ☀️"
			err := outputter.Send(context.Background(), content)
			checkDeliveryStatus(t, err, status)
			select {
			case payload := <-requests:
				if len(payload) != 1 || payload["content"] != content {
					t.Errorf("payload = %#v", payload)
				}
			default:
				t.Fatal("no webhook request received")
			}
		})
	}
}

func TestTelegramSend(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusBadRequest, http.StatusUnauthorized, http.StatusInternalServerError} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			requests := make(chan map[string]string, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/bottest-token/sendMessage" || r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("unexpected request: %s %s, headers %v", r.Method, r.URL.Path, r.Header)
				}
				var payload map[string]string
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					t.Error(err)
				}
				requests <- payload
				w.WriteHeader(status)
				_, _ = w.Write([]byte(`{"ok":true}`))
			}))
			defer server.Close()
			target, err := url.Parse(server.URL)
			if err != nil {
				t.Fatal(err)
			}
			transport := http.DefaultTransport.(*http.Transport).Clone()
			defer transport.CloseIdleConnections()
			setTestTransport(t, testTransport(func(r *http.Request) (*http.Response, error) {
				if r.URL.Scheme != "https" || r.URL.Host != "api.telegram.org" {
					return nil, fmt.Errorf("unexpected destination: %s", r.URL.Host)
				}
				// Route every request to the local test server, never the real API.
				local := r.Clone(r.Context())
				local.URL.Scheme, local.URL.Host = target.Scheme, target.Host
				return transport.RoundTrip(local)
			}))
			outputter := NewTelegramOutputter("test-token", "test-chat")
			if outputter.BotToken != "test-token" || outputter.ChatID != "test-chat" {
				t.Fatal("constructor lost credentials")
			}
			err = outputter.Send(context.Background(), "Hello.world!\n[Test]")
			checkDeliveryStatus(t, err, status)
			select {
			case payload := <-requests:
				if len(payload) != 3 || payload["chat_id"] != "test-chat" || payload["parse_mode"] != "MarkdownV2" || payload["text"] != "Hello\\.world\\!\n\\[Test\\]" {
					t.Errorf("payload = %#v", payload)
				}
			default:
				t.Fatal("no Telegram request received")
			}
		})
	}
}

func checkDeliveryStatus(t *testing.T, err error, status int) {
	t.Helper()
	if status >= 200 && status < 300 {
		if err != nil {
			t.Fatal(err)
		}
	} else if err == nil || !strings.Contains(err.Error(), fmt.Sprint(status)) {
		t.Fatalf("error = %v, want HTTP status %d", err, status)
	}
}

func TestDeliveryNetworkErrors(t *testing.T) {
	failure := errors.New("simulated connection failure")
	setTestTransport(t, testTransport(func(*http.Request) (*http.Response, error) { return nil, failure }))
	for name, outputter := range map[string]Outputter{
		"discord":  NewDiscordOutputter("https://example.invalid/webhook"),
		"telegram": NewTelegramOutputter("test-token", "test-chat"),
	} {
		t.Run(name, func(t *testing.T) {
			if err := outputter.Send(context.Background(), "test"); !errors.Is(err, failure) {
				t.Errorf("error = %v, want network failure", err)
			}
		})
	}
}

func TestDeliveryInvalidURLs(t *testing.T) {
	setTestTransport(t, testTransport(func(*http.Request) (*http.Response, error) {
		t.Error("invalid URL reached HTTP transport")
		return nil, errors.New("unexpected HTTP request")
	}))
	for name, outputter := range map[string]Outputter{
		"discord":  NewDiscordOutputter("://invalid"),
		"telegram": NewTelegramOutputter("invalid\ntoken", "test-chat"),
	} {
		t.Run(name, func(t *testing.T) {
			if err := outputter.Send(context.Background(), "test"); err == nil || !strings.Contains(err.Error(), "failed to create") {
				t.Errorf("error = %v, want invalid request error", err)
			}
		})
	}
}

func TestDeliveryCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { t.Error("canceled request reached server") }))
	defer server.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := NewDiscordOutputter(server.URL).Send(ctx, "test"); !errors.Is(err, context.Canceled) {
		t.Errorf("error = %v, want cancellation", err)
	}
}
