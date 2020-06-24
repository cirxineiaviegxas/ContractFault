package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestResolveConsumerPathsGlob verifies glob expansion, comma lists and
// literal-path fallback all resolve deterministically.
func TestResolveConsumerPathsGlob(t *testing.T) {
	dir := t.TempDir()
	for _, n := range []string{"b.json", "a.json"} {
		if err := os.WriteFile(filepath.Join(dir, n), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	paths, err := resolveConsumerPaths(filepath.Join(dir, "*.json"))
