// @zp-project: zp
// @zp-path: internal/gate/arbiter.go
// Arbiter packaging gate — called before any ZIP is written (ADR-047 §3.1).
// If violations are found, packaging is blocked and no ZIP is produced.
package gate

import (
	"fmt"
	"os"

	arbiter "github.com/Harshmaury/Arbiter/api"
)

// CheckPackaging runs the Arbiter static rule set against projectDir.
// Returns nil if all rules pass.
// Returns an error with formatted violation output if any rule fails.
// Prints results to stderr so the user sees them before the process exits.
func CheckPackaging(projectDir string) error {
	report, err := arbiter.VerifyPackaging(projectDir)
	if err != nil {
		// Fail-open: if Arbiter itself errors (e.g. no nexus.yaml), warn and continue.
		fmt.Fprintf(os.Stderr, "arbiter: gate error (continuing): %v\n", err)
		return nil
	}
	if report.OK() {
		fmt.Fprintf(os.Stderr, "arbiter: ✓ %d rule(s) passed\n", len(report.Passed))
		return nil
	}
	fmt.Fprintf(os.Stderr, "\nArbiter gate — violations found:\n")
	fmt.Fprint(os.Stderr, arbiter.FormatReport(report))
	return fmt.Errorf("package blocked — %d violation(s). Resolve and re-run zp", len(report.Violations))
}
