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
		return empty, fmt.Errorf("could not connect to the sharing server: %w", err)
	}
	defer response.Body.Close()
	data, err := io.ReadAll(io.LimitReader(response.Body, MaxPackageBytes+2048))
	if err != nil {
		return empty, err
	}
	if response.StatusCode != http.StatusOK {
		message := strings.TrimSpace(string(data))
		if message == "" || len(message) > 500 {
			message = "session code is invalid or sharing has ended"
		}
		return empty, errors.New(message)
	}
	if len(data) > MaxPackageBytes+2048 {
		return empty, errors.New("shared class package is too large")
	}
	var wrapper struct {
		Package json.RawMessage `json:"package"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return empty, fmt.Errorf("sharing server returned invalid JSON: %w", err)
	}
	if len(wrapper.Package) == 0 {
		return empty, errors.New("sharing server returned no class package")
	}
	return DecodePackage(wrapper.Package)
}
