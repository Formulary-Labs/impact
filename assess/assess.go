// Package assess implements the core computation for impact.
//
// impact reads a ControlCatalog and optional RiskCatalog and produces a
// deterministic impact assessment: one row per control, linking each control
// to its potential harms and any risk entries that reference it.
//
// Framework identity is derived from catalog metadata — no hardcoded framework
// logic. The tool is composable with any gemara ControlCatalog.
package assess

import (
	"strings"

	"github.com/Formulary-Labs/substrate/artifact"
)

// Row is one row of the impact assessment output.
type Row struct {
	ControlID    string
	Title        string
	Group        string
	Objective    string
	Impact       string // business/operational impact text from linked risk, or [DATA NEEDED]
	Severity     string // "Low", "Medium", "High", "Critical", or "[DATA NEEDED]"
	RiskIDs      string // comma-separated list of linked risk IDs
	Requirements int    // count of assessment requirements
}

// Options configures an impact assessment run.
type Options struct {
	Catalog     *artifact.ControlCatalog
	RiskCatalog *artifact.RiskCatalog // may be nil
}

// riskRef is an internal reference to a risk entry linked to a control.
type riskRef struct {
	id       string
	severity string
	impact   string
}

// Run computes the impact assessment rows from the provided options.
func Run(opts Options) []Row {
	// Build a group title lookup from the control catalog.
	groupTitles := make(map[string]string, len(opts.Catalog.Groups))
	for _, g := range opts.Catalog.Groups {
		groupTitles[g.Id] = g.Title
	}

	// Build a control→risks index from the RiskCatalog when available.
	// gemara Risk does not directly reference control IDs, but its Title and
	// Description often contain them. We index by exact ID match in title.
	controlRisks := make(map[string][]riskRef)
	if opts.RiskCatalog != nil {
		// Associate each risk with a control when the risk title or description
		// contains the control ID (e.g. "AC-1", "PR.AC-1"). Practitioners can
		// pre-filter the catalog to the controls they care about.
		for i := range opts.Catalog.Controls {
			c := &opts.Catalog.Controls[i]
			for _, r := range opts.RiskCatalog.Risks {
				if containsID(r.Title, c.Id) || containsID(r.Description, c.Id) {
					controlRisks[c.Id] = append(controlRisks[c.Id], riskRef{
						id:       r.Id,
						severity: r.Severity.String(),
						impact:   r.Impact,
					})
				}
			}
		}
	}

	rows := make([]Row, 0, len(opts.Catalog.Controls))
	for _, c := range opts.Catalog.Controls {
		groupTitle := groupTitles[c.Group]
		if groupTitle == "" {
			groupTitle = c.Group
		}

		risks := controlRisks[c.Id]
		severity := "[DATA NEEDED: severity — link a RiskCatalog with --risk-catalog]"
		impactText := "[DATA NEEDED: impact — describe the business or operational harm if this control fails]"
		var riskIDs []string
		if len(risks) > 0 {
			// Use the worst (highest) severity across linked risks.
			severity = worstSeverity(risks)
			for _, r := range risks {
				riskIDs = append(riskIDs, r.id)
				if impactText == "[DATA NEEDED: impact — describe the business or operational harm if this control fails]" && r.impact != "" {
					impactText = r.impact
				}
			}
		}

		rows = append(rows, Row{
			ControlID:    c.Id,
			Title:        c.Title,
			Group:        groupTitle,
			Objective:    truncate(c.Objective, 200),
			Impact:       impactText,
			Severity:     severity,
			RiskIDs:      strings.Join(riskIDs, "; "),
			Requirements: len(c.AssessmentRequirements),
		})
	}
	return rows
}

// Headers returns the CSV column headers for the impact assessment output.
func Headers() []string {
	return []string{
		"Control ID", "Title", "Category", "Objective",
		"Potential Impact", "Severity", "Linked Risk IDs", "Assessment Requirements",
	}
}

// ToCSVRow converts a Row to a string slice for CSV output.
func ToCSVRow(r Row) []string {
	return []string{
		r.ControlID, r.Title, r.Group, r.Objective,
		r.Impact, r.Severity, r.RiskIDs,
		itoa(r.Requirements),
	}
}

// worstSeverity returns the highest severity string from a list of risk refs.
func worstSeverity(risks []riskRef) string {
	order := map[string]int{
		"Critical": 3,
		"High":     2,
		"Medium":   1,
		"Low":      0,
		"Invalid":  -1,
	}
	worst := "Low"
	for _, r := range risks {
		if order[r.severity] > order[worst] {
			worst = r.severity
		}
	}
	return worst
}

// containsID returns true when haystack contains the given id as a word or token.
func containsID(haystack, id string) bool {
	if id == "" {
		return false
	}
	upper := strings.ToUpper(haystack)
	return strings.Contains(upper, strings.ToUpper(id))
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	result := make([]byte, 0, 10)
	for n > 0 {
		result = append([]byte{byte('0' + n%10)}, result...)
		n /= 10
	}
	return string(result)
}
