// Command contractfault is a deterministic API contract consumer-impact
// analyzer. It ingests two versions of a documented JSON contract plus a set
// of consumer usage manifests, classifies every difference as breaking,
// additive or behavioral, joins those changes against the consumers that
// depend on the affected elements, and emits a JSON or text seismic report
// with a CI-friendly exit code.
//
// Usage:
//
//	contractfault -old before.json -new after.json -consumers 'consumers/*.json' [flags]
//
// Flags:
//
//	-old            path to the previous contract version (required)
//	-new            path to the new contract version (required)
//	-consumers      glob or comma list of consumer manifest paths
//	-format         "text" (default) or "json"
//	-out            write the report to a file instead of stdout
//	-fail-on-behavioral   exit non-zero on behavioral-only shifts
//	-quiet          suppress the report body, print only the verdict line
//
// Exit codes: 0 stable, 1 shaken (behavioral), 2 rupture (breaking), 3 usage
// or IO error.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
