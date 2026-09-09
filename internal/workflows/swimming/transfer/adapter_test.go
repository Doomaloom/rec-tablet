package transfer

import (
	"encoding/json"
	"os"

	"strings"
	"testing"
)

func TestAdapterPreservesExplicitIDs(t *testing.T) {
	data, e := os.ReadFile("../../../../examples/sister-roster.json")
	if e != nil {
		t.Fatal(e)
	}
	var raw []map[string]any
	if e = json.Unmarshal(data, &raw); e != nil {
		t.Fatal(e)
	}
	students := raw[0]["students"].([]any)
	students[0].(map[string]any)["phone"] = "exclude-me"
	raw[0]["students"] = append(students, map[string]any{"waitlist": true, "name": "Waiting"})
	data, _ = json.Marshal(raw)
	p, e := Adapt(data, "demo-autumn-2026", "Alex")
	if e != nil {
		t.Fatal(e)
	}
	if len(p.Courses[0].Swimmers) != 2 || p.Courses[0].Swimmers[0].SourceID != "demo-swimmer-001" {
		t.Fatal(p)
	}
	body, _ := json.Marshal(p)
	if strings.Contains(string(body), "phone") {
		t.Fatal("contact data leaked")
	}
	if _, e = DecodePackage(body); e != nil {
		t.Fatal(e)
	}
	delete(students[0].(map[string]any), "sourceId")
	data, _ = json.Marshal(raw)
	if _, e = Adapt(data, "demo", "Alex"); e == nil {
		t.Fatal("adapter invented ID")
	}
}
