package assess_test

import (
	"testing"

	"github.com/Formulary-Labs/impact/assess"
	"github.com/Formulary-Labs/substrate/artifact"
)

func TestRunNoCatalog(t *testing.T) {
	cat := &artifact.ControlCatalog{}
	rows := assess.Run(assess.Options{Catalog: cat})
	if len(rows) != 0 {
		t.Errorf("expected 0 rows for empty catalog, got %d", len(rows))
	}
}

func TestRunBasic(t *testing.T) {
	cat := &artifact.ControlCatalog{
		Controls: []artifact.Control{
			{
				Id:        "AC-1",
				Title:     "Access Control Policy",
				Group:     "access",
				Objective: "Ensure access is controlled.",
				AssessmentRequirements: []artifact.AssessmentRequirement{
					{Id: "AC-1.01", Text: "Policy must be documented."},
				},
			},
		},
		Groups: []artifact.Group{
			{Id: "access", Title: "Access Control"},
		},
	}

	rows := assess.Run(assess.Options{Catalog: cat})
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	r := rows[0]
	if r.ControlID != "AC-1" {
		t.Errorf("expected ControlID=AC-1, got %q", r.ControlID)
	}
	if r.Group != "Access Control" {
		t.Errorf("expected Group=Access Control, got %q", r.Group)
	}
	if r.Requirements != 1 {
		t.Errorf("expected 1 requirement, got %d", r.Requirements)
	}
	if r.Severity != "[DATA NEEDED: severity — link a RiskCatalog with --risk-catalog]" {
		t.Errorf("expected DATA NEEDED severity without risk catalog, got %q", r.Severity)
	}
}

func TestHeaders(t *testing.T) {
	h := assess.Headers()
	if len(h) != 8 {
		t.Errorf("expected 8 headers, got %d: %v", len(h), h)
	}
}

func TestToCSVRow(t *testing.T) {
	r := assess.Row{
		ControlID:    "AC-1",
		Title:        "Access Control",
		Group:        "Access",
		Objective:    "Control access.",
		Impact:       "Unauthorized access",
		Severity:     "High",
		RiskIDs:      "RISK-001",
		Requirements: 2,
	}
	row := assess.ToCSVRow(r)
	if len(row) != 8 {
		t.Errorf("expected 8 CSV columns, got %d", len(row))
	}
	if row[0] != "AC-1" {
		t.Errorf("expected row[0]=AC-1, got %q", row[0])
	}
	if row[7] != "2" {
		t.Errorf("expected row[7]=2, got %q", row[7])
	}
}
