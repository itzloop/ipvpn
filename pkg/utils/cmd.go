package utils

import (
	"github.com/itzloop/ipvpn/pkg/log"
	"github.com/sirupsen/logrus"
	"os/exec"
	"strings"
)

func MustRunCmd(name string, args ...string) {
	if err := RunCmd(name, args...); err != nil {
		log.Logger.Panicln(err)
	}
}

func RunCmd(name string, args ...string) error {

	cmd := exec.Command(name, args...)
	cmd.Stdout = log.Logger.WriterLevel(logrus.InfoLevel)
	cmd.Stderr = log.Logger.WriterLevel(logrus.ErrorLevel)
	return cmd.Run()
}

func RunCmdWithLog(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = log.Logger.WriterLevel(logrus.InfoLevel)
	cmd.Stderr = log.Logger.WriterLevel(logrus.ErrorLevel)

	log.Logger.Debug(
		strings.Join(
			append([]string{"running cmd:", name}, args...), " ",
		),
	)

	return cmd.Run()
}
