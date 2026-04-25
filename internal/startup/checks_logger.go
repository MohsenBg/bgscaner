package startup

import (
	"bgscan/internal/logger"
	"os"
)

// checkLoggerHealth initializes all logging subsystems required by the
// application.
//
// It sequentially initializes:
//   - Core logger (system logs)
//   - UI logger (user-facing logs)
//   - Debug logger (development and diagnostic logs)
//
// If any logger fails to initialize, the program exits immediately since
// logging is considered a critical dependency for the application.
func checkLoggerHealth() {
	info("[INFO] Initializing loggers...")

	if err := logger.InitCore(); err != nil {
		errMsg("Core logger initialization failed", err)
		pressEnterToContinue()
		os.Exit(1)
	}
	success("[SUCCESS] Core logger initialized")

	if err := logger.InitUI(); err != nil {
		errMsg("UI logger initialization failed", err)
		pressEnterToContinue()
		os.Exit(1)
	}
	success("[SUCCESS] UI logger initialized")

	if err := logger.InitDebug(); err != nil {
		errMsg("Debug logger initialization failed", err)
		pressEnterToContinue()
		os.Exit(1)
	}
	success("[SUCCESS] Debug logger initialized")

	success("[LOGGER] Health check completed successfully ✅")
}
