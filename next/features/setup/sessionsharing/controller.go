package sessionsharing

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"rectablet/next/contracts"
)

// SessionSharing owns code entry and server settings dialogs.
type SessionSharing struct {
	Server       func() string
	OnJoin       func(string) error
	OnSaveServer func(string) error
	OnError      func(error)
}

func (c *SessionSharing) Row() contracts.Row {
	return contracts.Row{ID: "share", Label: "Import from Session Code"}
}
func (c *SessionSharing) SubmitCode(code string) error {
	if len(code) != 6 {
		return fmt.Errorf("enter the six digit instructor code")
	}
	return c.OnJoin(code)
}
func (c *SessionSharing) SaveServer(value string) error {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("enter a server URL beginning with https:// (localhost may use http://)")
	}
	if parsed.Scheme == "http" && strings.ToLower(parsed.Hostname()) != "localhost" && (net.ParseIP(parsed.Hostname()) == nil || !net.ParseIP(parsed.Hostname()).IsLoopback()) {
		return fmt.Errorf("use https for network servers; http is allowed only for localhost development")
	}
	return c.OnSaveServer(parsed.String())
}
