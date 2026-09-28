package accrual

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestClientFetchStatuses(t *testing.T) {
	tests := []struct {
		name        string
		status      string
		wantStatus  string
		wantAccrual bool
	}{
		{name: "registered", status: StatusRegistered, wantStatus: StatusRegistered},
		{name: "processing", status: StatusProcessing, wantStatus: StatusProcessing},
		{name: "invalid", status: StatusInvalid, wantStatus: StatusInvalid},
		{name: "processed", status: StatusProcessed, wantStatus: StatusProcessed, wantAccrual: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
				if tt.wantAccrual {
					return jsonResponse(http.StatusOK, `{"order":"12345678903","status":"`+tt.status+`","accrual":500}`), nil
				}
				return jsonResponse(http.StatusOK, `{"order":"12345678903","status":"`+tt.status+`"}`), nil
			})

			result, err := NewClient("http://accrual.test", &http.Client{Transport: transport}).Fetch(context.Background(), "12345678903")
			if err != nil {
				t.Fatal(err)
			}
			if result.Status != tt.wantStatus {
				t.Fatalf("status = %q, want %q", result.Status, tt.wantStatus)
			}
			if (result.Accrual != nil) != tt.wantAccrual {
				t.Fatalf("accrual presence = %v, want %v", result.Accrual != nil, tt.wantAccrual)
			}
		})
	}
}

func TestClientFetchNoContent(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return response(http.StatusNoContent, ""), nil
	})

	result, err := NewClient("http://accrual.test", &http.Client{Transport: transport}).Fetch(context.Background(), "12345678903")
	if err != nil {
		t.Fatal(err)
	}
	if !result.NoContent {
		t.Fatal("expected NoContent result")
	}
}

func TestClientFetchRateLimited(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		resp := response(http.StatusTooManyRequests, "")
		resp.Header.Set("Retry-After", "2")
		return resp, nil
	})

	result, err := NewClient("http://accrual.test", &http.Client{Transport: transport}).Fetch(context.Background(), "12345678903")
	if !errors.Is(err, ErrRateLimited) {
		t.Fatalf("err = %v, want ErrRateLimited", err)
	}
	if result.RetryAfter != 2*time.Second {
		t.Fatalf("RetryAfter = %v", result.RetryAfter)
	}
}

func TestClientFetchNoAddress(t *testing.T) {
	_, err := NewClient("", nil).Fetch(context.Background(), "12345678903")
	if !errors.Is(err, ErrNoAddress) {
		t.Fatalf("err = %v, want ErrNoAddress", err)
	}
}

func TestClientFetchUnexpectedStatus(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return response(http.StatusInternalServerError, ""), nil
	})

	_, err := NewClient("http://accrual.test", &http.Client{Transport: transport}).Fetch(context.Background(), "12345678903")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestClientFetchBadJSON(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusOK, `{`), nil
	})

	_, err := NewClient("http://accrual.test", &http.Client{Transport: transport}).Fetch(context.Background(), "12345678903")
	if err == nil {
		t.Fatal("expected error")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func jsonResponse(status int, body string) *http.Response {
	resp := response(status, body)
	resp.Header.Set("Content-Type", "application/json")
	return resp
}

func response(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
