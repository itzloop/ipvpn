package log

import (
	"github.com/itzloop/ipvpn/pkg/config"
	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

func init() {
	Logger = logrus.New()
	// TODO configure
	Logger.SetLevel(logrus.DebugLevel)
}

func Configure(config config.Config) {
	// TODO configure logger
}
