package utils

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"path"
	"strings"
)

type ConsoleFormatter struct {
	logrus.TextFormatter
}

func (c *ConsoleFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	logStr := fmt.Sprintf("%s %s %s:%d %v\n",
		entry.Time.Format("2006/01/02 15:04:05"),
		strings.ToUpper(entry.Level.String()),
		path.Base(entry.Caller.File),
		entry.Caller.Line,
		entry.Message,
	)
	return []byte(logStr), nil
}
