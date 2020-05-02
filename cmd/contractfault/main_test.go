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
