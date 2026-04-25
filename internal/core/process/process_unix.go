//go:build !windows

package process

import (
	"os/exec"
	"syscall"
)

// setSysProcAttr configures the command so that it starts in its own process
// group. This allows cleanup functions (StopGracefully, Kill) to send signals to
// the entire group via a negative PID.
//
// On Unix systems:
//   - Setpgid=true: the child becomes the leader of a new process group
//   - This ensures that TERM/KILL signals affect all subprocesses spawned
//     by the target binary, not just the direct child process.
//
// Windows has its own implementation in process_windows.go.
func setSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}

// signalTerminate sends SIGTERM to the process group. Using a negative PID
// ensures that *every process in the child's group* receives the signal.
//
// This supports graceful shutdown of binaries that spawn helper processes,
// internal workers, or tunnels that must be terminated together.
func signalTerminate(cmd *exec.Cmd) {
	_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
}

// killProcess forcefully terminates the entire process group by sending SIGKILL.
// Like signalTerminate, it targets the process group using a negative PID.
//
// This is used when graceful termination fails or during forced cleanup.
func killProcess(cmd *exec.Cmd) {
	_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}
