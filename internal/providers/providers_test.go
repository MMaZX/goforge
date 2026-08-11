package providers

import (
	"strings"
	"testing"
)

func TestResolveCanonicalAndAliases(t *testing.T) {
	cases := map[string]string{
		"postgres":   "postgres",
		"Postgres":   "postgres",
		" postgres ": "postgres",
		"postgresql": "postgres",
		"pg":         "postgres",
		"mariadb":    "mariadb",
		"MariaDB":    "mariadb",
		"mysql":      "mariadb",
		"MYSQL":      "mariadb",
	}
	for input, want := range cases {
		d, err := Resolve(input)
		if err != nil {
			t.Errorf("Resolve(%q): unexpected error: %v", input, err)
			continue
		}
		if d.Driver != want {
			t.Errorf("Resolve(%q) = %q, want %q", input, d.Driver, want)
		}
	}
}

func TestResolveUnknownListsSupportedNames(t *testing.T) {
	_, err := Resolve("oracle")
	if err == nil {
		t.Fatal("expected error for unknown driver")
	}
	for _, name := range Names() {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("expected error to mention %q, got: %v", name, err)
		}
	}
}

func TestListAndNamesConsistent(t *testing.T) {
	list := List()
	names := Names()
	if len(list) != len(names) {
		t.Fatalf("List() has %d entries, Names() has %d", len(list), len(names))
	}
	for i, d := range list {
		if d.Driver != names[i] {
			t.Errorf("List()[%d].Driver = %q, Names()[%d] = %q", i, d.Driver, i, names[i])
		}
		if d.Label == "" {
			t.Errorf("Descriptor %q has empty Label", d.Driver)
		}
		if d.ExampleDSN == "" {
			t.Errorf("Descriptor %q has empty ExampleDSN", d.Driver)
		}
	}
}

func TestListReturnsACopy(t *testing.T) {
	list := List()
	list[0].Driver = "tampered"
	if registry[0].Driver == "tampered" {
		t.Fatal("List() must return a copy, not the internal registry slice")
	}
}
