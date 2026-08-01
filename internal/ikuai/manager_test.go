package ikuai

import (
	"testing"
)

// newRegistry builds an empty Registry for testing (the package-global one is
// shared, so tests use a fresh local instance).
func newRegistry() *Registry {
	return &Registry{managers: map[string]*Manager{}}
}

// stubManager builds a Manager with nil client/api — enough to exercise the
// registry's map bookkeeping without spinning up real SDK clients.
func stubManager(name string) *Manager {
	return &Manager{name: name}
}

func TestRegistryAddGetDefault(t *testing.T) {
	r := newRegistry()
	if r.Default() != nil {
		t.Fatal("empty registry default should be nil")
	}
	if got := r.Get("nope"); got != nil {
		t.Errorf("Get on missing key = %v, want nil", got)
	}

	m1 := stubManager("office")
	r.Add("office", m1)
	if r.Default() == nil || r.Default().name != "office" {
		t.Errorf("first Add should become default, got %v", r.Default())
	}
	if got := r.Get("office"); got != m1 {
		t.Error("Get should return the added manager")
	}

	m2 := stubManager("home")
	r.Add("home", m2)
	// default stays the first one.
	if r.Default().name != "office" {
		t.Errorf("default changed after second Add: %v", r.Default())
	}
	if names := r.Names(); len(names) != 2 {
		t.Errorf("Names() = %v, want 2 entries", names)
	}
}

func TestRegistryRemoveReelectsDefault(t *testing.T) {
	r := newRegistry()
	r.Add("office", stubManager("office"))
	r.Add("home", stubManager("home"))
	if r.Default().name != "office" {
		t.Fatalf("default = %v, want office", r.Default())
	}

	r.Remove("office")
	if r.Default() == nil {
		t.Fatal("default nil after removing current default; should re-elect")
	}
	if r.Default().name != "home" {
		t.Errorf("re-elected default = %v, want home", r.Default())
	}
	if r.Get("office") != nil {
		t.Error("removed manager should be gone")
	}

	// Remove the last one.
	r.Remove("home")
	if r.Default() != nil {
		t.Errorf("default = %v after removing all, want nil", r.Default())
	}
}

func TestRegistryRemoveMissingIsNoop(t *testing.T) {
	r := newRegistry()
	r.Add("office", stubManager("office"))
	r.Remove("nonexistent") // must not panic / change state
	if r.Get("office") == nil {
		t.Error("removing a missing key should not affect existing entries")
	}
}
