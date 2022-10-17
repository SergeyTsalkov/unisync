package progressbar

import (
	"fmt"
	"math"
	"strings"
	"time"
	"unisync/overwriter"
)

// 57% [=========>-----------] ETA: 4m33s
func Draw(progress int, eta int) error {
	if !overwriter.IsTerminal() {
		return fmt.Errorf("stdout is not terminal")
	}

	termWidth, err := overwriter.TerminalWidth()
	if err != nil {
		return err
	}

	width := termWidth / 2

	part1 := fmt.Sprintf("  %2v%%", progress)
	part2 := ""
	part3 := fmt.Sprintf("ETA: %v", time.Second*time.Duration(eta))

	coreWidth := width - len(part1) - len(part3) - 2

	if coreWidth > 0 {
		coreWidthFilled := int(math.Round(float64(coreWidth) * (float64(progress) / float64(100))))

		if progress == 100 {
			part2 = fmt.Sprintf("[%v]", strings.Repeat("=", coreWidth))
		} else if coreWidthFilled == 0 {
			part2 = fmt.Sprintf("[%v]", strings.Repeat("-", coreWidth))
		} else {
			part2 = fmt.Sprintf("[%v>%v]", strings.Repeat("=", coreWidthFilled-1), strings.Repeat("-", coreWidth-coreWidthFilled))
		}
	}

	err = overwriter.Println(part1, part2, part3)
	if err != nil {
		return err
	}

	return overwriter.Flush()
}

func Reset() error {
	return overwriter.Reset()
}
