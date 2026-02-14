package scanner

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/fatih/color"
)

var SKIP_PROGRESS_BAR_REPLACE atomic.Bool

func StartProgressTracker(
	scanned *uint64,
	found *uint64,
	total int,
	interval time.Duration,
) chan<- struct{} {

	done := make(chan struct{})
	start := time.Now()
	barWidth := 40

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s := int(atomic.LoadUint64(scanned))
				f := atomic.LoadUint64(found)

				if s > total {
					s = total
				}

				elapsed := time.Since(start).Truncate(time.Second)

				percent := float64(s) / float64(total)
				filled := int(percent * float64(barWidth))
				empty := barWidth - filled

				bar :=
					color.GreenString(strings.Repeat("█", filled)) +
						color.HiBlackString(strings.Repeat("░", empty))

				var eta string
				if percent > 0 {
					remaining := time.Duration(float64(elapsed) * (1 - percent) / percent)
					eta = remaining.Truncate(time.Second).String()
				} else {
					eta = "?"
				}

				if !SKIP_PROGRESS_BAR_REPLACE.Load() {
					// Move cursor up 2 lines and clear
					fmt.Print("\033[2A\033[J")
				}
				SKIP_PROGRESS_BAR_REPLACE.Store(false)

				// Line 1: stats
				fmt.Printf(
					"%s | %s | %s | %s\n",
					color.YellowString("scanned:%d", s),
					color.CyanString("left:%d", total-s),
					color.GreenString("found:%d", f),
					color.MagentaString("elapsed:%v", elapsed),
				)

				// Line 2: bar
				fmt.Printf(
					"%s %s (%s left)\n",
					color.CyanString("[%s]", bar),
					color.YellowString("%.1f%%", percent*100),
					color.HiBlackString(eta),
				)

			case <-done:
				fmt.Print("\033[2A\033[J")
				color.Green("[✓] Scan completed successfully.")
				return
			}
		}
	}()

	// Reserve two lines
	fmt.Println()
	fmt.Println()

	return done
}
