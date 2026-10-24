package setup

import (
	"errors"
	"testing"

	"rectablet/next/features/setup/sessionsharing"
	"rectablet/next/navigation"
)

func TestSessionSharingValidatesBeforeCallingParent(t *testing.T) {
	var joined, saved string
	share := sessionsharing.SessionSharing{
		OnJoin:       func(code string) error { joined = code; return nil },
		OnSaveServer: func(server string) error { saved = server; return nil },
	}
	for _, code := range []string{"", "00123", "0012345"} {
		if err := share.SubmitCode(code); err == nil || joined != "" {
			t.Fatalf("invalid code %q reached download callback", code)
		}
	}
	if err := share.SubmitCode("001234"); err != nil || joined != "001234" {
		t.Fatalf("leading-zero code: received %q, error %v", joined, err)
	}
	for _, server := range []string{"", "ftp://example.com", "http://example.com", "https://user@example.com", "https://example.com?key=value", "https://example.com#fragment"} {
		if err := share.SaveServer(server); err == nil || saved != "" {
			t.Fatalf("invalid server %q reached settings callback", server)
		}
	}
	for _, server := range []string{"https://example.com/", "http://localhost:8080", "http://127.0.0.1:8080", "http://[::1]:8080"} {
		if err := share.SaveServer(" " + server + " "); err != nil || saved != server {
			t.Fatalf("server %q: received %q, error %v", server, saved, err)
		}
	}
	failure := errors.New("settings write failed")
	share.OnSaveServer = func(string) error { return failure }
	if err := share.SaveServer("https://example.com"); !errors.Is(err, failure) {
		t.Fatalf("settings failure = %v", err)
	}
}

func TestSetupBindingsUseConfiguredCallbacksAndPreserveErrors(t *testing.T) {
	nav := navigation.New()
	component := New(nav, &transferStore{})
	nav.Features[navigation.Maintenance] = component
	nav.Push(navigation.Maintenance, nil)
	if err := nav.Refresh(); err != nil {
		t.Fatal(err)
	}
	// Device callbacks arrive after the component is constructed.
	failure := errors.New("settings write failed")
	component.ShareServer = "https://old.example"
	component.SaveShareServer = func(string) error { return failure }
	if err := component.sessionSharing.SaveServer("https://new.example/"); !errors.Is(err, failure) || component.ShareServer != "https://old.example" {
		t.Fatalf("failed settings write changed Server: %q, error %v", component.ShareServer, err)
	}
	component.SaveShareServer = func(string) error { return nil }
	if err := component.sessionSharing.SaveServer("https://new.example/"); err != nil || component.sessionSharing.Server() != "https://new.example" {
		t.Fatalf("server was not updated: %q, error %v", component.ShareServer, err)
	}
	state, err := component.Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Rows) != 3 {
		t.Fatalf("unexpected setup controls: %+v", state.Rows)
	}
	for _, row := range state.Rows {
		if row.ID != "export" && row.ID != "server" && row.ID != "display" {
			t.Fatalf("login bypass exposed: %s", row.ID)
		}
	}
}
