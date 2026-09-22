// impact produces a framework-agnostic impact assessment CSV from a gemara
// ControlCatalog and optional RiskCatalog.
//
// Usage:
//
//	impact --catalog <path> [--risk-catalog <path>] [--output <path>] [--format json|csv]
package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Formulary-Labs/impact/assess"
	"github.com/Formulary-Labs/substrate/artifact"
	"github.com/Formulary-Labs/substrate/exit"
	"github.com/Formulary-Labs/substrate/provenance"
)

const version = "0.1.0"

func main() {
	var (
		catalogFlag     = flag.String("catalog", "", "Path to gemara ControlCatalog YAML/JSON (required)")
		riskCatalogFlag = flag.String("risk-catalog", "", "Path to gemara RiskCatalog YAML/JSON (optional)")
		outputFlag      = flag.String("output", "impact-assessment.csv", "Output file path (use - for stdout)")
		formatFlag      = flag.String("format", "csv", "Output format: csv, json")
		programFlag     = flag.String("program", "", "Program slug (for provenance)")
		dryRunFlag      = flag.Bool("dry-run", false, "Print row count without writing output file")
		versionFlag     = flag.Bool("version", false, "Print version and exit")
	)
	flag.Usage = usage
	flag.Parse()

	if *versionFlag {
		fmt.Printf("impact version %s\n", version)
		os.Exit(exit.OK)
	}

	if *catalogFlag == "" {
		fmt.Fprintln(os.Stderr, `{"error": "--catalog is required", "code": 2}`)
		flag.Usage()
		os.Exit(exit.ToolError)
	}

	cat, err := artifact.LoadControlCatalog(*catalogFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, `{"error": %q, "code": 2}`+"\n", err.Error())
		os.Exit(exit.ToolError)
	}

	opts := assess.Options{Catalog: cat}

	if *riskCatalogFlag != "" {
		rc, err := artifact.LoadRiskCatalog(*riskCatalogFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, `{"error": %q, "code": 2}`+"\n", err.Error())
			os.Exit(exit.ToolError)
		}
		opts.RiskCatalog = rc
	}

	rows := assess.Run(opts)

	if *dryRunFlag {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(map[string]interface{}{ //nolint:errcheck
			"dry_run":   true,
			"rows":      len(rows),
			"catalog":   cat.Metadata.Id,
			"framework": frameworkFromCatalog(cat),
		})
		os.Exit(exit.OK)
	}

	switch *formatFlag {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(map[string]interface{}{
			"catalog":   cat.Metadata.Id,
			"framework": frameworkFromCatalog(cat),
			"rows":      rows,
		}); err != nil {
			fmt.Fprintf(os.Stderr, `{"error": %q, "code": 2}`+"\n", err.Error())
			os.Exit(exit.ToolError)
		}
	default: // csv
		if err := writeCSV(*outputFlag, rows); err != nil {
			fmt.Fprintf(os.Stderr, `{"error": %q, "code": 2}`+"\n", err.Error())
			os.Exit(exit.ToolError)
		}
		fmt.Fprintf(os.Stderr, "wrote %s (%d rows)\n", *outputFlag, len(rows))
	}

	_ = provenance.Write("logs/provenance.jsonl", provenance.Entry{
		Spec:        "functions/impact-spec.md",
		Output:      *outputFlag,
		OutputType:  "other",
		Program:     *programFlag,
		Purpose:     fmt.Sprintf("impact: %d controls assessed from %s", len(rows), filepath.Base(*catalogFlag)),
		Reusability: provenance.Instance,
		QualityGate: provenance.Pass,
		Tool:        "impact",
		ToolVersion: version,
	})
}

func writeCSV(path string, rows []assess.Row) error {
	var w *csv.Writer
	if path == "-" {
		w = csv.NewWriter(os.Stdout)
	} else {
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil && filepath.Dir(path) != "." {
			return fmt.Errorf("creating output directory: %w", err)
		}
		f, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("creating output file: %w", err)
		}
		defer f.Close() //nolint:errcheck
		w = csv.NewWriter(f)
	}
	if err := w.Write(assess.Headers()); err != nil {
		return err
	}
	for _, r := range rows {
		if err := w.Write(assess.ToCSVRow(r)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

// frameworkFromCatalog extracts a human-readable framework identifier from catalog
// metadata. Falls back to the catalog ID when description is not set.
func frameworkFromCatalog(cat *artifact.ControlCatalog) string {
	if cat.Metadata.Description != "" {
		return cat.Metadata.Description
	}
	if cat.Metadata.Id != "" {
		return cat.Metadata.Id
	}
	return cat.Title
}

func usage() {
	fmt.Fprintln(os.Stderr, `impact — framework-agnostic impact assessment from a ControlCatalog

Usage:
  impact --catalog <path> [flags]

Flags:
  --catalog string        Path to gemara ControlCatalog YAML/JSON (required)
  --risk-catalog string   Path to gemara RiskCatalog YAML/JSON (optional; links risk severity to controls)
  --output string         Output CSV path (default: impact-assessment.csv; use - for stdout)
  --format string         Output format: csv (default), json
  --program string        Program slug for provenance (optional)
  --dry-run               Print row count without writing output file
  --version               Print version and exit

Framework identity is derived from catalog metadata — no framework logic is hardcoded.
impact works with any gemara ControlCatalog regardless of standard.

Examples:
  impact --catalog data/catalogs/iso42001.yaml
  impact --catalog data/catalogs/iso42001.yaml --risk-catalog data/risks/org-risks.yaml
  impact --catalog data/catalogs/iso42001.yaml --output reports/impact-assessment.csv
  impact --catalog data/catalogs/iso42001.yaml --format json`)
}
