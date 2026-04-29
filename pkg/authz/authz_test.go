package authz

import "testing"

func TestIDs(t *testing.T) {
	type myID string

	input := []myID{"a", "b", "c"}
	got := IDs(input)

	if len(got) != 3 {
		t.Fatalf("IDs() returned %d elements, want 3", len(got))
	}
	for i, want := range []ID{"a", "b", "c"} {
		if got[i] != want {
			t.Errorf("IDs()[%d] = %q, want %q", i, got[i], want)
		}
	}
}

func TestIDsEmpty(t *testing.T) {
	got := IDs([]string{})
	if len(got) != 0 {
		t.Errorf("IDs(empty) returned %d elements, want 0", len(got))
	}
}

func TestFromIDs(t *testing.T) {
	type myID string

	input := []ID{"x", "y", "z"}
	got := FromIDs[myID](input)

	if len(got) != 3 {
		t.Fatalf("FromIDs() returned %d elements, want 3", len(got))
	}
	for i, want := range []myID{"x", "y", "z"} {
		if got[i] != want {
			t.Errorf("FromIDs()[%d] = %q, want %q", i, got[i], want)
		}
	}
}

func TestFromIDsEmpty(t *testing.T) {
	got := FromIDs[string]([]ID{})
	if len(got) != 0 {
		t.Errorf("FromIDs(empty) returned %d elements, want 0", len(got))
	}
}
