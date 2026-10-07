package common

import (
	"os"
	"testing"
)

func TestRingBuffer(t *testing.T) {
	rb := NewRingBuffer[int](3)

	rb.Push(1)
	rb.Push(2)
	rb.Push(3)
	if rb.Len() != 3 {
		t.Fatalf("expected 3 items, got %d", rb.Len())
	}

	rb.Push(4) // evicts 1
	if rb.Len() != 3 {
		t.Fatalf("expected 3 items after eviction, got %d", rb.Len())
	}

	latest := rb.List(2)
	if len(latest) != 2 || latest[0] != 4 || latest[1] != 3 {
		t.Fatalf("unexpected list items: %v", latest)
	}

	all := rb.All()
	if len(all) != 3 || all[0] != 2 || all[1] != 3 || all[2] != 4 {
		t.Fatalf("unexpected all items: %v", all)
	}
}

func TestLoadSeedFile(t *testing.T) {
	// Create temporary json file
	tmpFile, err := os.CreateTemp("", "seed-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	_, _ = tmpFile.WriteString(`{"name": "ai-meter", "version": 1}`)
	_ = tmpFile.Close()

	var data struct {
		Name    string `json:"name"`
		Version int    `json:"version"`
	}

	if err := LoadSeedFile(tmpFile.Name(), &data); err != nil {
		t.Fatalf("failed to load seed file: %v", err)
	}
	if data.Name != "ai-meter" || data.Version != 1 {
		t.Fatalf("unexpected seed content: %+v", data)
	}
}
