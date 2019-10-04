package config

import (
	"reflect"
	"testing"

	"github.com/kr/pretty"
	"github.com/pokanop/nostromo/model"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		expErr bool
	}{
		{"invalid path", "", true},
		{"missing path", "/does/not/exist/.nostromo", true},
		{"bad file format", "../testdata/bad.nostromo", true},
		// {"good file format", "../testdata/good.nostromo", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c, err := Parse(test.path)
			if test.expErr && err == nil {
				t.Errorf("expected error but got none")
			} else if !test.expErr {
				if err != nil {
					t.Errorf("expected no error but got %s", err)
				} else if c.manifest == nil {
					t.Errorf("manifest is nil")
				}
			}
		})
	}
}

func TestSave(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		manifest *model.Manifest
		expErr   bool
	}{
		{"invalid path", "", nil, true},
		{"nil manifest", "path", nil, true},
		{"no perms", "/tmp/no-perms/.nostromo", fakeManifest(), true},
		{"valid manifest", "/tmp/test.nostromo", fakeManifest(), false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := NewConfig(test.path, test.manifest)
			err := c.Save()
			if test.expErr && err == nil {
				t.Errorf("expected error but got none")
			} else if !test.expErr && err != nil {
				t.Errorf("expected no error but got %s", err)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		manifest *model.Manifest
		expErr   bool
	}{
		{"invalid path", "", nil, true},
		{"missing path", "/does/not/exist", nil, true},
		{"valid path", "/tmp/test.nostromo", fakeManifest(), false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := NewConfig(test.path, test.manifest)
			if test.manifest != nil {
				err := c.Save()
				if err != nil {
					t.Errorf("unable to save temporary manifest: %s", err)
				}
			}
			err := c.Delete()
			if test.expErr && err == nil {
				t.Errorf("expected error but got none")
			} else if !test.expErr && err != nil {
				t.Errorf("expected no error but got %s", err)
			}
		})
	}
}

func TestExists(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"invalid path", "", false},
		{"missing path", "/does/not/exist", false},
		{"valid path", "../testdata/good.nostromo", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := NewConfig(test.path, nil)
			if actual := c.Exists(); actual != test.expected {
				t.Errorf("expected: %t, actual: %t", test.expected, actual)
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{"no key", "", "key not found"},
		{"missing key", "missing", "key not found"},
		{"verbose", "verbose", "true"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := NewConfig("path", fakeManifest())
			c.Manifest().Config.Verbose = true
			if actual := c.Get(test.key); actual != test.expected {
				t.Errorf("expected: %s, actual: %s", test.expected, actual)
			}
		})
	}
}

func TestSet(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		expErr   bool
		expected string
	}{
		{"no key", "", "", true, ""},
		{"missing key", "missing", "", true, ""},
		{"verbose empty", "verbose", "", true, ""},
		{"verbose true", "verbose", "true", false, "true"},
		{"verbose false", "verbose", "false", false, "false"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := NewConfig("path", fakeManifest())
			err := c.Set(test.key, test.value)
			if test.expErr && err == nil {
				t.Errorf("expected error but got none")
			} else if !test.expErr && err != nil {
				t.Errorf("expected no error but got %s", err)
			} else if !test.expErr {
				if actual := c.Get(test.key); actual != test.expected {
					t.Errorf("expected: %s, actual: %s", test.expected, actual)
				}
			}
		})
	}
}

func TestKeys(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected []string
	}{
		{"keys", NewConfig("path", fakeManifest()), []string{"verbose"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := test.config.Manifest().Config.Keys(); !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("expected: %s, actual: %s", test.expected, actual)
			}
		})
	}
}

func TestFields(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected map[string]interface{}
	}{
		{
			"keys",
			NewConfig("path", fakeManifest()),
			map[string]interface{}{
				"verbose": false,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := test.config.Manifest().Config.Fields(); !reflect.DeepEqual(actual, test.expected) {
				pretty.Println(actual)
				t.Errorf("expected: %s, actual: %s", test.expected, actual)
			}
		})
	}
}

func fakeManifest() *model.Manifest {
	m := model.NewManifest()
	m.AddCommand("one.two.three", "command", "")
	m.AddSubstitution("one.two", "name", "alias")
	return m
}
