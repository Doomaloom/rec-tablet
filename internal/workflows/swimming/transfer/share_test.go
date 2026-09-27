package transfer

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestFetchSharedPackage(t *testing.T) {
	client := &http.Client{Transport: roundTrip(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost || req.URL.Path != "/api/device-shares/redeem" {
			t.Fatalf("unexpected request: %s %s", req.Method, req.URL.Path)
		}
		var request struct {
			Code string `json:"code"`
		}
		if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		if request.Code != "001234" {
			t.Fatalf("code lost leading zeroes: %q", request.Code)
		}
		body := `{"package":{"schemaVersion":1,"kind":"rec-tablet-classes","datasetId":"demo","instructor":"Alex","courses":[]}}`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}
	pkg, err := fetchSharedPackage(context.Background(), "https://share.example", "001234", client)
	if err != nil {
		t.Fatal(err)
	}
	if pkg.DatasetID != "demo" || pkg.Instructor != "Alex" {
		t.Fatalf("wrong package returned: %#v", pkg)
	}
}

func TestFetchSharedPackageReportsServerErrorsAndRejectsBadCodes(t *testing.T) {
	client := &http.Client{Transport: roundTrip(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(strings.NewReader("shared device session not found\n")), Header: make(http.Header)}, nil
	})}
	if _, err := fetchSharedPackage(context.Background(), "https://share.example", "12345", client); err == nil || !strings.Contains(err.Error(), "six digit") {
		t.Fatalf("bad code error: %v", err)
	}
	if _, err := fetchSharedPackage(context.Background(), "https://share.example", "123456", client); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("server error: %v", err)
	}
}

func TestFetchSharedPackageRequiresTLSForRemoteServers(t *testing.T) {
	client := &http.Client{Transport: roundTrip(func(req *http.Request) (*http.Response, error) {
		t.Fatalf("remote HTTP server was contacted: %s", req.URL)
		return nil, nil
	})}
	if _, err := fetchSharedPackage(context.Background(), "http://remote.example", "123456", client); err == nil || !strings.Contains(err.Error(), "HTTPS") {
		t.Fatalf("remote HTTP server was accepted: %v", err)
	}
}

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }
