//go:build windows

package process

import (
	"os"
	"os/exec"
)

// setSysProcAttr configures Windows process attributes.
//
// On Windows, no special configuration is required for process‑group handling.
// Windows does not support Unix-style process groups (negative PIDs), so
// StopGracefully and Kill rely directly on Process.Signal and Process.Kill.
func setSysProcAttr(cmd *exec.Cmd) {
	// Nothing needed for Windows.
}

// signalTerminate attempts to gracefully stop the process by sending an
// os.Interrupt signal. This maps to a CTRL_BREAK_EVENT for console processes.
// If the target application supports interrupt‑based shutdown, it can exit
// cleanly during StopGracefully.
//
// Note: Unlike Unix SIGTERM, Windows interrupt delivery depends on whether
// the child process is attached to a console and how it handles events.
func signalTerminate(cmd *exec.Cmd) {
	_ = cmd.Process.Signal(os.Interrupt)
}

// killProcess forcefully terminates the process using Process.Kill().
//
// Windows does not provide a portable way to kill a whole process tree;
// callers should be aware that child processes spawned by the target binary
// may outlive their parent unless the binary handles cleanup itself.
func killProcess(cmd *exec.Cmd) {
	_ = cmd.Process.Kill()
}
