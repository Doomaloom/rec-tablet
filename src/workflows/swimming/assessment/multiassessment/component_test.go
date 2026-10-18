package multiassessment

import (
	"reflect"
	"testing"

	"rectablet/src/workflows/swimming/domain"
)

func TestBalancedMultiGroups(t *testing.T) {
	for _, tc := range []struct {
		students, groups int
		sizes            []int
		label            string
	}{
		{10, 3, []int{4, 3, 3}, "1 group of 4, 2 groups of 3"},
		{8, 3, []int{3, 3, 2}, "2 groups of 3, 1 group of 2"},
		{8, 2, []int{4, 4}, "2 groups of 4"},
		{1, 1, []int{1}, "1 group of 1"},
		{0, 1, nil, ""}, {4, 5, nil, ""}, {4, 0, nil, ""},
	} {
		if got := BalancedSizes(tc.students, tc.groups); !reflect.DeepEqual(got, tc.sizes) {
			t.Fatalf("split %d/%d: %v", tc.students, tc.groups, got)
		}
		if got := SplitLabel(tc.students, tc.groups); got != tc.label {
			t.Fatalf("label: %q, want %q", got, tc.label)
		}
	}
	for students := 1; students <= 50; students++ {
		for groups := 1; groups <= students; groups++ {
			sizes := BalancedSizes(students, groups)
			total := 0
			for _, size := range sizes {
				total += size
				if size < 1 || sizes[0]-size > 1 {
					t.Fatal("unbalanced group", sizes)
				}
			}
			if total != students {
				t.Fatal("split lost students", sizes)
			}
		}
	}
}

func TestMultiAssignmentConstraintsAndReassignment(t *testing.T) {
	c := &Component{State: State{Session: &Session{
		Roster: []domain.Enrollment{{ID: "a"}, {ID: "b"}, {ID: "c"}},
		Groups: []Group{{Number: 1, Capacity: 2}, {Number: 2, Capacity: 1}},
	}}}
	for _, id := range []string{"a", "b", "c", "outsider"} {
		if err := c.ToggleMember(id); err != nil {
			t.Fatal(err)
		}
	}
	if !reflect.DeepEqual(c.State.Session.Groups[0].Members, []string{"a", "b"}) || c.Ready() {
		t.Fatal("group capacity not enforced")
	}
	c.State.Session.AssignmentGroup = 1
	_ = c.ToggleMember("a")
	if len(c.State.Session.Groups[1].Members) != 0 {
		t.Fatal("duplicate assignment accepted")
	}
	_ = c.ToggleMember("c")
	if !c.Ready() {
		t.Fatal("complete assignment rejected")
	}
	_ = c.ToggleMember("c")
	c.State.Session.AssignmentGroup = 0
	_ = c.ToggleMember("b")
	_ = c.ToggleMember("c")
	c.State.Session.AssignmentGroup = 1
	_ = c.ToggleMember("b")
	if !c.Ready() {
		t.Fatal("reassignment failed")
	}
	c.State.Session.Groups[1].Members = []string{"outsider"}
	if c.Ready() {
		t.Fatal("ineligible student accepted")
	}
	c.State.Session.Groups[1].Members = []string{"a"}
	if c.Ready() {
		t.Fatal("duplicate assignment accepted by final validation")
	}
}
