package main

import (
	"bgscan/input"
	"bgscan/pinger"
	"bgscan/scanner"
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"
)

const (
	DefaultFlushInterval  = 1
	DefaultStatusInterval = 1000 * time.Millisecond
	DefaultIPsFile        = "ips.txt"
	DefaultOutputPrefix   = "results"
)

func main() {
	printBanner()
	printIntro()

	config := askConfiguration()
	clearTerminal()
	printBanner()

	if err := scanner.RunScannerMemorySafe(config); err != nil {
		fmt.Printf("Error during scan: %v\n", err)
	}
}

func askConfiguration() scanner.ScannerConfig {
	reader := bufio.NewReader(os.Stdin)

	// -------- select mode first --------
	mode := askPingMode(reader)

	// -------- recommended defaults by mode --------
	var (
		defaultThreads int
		defaultTimeout time.Duration
		defaultPort    int
	)

	switch mode {
	case pinger.PingICMP:
		defaultThreads = 30
		defaultTimeout = 1 * time.Second
		defaultPort = 0

	case pinger.PingHTTP:
		defaultThreads = 40
		defaultTimeout = 4 * time.Second
		defaultPort = 443

	default: // TCP
		defaultThreads = 200
		defaultTimeout = 1500 * time.Millisecond
		defaultPort = 443
	}

	// -------- host only for HTTP --------
	host := ""
	if mode == pinger.PingHTTP {
		host = input.AskStringDefault(
			reader,
			"Host header for HTTP mode (empty = no Host header): ",
			"",
		)
	}

	// -------- ICMP privilege check --------
	if mode == pinger.PingICMP {
		if !pinger.CanUseICMP() {
			fmt.Println()
			fmt.Println("────────────────────────────────────────────────────────")
			fmt.Println("ICMP scan is not available on this system.")
			fmt.Println()
			fmt.Println("Reason:")
			fmt.Println("ICMP requires creating raw network sockets.")
			fmt.Println("On most systems, this is only allowed for:")
			fmt.Println("  • root user, or")
			fmt.Println("  • binaries with CAP_NET_RAW capability")
			fmt.Println()
			fmt.Println("What this means:")
			fmt.Println("Your current user does not have permission to send ICMP packets,")
			fmt.Println("so a real 'ping' scan cannot be performed.")
			fmt.Println()
			fmt.Println("What bgscan will do:")
			fmt.Println("The scan will automatically fall back to TCP mode,")
			fmt.Println("which works without special privileges.")
			fmt.Println("────────────────────────────────────────────────────────")
			fmt.Println()

			mode = pinger.PingTCP
		}
	}

	return scanner.ScannerConfig{
		Mode:           mode,
		Host:           host,
		Threads:        input.AskIntDefault(reader, fmt.Sprintf("Threads [%d]: ", defaultThreads), defaultThreads),
		PrintSuccess:   input.AskYesNoDefault(reader, "Print After found valid ip (y/n)[y]", true),
		ShuffleIPs:     input.AskYesNoDefault(reader, "Shuffle IP list before scanning? (y/n) [y]: ", true),
		StopAfterFound: input.AskIntDefault(reader, "Stop scan after N working IPs (empty = scan all): ", 0),
		MaxIPsToTest:   input.AskIntDefault(reader, "Maximum IPs to test (empty = all): ", 0),
		Port:           input.AskIntDefault(reader, fmt.Sprintf("Port [%d]: ", defaultPort), defaultPort),
		Timeout:        input.AskDurationDefault(reader, fmt.Sprintf("Per-IP timeout in ms [%d]: ", defaultTimeout.Milliseconds()), defaultTimeout),
		FlushInterval:  input.AskIntDefault(reader, fmt.Sprintf("Flush results every N successes [%d]: ", DefaultFlushInterval), DefaultFlushInterval),
		StatusInterval: input.AskDurationDefault(reader, fmt.Sprintf("Status update interval in ms [%d]: ", DefaultStatusInterval.Milliseconds()), DefaultStatusInterval),
		Verbose:        input.AskYesNoDefault(reader, "Verbose mode? (y/n) [n]: ", false),
		IPsFile:        input.AskStringDefault(reader, fmt.Sprintf("Path to IPs file [%s]: ", DefaultIPsFile), DefaultIPsFile),
		OutputPrefix:   input.AskStringDefault(reader, fmt.Sprintf("Output file prefix [%s]: ", DefaultOutputPrefix), DefaultOutputPrefix),
	}
}

// askPingMode prompts the user to choose TCP, ICMP, or HTTP
func askPingMode(reader *bufio.Reader) pinger.PingMode {
	fmt.Println("Select ping mode:")
	fmt.Println("1) TCP   (recommended, no special permissions)")
	fmt.Println("2) ICMP  (classic ping, requires root or CAP_NET_RAW)")
	fmt.Println("3) HTTP  (checks web servers, Host header optional)")
	choice := input.AskIntDefault(reader, "Choice [1]: ", 1)

	switch choice {
	case 2:
		return pinger.PingICMP
	case 3:
		return pinger.PingHTTP
	default:
		return pinger.PingTCP
	}
}

func printBanner() {
	fmt.Println()
	fmt.Print("\033[1;36m")
	fmt.Println("██████╗  ██████╗      ███████╗ ██████╗ █████╗ ███╗   ██╗")
	fmt.Println("██╔══██╗██╔════╝      ██╔════╝██╔════╝██╔══██╗████╗  ██║")
	fmt.Println("██████╔╝██║  ███╗     ███████╗██║     ███████║██╔██╗ ██║")
	fmt.Println("██╔══██╗██║   ██║     ╚════██║██║     ██╔══██║██║╚██╗██║")
	fmt.Println("██████╔╝╚██████╔╝     ███████║╚██████╗██║  ██║██║ ╚████║")
	fmt.Println("╚═════╝  ╚═════╝      ╚══════╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═══╝")
	fmt.Print("\033[0m")
	fmt.Println("=========================================================")
}

func printIntro() {
	fmt.Println()
	fmt.Println("This tool scans a list of IP addresses using TCP, ICMP, or HTTP checks.")
	fmt.Println("You will now configure how the scan should run.")
	fmt.Println("Nothing starts automatically.")
	fmt.Println("After configuration is complete, scanning will begin.")
	fmt.Println()
}

func clearTerminal() {
	switch runtime.GOOS {
	case "windows":
		runCmd("cmd", "/c", "cls")
	default:
		runCmd("clear")
	}
}

func runCmd(name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

