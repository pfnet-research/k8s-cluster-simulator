package log

import (
	"github.com/sirupsen/logrus"
)

// IsDebugEnabled returns whether the debug log is enabled.
func IsDebugEnabled() bool {
	return logrus.GetLevel() >= logrus.DebugLevel
}
