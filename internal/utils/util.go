package utils

import (
	"strings"
	"time"
)

func FmtDuration(d time.Duration) string {
	slice := strings.Split(d.String(), ".")
	if len(slice) == 2 {
		return slice[0] + "s"
	}
	return d.String()
}
