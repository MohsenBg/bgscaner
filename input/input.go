package input

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// AskIntDefault asks for an integer input.
// If the user presses Enter without typing anything, returns defaultVal.
func AskIntDefault(reader *bufio.Reader, question string, defaultVal int) int {
	for {
		fmt.Print(question)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			return defaultVal
		}

		val, err := strconv.Atoi(input)
		if err == nil {
			return val
		}
		fmt.Println("invalid number, try again")
	}
}

// AskYesNoDefault asks a yes/no question.
// If the user presses Enter without typing anything, returns defaultVal.
func AskYesNoDefault(reader *bufio.Reader, question string, defaultVal bool) bool {
	for {
		fmt.Print(question)
		input, _ := reader.ReadString('\n')
		input = strings.ToLower(strings.TrimSpace(input))

		if input == "" {
			return defaultVal
		}
		if input == "y" || input == "yes" {
			return true
		}
		if input == "n" || input == "no" {
			return false
		}
		fmt.Println("please enter y or n")
	}
}

// AskStringDefault asks for a string input.
// If the user presses Enter without typing anything, returns defaultVal.
func AskStringDefault(reader *bufio.Reader, question string, defaultVal string) string {
	fmt.Print(question)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultVal
	}
	return input
}

// AskDurationDefault asks for a duration in milliseconds.
// If the user presses Enter without typing anything, it returns defaultVal.
func AskDurationDefault(
	reader *bufio.Reader,
	question string,
	defaultVal time.Duration,
) time.Duration {

	for {
		fmt.Print(question)

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("failed to read input")
			continue
		}

		input = strings.TrimSpace(input)

		if input == "" {
			return defaultVal
		}

		milliseconds, err := strconv.Atoi(input)
		if err != nil || milliseconds < 0 {
			fmt.Println("invalid number, enter a non-negative integer")
			continue
		}

		return time.Duration(milliseconds) * time.Millisecond
	}
}
