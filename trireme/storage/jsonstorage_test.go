package storage

import (
	"os"
	"testing"
)

var filename = "./testdata/storage.json"

func TestJSONStorage(t *testing.T) {

	// Interface Canary
	var _ Storer = &JSONStorage{}

	js, err := New(filename)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := js.Get("testdata", "test1"); err == nil {
		t.Errorf("Test data appears to be tainted by existing tests.")
	}

	if err := js.Set("testdata", "test1", "C0FF33"); err != nil {
		t.Errorf("Failed to set data: %s", err)
	}

	if val, err := js.Get("testdata", "test1"); err != nil {
		t.Errorf("Failed to get data: %s", err)
	} else if val != "C0FF33" {
		t.Errorf("Expected 'C0FF33', got '%s'", val)
	}

	if err := js.Remove("testdata", "test1"); err != nil {
		t.Errorf("Failed to remove data: %s", err)
	}

	if _, err := js.Get("testdata", "test1"); err == nil {
		t.Errorf("Test data was not actually deleted.")
	}

	if _, err := New(filename); err != nil {
		t.Errorf("Failed to re-open the new testing file.", err)
	}

	os.Remove(filename)
}
