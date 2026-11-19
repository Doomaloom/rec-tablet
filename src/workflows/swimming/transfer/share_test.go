package transfer

import (
	"context"
	"encoding/json"
	"errors"
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
	if _, err := fetchSharedPackage(context.Background(), "https://share.example", "123456", client); err == nil || err.Error() != CodeNotRecognisedMessage {
		t.Fatalf("server error: %v", err)
	}
}

// The keypad shows plain instructions, never the server's response body (triage T1).
func TestFetchSharedPackageMapsFailuresToInstructorMessages(t *testing.T) {
	cases := []struct {
		status int
		body   string
		want   string
	}{
		{http.StatusNotFound, "shared device session not found\n", CodeNotRecognisedMessage},
		{http.StatusForbidden, "Binding no longer authorized\n", CodeNotRecognisedMessage},
		{http.StatusTooManyRequests, "too many code attempts; try again shortly\n", TooManyTriesMessage},
		{http.StatusBadRequest, "Invalid request body\n", ShareFailedMessage},
		{http.StatusBadGateway, "<html>bad gateway</html>", ShareFailedMessage},
		{http.StatusOK, "<html>Wi-Fi sign in</html>", ShareFailedMessage},
	}
	for _, c := range cases {
		client := &http.Client{Transport: roundTrip(func(_ *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: c.status, Body: io.NopCloser(strings.NewReader(c.body)), Header: make(http.Header)}, nil
		})}
		_, err := fetchSharedPackage(context.Background(), "https://share.example", "123456", client)
		if err == nil || err.Error() != c.want {
			t.Errorf("status %d: got %v, want %q", c.status, err, c.want)
		}
	}
	offline := errors.New("dial tcp: no route to host")
	client := &http.Client{Transport: roundTrip(func(_ *http.Request) (*http.Response, error) { return nil, offline })}
	_, err := fetchSharedPackage(context.Background(), "https://share.example", "123456", client)
	if err == nil || err.Error() != NoConnectionMessage || !errors.Is(err, offline) {
		t.Errorf("offline: got %v", err)
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

func TestFetchSharedPackageSyncCredentials(t *testing.T) {
	const identity = `"sync":{"schemaVersion":1,"sessionId":"session","datasetId":"demo","instructorId":"instructor","courses":[]},`
	pkg := func(sync string) string {
		return `{` + sync + `"schemaVersion":1,"kind":"rec-tablet-classes","datasetId":"demo","instructor":"Alex","courses":[]}`
	}
	grant := func(session, dataset, instructor, upload, read string, version int) string {
		body, _ := json.Marshal(map[string]any{"schemaVersion": version, "sessionId": session, "datasetId": dataset, "instructorId": instructor, "uploadToken": upload, "readToken": read, "expiresAt": "2099-01-01T00:00:00Z"})
		return string(body)
	}
	fetch := func(body string) (ClassPackage, error) {
		client := &http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
		})}
		return fetchSharedPackage(context.Background(), "https://share.example/", "001234", client)
	}
	valid := grant("session", "demo", "instructor", "upload-secret", "read-secret", 1)
	bound, err := fetch(`{"package":` + pkg(identity) + `,"sync":` + valid + `}`)
	if err != nil {
		t.Fatal(err)
	}
	c := bound.Credentials
	if c == nil || bound.Sync == nil || c.UploadToken != "upload-secret" || c.ReadToken != "read-secret" || c.Server != "https://share.example" || c.InstructorID != "instructor" || c.SessionID != "session" {
		t.Fatalf("bound redemption credentials: %+v", c)
	}
	// Capabilities exist only in memory; a re-serialized package is import-only.
	saved, err := json.Marshal(bound)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(saved), "secret") || strings.Contains(string(saved), "Token") {
		t.Fatal("package serialization leaked capabilities")
	}
	// Identity metadata without a server-issued grant never authorizes sync.
	unbound, err := fetch(`{"package":` + pkg(identity) + `}`)
	if err != nil || unbound.Credentials != nil {
		t.Fatalf("package without a grant gained credentials: %v", err)
	}
	guest, err := fetch(`{"package":` + pkg("") + `}`)
	if err != nil || guest.Credentials != nil || guest.Sync != nil {
		t.Fatalf("guest package is not import-only: %v", err)
	}
	for name, wrapper := range map[string]string{
		"guest package":    `{"package":` + pkg("") + `,"sync":` + valid + `}`,
		"other session":    `{"package":` + pkg(identity) + `,"sync":` + grant("other", "demo", "instructor", "u", "r", 1) + `}`,
		"other dataset":    `{"package":` + pkg(identity) + `,"sync":` + grant("session", "other", "instructor", "u", "r", 1) + `}`,
		"other instructor": `{"package":` + pkg(identity) + `,"sync":` + grant("session", "demo", "other", "u", "r", 1) + `}`,
		"no upload token":  `{"package":` + pkg(identity) + `,"sync":` + grant("session", "demo", "instructor", "", "r", 1) + `}`,
		"no read token":    `{"package":` + pkg(identity) + `,"sync":` + grant("session", "demo", "instructor", "u", "", 1) + `}`,
		"future schema":    `{"package":` + pkg(identity) + `,"sync":` + grant("session", "demo", "instructor", "u", "r", 2) + `}`,
		"identity dataset": `{"package":` + pkg(strings.Replace(identity, `"datasetId":"demo"`, `"datasetId":"other"`, 1)) + `,"sync":` + valid + `}`,
	} {
		if got, err := fetch(wrapper); err == nil || got.Credentials != nil {
			t.Fatalf("%s: inconsistent grant accepted", name)
		}
	}
}
