package config

import "net"

type Config struct {
	CIDRConfig CIDRConfig
}

type CIDRConfig struct {
	CIDR string
	Ip   net.IP
	Mask *net.IPNet
}
