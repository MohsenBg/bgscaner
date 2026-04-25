package startup

import (
	"bgscan/internal/core/dns"
	"bgscan/internal/core/process"
)

// checkDNSTTHealth verifies that the DNSTT client binary is present,
// executable, and behaves as expected.
//
// It performs three steps:
//  1. Locate the dnstt-client binary.
//  2. Ensure the binary is marked executable (where applicable).
//  3. Run a lightweight verification via dns.VerifyDNSTTClient.
func checkDNSTTHealth() {
	info("[INFO] Finding DNSTT client...")

	path, err := dns.FindDNSTTClient()
	if err != nil {
		binaryMissing("DNSTT", "dnstt-client")
		errMsg("Binary lookup error", err)
		return
	}

	successf("[SUCCESS] DNSTT found at: %s", path)

	info("[INFO] Ensuring DNSTT client binary is executable...")
	if err := process.EnsureExecutable(path); err != nil {
		errMsg("Failed to set executable bit for DNSTT client", err)
		return
	}

	success("[SUCCESS] DNSTT binary is executable")

	info("[INFO] Verifying DNSTT client...")

	if err := dns.VerifyDNSTTClient(); err != nil {
		errMsg("DNSTT client validation failed", err)
		return
	}

	success("[DNSTT] Health check completed successfully ✅")
}

// checkSlipstreamHealth verifies that the Slipstream client binary is
// present, executable, and responsive.
//
// It performs three steps:
//  1. Locate the slipstream-client binary.
//  2. Ensure the binary is marked executable (where applicable).
//  3. Run a lightweight verification via dns.VerifySlipstreamClient.
func checkSlipstreamHealth() {
	info("[INFO] Finding Slipstream client...")

	path, err := dns.FindSlipstreamClient()
	if err != nil {
		binaryMissing("Slipstream", "slipstream-client")
		errMsg("Binary lookup error", err)
		return
	}

	successf("[SUCCESS] Slipstream found at: %s", path)

	info("[INFO] Ensuring Slipstream client binary is executable...")
	if err := process.EnsureExecutable(path); err != nil {
		errMsg("Failed to set executable bit for Slipstream client", err)
		return
	}

	success("[SUCCESS] Slipstream binary is executable")

	info("[INFO] Verifying Slipstream client...")

	if err := dns.VerifySlipstreamClient(); err != nil {
		errMsg("Slipstream client validation failed", err)
		return
	}

	success("[Slipstream] Health check completed successfully ✅")
}
