package setup

import (
	"errors"
	"testing"

	"rectablet/src/workflows/swimming/navigation"
)

func TestSessionSharingValidatesBeforeCallingParent(t *testing.T) {
	var joined, saved string
	share := SessionSharing{
		onJoin:       func(code string) error { joined = code; return nil },
		onSaveServer: func(server string) error { saved = server; return nil },
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
	share.onSaveServer = func(string) error { return failure }
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
	var joined, switched string
	failure := errors.New("download failed")
	component.JoinShare = func(code string, done func(error)) { joined = code; done(failure) }
	component.SwitchProfile = func(key string) error { switched = key; return nil }
	if err := component.sessionSharing.SubmitCode("001234"); err != nil {
		t.Fatal(err)
	}
	if joined != "001234" || !nav.State.Error || nav.State.Notice != failure.Error() {
		t.Fatalf("download status = %+v, code %q", nav.State, joined)
	}
	if err := component.profilePicker.Select("profile:saved-session"); err != nil || switched != "saved-session" {
		t.Fatalf("profile callback: key %q, error %v", switched, err)
	}
	component.ShareServer = "https://old.example"
	component.SaveShareServer = func(string) error { return failure }
	if err := component.sessionSharing.SaveServer("https://new.example/"); !errors.Is(err, failure) || component.ShareServer != "https://old.example" {
		t.Fatalf("failed settings write changed server: %q, error %v", component.ShareServer, err)
	}
	component.SaveShareServer = func(string) error { return nil }
	if err := component.sessionSharing.SaveServer("https://new.example/"); err != nil || component.sessionSharing.server() != "https://new.example" {
		t.Fatalf("server was not updated: %q, error %v", component.ShareServer, err)
	}
	var cancelled bool
	component.CancelShare = func() { cancelled = true }
	state, err := component.Build()
	if err != nil {
		t.Fatal(err)
	}
	state.OnLeave()
	if !cancelled {
		t.Fatal("leaving setup did not cancel the session download")
	}
}
