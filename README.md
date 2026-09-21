# impact

Framework-agnostic impact assessment from a gemara `ControlCatalog`.

## Overview

`impact` reads a gemara `ControlCatalog` and produces `impact-assessment.csv`: one row per control, mapping each control to its potential harms, severity, and linked risk IDs. An optional `RiskCatalog` connects controls to documented organizational risks.

Framework identity is derived from catalog metadata — no framework logic is hardcoded. `impact` works with any gemara `ControlCatalog` regardless of which compliance standard it covers.

## Usage

```
impact --catalog <path> [flags]

Flags:
  --catalog string        Path to gemara ControlCatalog YAML/JSON (required)
  --risk-catalog string   Path to gemara RiskCatalog YAML/JSON (optional)
  --output string         Output CSV path (default: impact-assessment.csv)
  --format string         Output format: csv (default), json
  --program string        Program slug for provenance
  --dry-run               Print row count without writing output
  --version               Print version and exit
```

## Output

`impact-assessment.csv` with columns:

| Column | Source |
|---|---|
| Control ID | ControlCatalog |
| Title | ControlCatalog |
| Category | ControlCatalog group |
| Objective | ControlCatalog |
| Potential Impact | RiskCatalog (or `[DATA NEEDED]`) |
| Severity | RiskCatalog (or `[DATA NEEDED]`) |
| Linked Risk IDs | RiskCatalog |
| Assessment Requirements | ControlCatalog |

Cells marked `[DATA NEEDED]` require AI agent judgment to complete — they are explicit handoffs to the regimen layer.

## Install

```sh
go install github.com/Formulary-Labs/impact/cmd/impact@latest
```

## Part of Formulary

`impact` is part of the [Formulary](https://github.com/Formulary-Labs) compliance micro-tool ecosystem. It is composable: use it alone, or combine it with `specimen`, `appraise`, and `formula` for a complete program artifact pipeline.
