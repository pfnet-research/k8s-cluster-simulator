package log

import (
	"github.com/sirupsen/logrus"
)

func IsDebugEnabled() bool {
	return logrus.GetLevel() >= logrus.DebugLevel
}
