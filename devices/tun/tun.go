package tun

import (
	"github.com/itzloop/ipvpn/pkg/config"
	"github.com/itzloop/ipvpn/pkg/log"
	"github.com/itzloop/ipvpn/pkg/utils"
	"github.com/net-byte/water"
	"github.com/sirupsen/logrus"
	"runtime"
	"sync"
)

type Interface struct {
	*water.Interface
	closed bool
	l      sync.Mutex
	cfg    *config.Config
}

func NewTunDevice(cfg *config.Config) (*Interface, error) {
	// TODO configure

	entry := log.Logger.WithFields(logrus.Fields{
		"dev_name": "tun0",
		"cidr":     cfg.CIDRConfig.CIDR,
	})

	entry.Info("creating tun device")

	c := water.Config{DeviceType: water.TUN}

	c.PlatformSpecificParams = water.PlatformSpecificParams{}
	if runtime.GOOS == "windows" {
		c.PlatformSpecificParams = water.PlatformSpecificParams{
			Name:    "tun0",
			Network: []string{cfg.CIDRConfig.CIDR},
		}
	} else {
		c.PlatformSpecificParams.Name = "tun0"
	}

	dev, err := water.New(c)
	if err != nil {
		return nil, err
	}

	devWrapper := Interface{
		Interface: dev,
		closed:    false,
		l:         sync.Mutex{},
		cfg:       cfg,
	}

	entry.Info("configuring tun device")
	if err = devWrapper.configure(); err != nil {
		return nil, err
	}

	return &devWrapper, nil
}

func (i *Interface) configure() error {
	// TODO configure from config.Config
	if err := utils.RunCmdWithLog("/sbin/ip", "addr", "add", i.cfg.CIDRConfig.CIDR, "dev", i.Name()); err != nil {
		return err
	}

	return utils.RunCmdWithLog("/sbin/ip", "link", "set", "dev", i.Name(), "up")
}

func (i *Interface) Close() error {
	i.l.Lock()
	defer i.l.Unlock()
	if i.closed {
		return nil
	}

	i.closed = true
	err := utils.RunCmdWithLog("/sbin/ip", "link", "set", "dev", i.Name(), "down")
	err2 := i.Interface.Close()
	if err != nil {
		return err
	}

	return err2
}
