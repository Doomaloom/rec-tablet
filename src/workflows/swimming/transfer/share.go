package transfer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const DefaultShareServer = "https://decksupervisor.fly.dev"

// Messages shown on the login keypad when a code cannot be redeemed. They fit
// the keypad's notice line (two lines at 800x480) and say what to do next.
const (
	CodeNotRecognisedMessage = "Code not recognised, or sharing has ended. Ask the deck supervisor for a new code."
	TooManyTriesMessage      = "Too many tries. Wait a minute, then try again."
	NoConnectionMessage      = "Cannot reach the server. Check the Wi-Fi connection, then try again."
	ShareFailedMessage       = "Could not load your classes. Try again, or ask the deck supervisor for help."
)

// ShareError carries an instructor-facing message and keeps the cause for errors.Is.
type ShareError struct {
	Message string
	Cause   error
}

func (e *ShareError) Error() string { return e.Message }
func (e *ShareError) Unwrap() error { return e.Cause }

// redeemError maps the redeem endpoint's HTTP status; the body is not shown.
func redeemError(status int, body string) error {
	cause := fmt.Errorf("redeem failed with status %d: %.200s", status, strings.TrimSpace(body))
	switch status {
	case http.StatusNotFound, http.StatusForbidden, http.StatusGone:
		return &ShareError{CodeNotRecognisedMessage, cause}
	case http.StatusTooManyRequests:
		return &ShareError{TooManyTriesMessage, cause}
	default:
		return &ShareError{ShareFailedMessage, cause}
	}
}

func FetchSharedPackage(ctx context.Context, server, code string) (ClassPackage, error) {
	return fetchSharedPackage(ctx, server, code, &http.Client{Timeout: 25 * time.Second})
}

func fetchSharedPackage(ctx context.Context, server, code string, client *http.Client) (ClassPackage, error) {
	var empty ClassPackage
	server = strings.TrimRight(strings.TrimSpace(server), "/")
	if server == "" {
		server = DefaultShareServer
	}
	if len(code) != 6 {
		return empty, errors.New("enter a six digit instructor code")
	}
	for _, digit := range code {
		if digit < '0' || digit > '9' {
			return empty, errors.New("enter a six digit instructor code")
		}
	}
	parsedServer, err := url.Parse(server)
	if err != nil || parsedServer.Host == "" || parsedServer.User != nil || parsedServer.RawQuery != "" || parsedServer.Fragment != "" || (parsedServer.Scheme != "https" && parsedServer.Scheme != "http") {
		return empty, errors.New("share server must be a valid HTTPS URL")
	}
	if parsedServer.Scheme == "http" && strings.ToLower(parsedServer.Hostname()) != "localhost" && (net.ParseIP(parsedServer.Hostname()) == nil || !net.ParseIP(parsedServer.Hostname()).IsLoopback()) {
		return empty, errors.New("share servers must use HTTPS except for localhost development")
	}
	body, err := json.Marshal(map[string]string{"code": code})
	if err != nil {
		return empty, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, server+"/api/device-shares/redeem", bytes.NewReader(body))
	if err != nil {
		return empty, err
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := client.Do(req)
	if err != nil {
		return empty, &ShareError{NoConnectionMessage, err}
	}
	defer response.Body.Close()
	data, err := io.ReadAll(io.LimitReader(response.Body, MaxPackageBytes+2048))
	if err != nil {
		return empty, err
	}
	if response.StatusCode != http.StatusOK {
		return empty, redeemError(response.StatusCode, string(data))
	}
	if len(data) > MaxPackageBytes+2048 {
		return empty, errors.New("shared class package is too large")
	}
	var wrapper struct {
		Package json.RawMessage  `json:"package"`
		Sync    *SyncCredentials `json:"sync"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		// A Wi-Fi sign-in page answering in place of the server lands here.
		return empty, &ShareError{ShareFailedMessage, fmt.Errorf("sharing server returned invalid JSON: %w", err)}
	}
	if len(wrapper.Package) == 0 {
		return empty, &ShareError{ShareFailedMessage, errors.New("sharing server returned no class package")}
	}
	pkg, err := DecodePackage(wrapper.Package)
	if err != nil {
		return empty, err
	}
	if c := wrapper.Sync; c != nil {
		if pkg.Sync == nil || c.SchemaVersion != 1 || pkg.Sync.SchemaVersion != 1 || c.SessionID != pkg.Sync.SessionID || c.DatasetID != pkg.DatasetID || c.DatasetID != pkg.Sync.DatasetID || c.InstructorID != pkg.Sync.InstructorID || c.UploadToken == "" || c.ReadToken == "" {
			return empty, errors.New("invalid results sync credentials")
		}
		c.Server = server
		pkg.Credentials = c
	}
	return pkg, nil
}
