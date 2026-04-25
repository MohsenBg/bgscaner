package startup

import "bgscan/internal/core/config"

// runConfigValidation normalizes all configuration sections and
// returns a ValidationReport describing applied changes or errors.
//
// Each configuration module is responsible for validating its own
// constraints and reporting adjustments through the shared report.
func runConfigValidation() *config.ValidationReport {
	rep := &config.ValidationReport{}

	config.GetGeneral().Normalize(rep)
	config.GetWriter().Normalize(rep)
	config.GetICMP().Normalize(rep)
	config.GetTCP().Normalize(rep)
	config.GetHTTP().Normalize(rep)
	config.GetXray().Normalize(rep)
	config.GetDNS().Normalize(rep)

	return rep
}

// checkConfigHealth initializes the configuration system,
// performs validation, and logs the resulting status.
//
// If initialization fails, validation is skipped.
func checkConfigHealth() {
	info("[INFO] Initializing configuration system...")

	if err := config.Init(); err != nil {
		errMsg("Failed to initialize configs", err)
		return
	}

	success("[SUCCESS] Config files loaded")

	info("[INFO] Validating configuration values...")

	rep := runConfigValidation()

	printValidationReport(rep)

	if rep.HasErrors() {
		warn("[CONFIG] Completed with validation errors ⚠")
	} else {
		success("[CONFIG] Completed successfully ✅")
	}
}

// printValidationReport logs configuration adjustments recorded
// during normalization.
//
// Only changes (not errors) are printed here. Error reporting
// is handled separately through ValidationReport.
func printValidationReport(rep *config.ValidationReport) {
	if len(rep.Changes) == 0 {
		return
	}

	warn("[WARNING] Configuration adjustments applied:")

	for _, ch := range rep.Changes {
		warnf(
			"  %s: %v → %v (%s)",
			ch.Field,
			ch.Old,
			ch.New,
			ch.Note,
		)
	}
}
