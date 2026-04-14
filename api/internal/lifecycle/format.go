package lifecycle

import (
	"fmt"
	"strings"
	"time"
)

const badgeWidth = 12

func sgr(codes ...int) string {
	parts := make([]string, len(codes))
	for i, c := range codes {
		parts[i] = fmt.Sprintf("%d", c)
	}
	return "\x1b[" + strings.Join(parts, ";") + "m"
}

var (
	sgrReset = sgr(0)
	sgrBold  = sgr(1)
	sgrDim   = sgr(2)

	sgrFgRed   = sgr(31)
	sgrFgGreen = sgr(32)
)

func makeBadge(label string, codes ...int) string {
	pad := badgeWidth - len(label) - 2
	prefix := ""
	if pad > 0 {
		prefix = strings.Repeat(" ", pad)
	}
	return prefix + sgr(codes...) + " " + label + " " + sgrReset
}

var phaseBadges = map[string]string{
	"setup":      makeBadge("SETUP", 1, 97, 46),
	"post_start": makeBadge("POST-START", 1, 97, 44),
	"health":     makeBadge("HEALTH", 1, 97, 42),
	"snapshot":   makeBadge("SNAPSHOT", 1, 97, 46),
	"event":      makeBadge("EVENT", 1, 97, 45),
	"pre_stop":   makeBadge("PRE-STOP", 1, 97, 45),
	"shutdown":   makeBadge("SHUTDOWN", 1, 97, 41),
	"expired":    makeBadge("EXPIRED", 1, 97, 43),
}

var gutter = strings.Repeat(" ", len("15:04:05")+1+badgeWidth+1)

func FormatEntry(e Entry) string {
	ts := sgrDim + e.Time.Format(time.TimeOnly) + sgrReset
	badge := phaseBadge(e.Phase)
	prefix := ts + " " + badge + " "

	switch e.Level {
	case LevelInfo:
		return prefix + sgrBold + e.Message + sgrReset
	case LevelSuccess:
		return prefix + sgrFgGreen + e.Message + sgrReset
	case LevelError:
		return prefix + sgrFgRed + sgrBold + e.Message + sgrReset
	case LevelOutput:
		return gutter + sgrDim + "│" + sgrReset + " " + e.Message
	case LevelDetail:
		return gutter + sgrDim + "│ " + e.Message + sgrReset
	case LevelWait:
		return gutter + sgrDim + "· " + e.Message + sgrReset
	default:
		return prefix + e.Message
	}
}

func IsVerbose(level Level) bool {
	return level == LevelOutput || level == LevelDetail || level == LevelWait
}

func phaseBadge(phase string) string {
	if badge, ok := phaseBadges[phase]; ok {
		return badge
	}
	return makeBadge(phase, 1, 97, 44)
}
