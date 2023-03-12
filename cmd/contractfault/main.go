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
	"sort"
	"strings"

	"github.com/seismic-labs/contractfault/internal/analyze"
	"github.com/seismic-labs/contractfault/internal/consumer"
	"github.com/seismic-labs/contractfault/internal/contract"
	"github.com/seismic-labs/contractfault/internal/impact"
	"github.com/seismic-labs/contractfault/internal/report"
)

const usageExit = 3

type options struct {
	oldPath          string
	newPath          string
	consumers        string
	format           string
	out              string
	failOnBehavioral bool
	quiet            bool
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run is the testable core: it parses args, executes the pipeline and returns
// the process exit code, writing report output to out and diagnostics to errw.
func run(args []string, out, errw *os.File) int {
	opts, err := parseFlags(args, errw)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(errw, "contractfault: %v\n", err)
		return usageExit
	}

	oldC, err := contract.Load(opts.oldPath)
	if err != nil {
		fmt.Fprintf(errw, "contractfault: %v\n", err)
		return usageExit
	}
	newC, err := contract.Load(opts.newPath)
	if err != nil {
		fmt.Fprintf(errw, "contractfault: %v\n", err)
		return usageExit
	}

	paths, err := resolveConsumerPaths(opts.consumers)
	if err != nil {
		fmt.Fprintf(errw, "contractfault: %v\n", err)
		return usageExit
	}
	consumers, err := consumer.LoadAll(paths)
	if err != nil {
		fmt.Fprintf(errw, "contractfault: %v\n", err)
		return usageExit
	}

	diff, err := analyze.Compare(oldC, newC)
	if err != nil {
		fmt.Fprintf(errw, "contractfault: %v\n", err)
		return usageExit
	}
	rep := impact.Build(diff, consumers)

	var rendered []byte
	switch opts.format {
	case "json":
		rendered, err = report.JSON(rep)
		if err != nil {
			fmt.Fprintf(errw, "contractfault: %v\n", err)
			return usageExit
		}
	case "text":
		rendered = []byte(report.Text(rep))
	default:
		fmt.Fprintf(errw, "contractfault: unknown format %q (want text|json)\n", opts.format)
		return usageExit
	}

	if opts.quiet {
		line := fmt.Sprintf("%s %s->%s magnitude=%.1f verdict=%s breaking=%d\n",
			rep.Service, rep.FromVersion, rep.ToVersion,
			rep.Summary.Magnitude, rep.Summary.Verdict, rep.Summary.Breaking)
		if opts.format == "json" {
			// In quiet+json mode still emit full JSON to the file/stdout so the
			// viewer has data, but keep stderr terse.
			fmt.Fprint(errw, line)
		} else {
			rendered = []byte(line)
		}
	}

	if opts.out != "" {
		if err := os.WriteFile(opts.out, rendered, 0o644); err != nil {
			fmt.Fprintf(errw, "contractfault: writing %s: %v\n", opts.out, err)
			return usageExit
		}
	} else {
		out.Write(rendered)
	}

	return report.ExitCode(rep, opts.failOnBehavioral)
}

func parseFlags(args []string, errw *os.File) (*options, error) {
	fs := flag.NewFlagSet("contractfault", flag.ContinueOnError)
	fs.SetOutput(errw)
	opts := &options{}
	fs.StringVar(&opts.oldPath, "old", "", "path to the previous contract version (required)")
	fs.StringVar(&opts.newPath, "new", "", "path to the new contract version (required)")
	fs.StringVar(&opts.consumers, "consumers", "", "glob or comma-separated list of consumer manifest paths")
	fs.StringVar(&opts.format, "format", "text", "output format: text|json")
	fs.StringVar(&opts.out, "out", "", "write report to file instead of stdout")
	fs.BoolVar(&opts.failOnBehavioral, "fail-on-behavioral", false, "treat behavioral-only shifts as a failure (exit 2)")
	fs.BoolVar(&opts.quiet, "quiet", false, "print only the verdict summary line")
	fs.Usage = func() {
		fmt.Fprintf(errw, "contractfault: API contract consumer-impact analyzer\n\n")
		fmt.Fprintf(errw, "usage: contractfault -old before.json -new after.json -consumers 'consumers/*.json'\n\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return nil, err
