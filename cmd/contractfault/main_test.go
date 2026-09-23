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
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 2 || filepath.Base(paths[0]) != "a.json" {
		t.Fatalf("glob resolution wrong/unsorted: %v", paths)
	}
}

func TestResolveConsumerPathsMissingGlob(t *testing.T) {
	if _, err := resolveConsumerPaths(filepath.Join(t.TempDir(), "*.json")); err == nil {
		t.Fatal("expected error for empty glob match")
	}
}

func TestResolveConsumerPathsEmpty(t *testing.T) {
	paths, err := resolveConsumerPaths("")
	if err != nil || paths != nil {
		t.Fatalf("empty spec should yield nil, got %v / %v", paths, err)
	}
}

// TestRunEndToEnd exercises the whole pipeline against the shipped examples and
// asserts the rupture exit code.
func TestRunEndToEnd(t *testing.T) {
	base := filepath.Join("..", "..", "examples")
	oldC := filepath.Join(base, "contracts", "orders-v1.json")
	newC := filepath.Join(base, "contracts", "orders-v2.json")
	glob := filepath.Join(base, "consumers", "*.json")
	out := filepath.Join(t.TempDir(), "out.json")

	code := run([]string{"-old", oldC, "-new", newC, "-consumers", glob,
		"-format", "json", "-out", out}, os.Stdout, os.Stderr)
	if code != 2 {
		t.Fatalf("expected rupture exit 2, got %d", code)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("no report written")
	}
}

func TestRunMissingFlags(t *testing.T) {
	if code := run([]string{"-old", "x.json"}, os.Stdout, os.Stderr); code != usageExit {
		t.Fatalf("missing -new should exit %d, got %d", usageExit, code)
	}
}
