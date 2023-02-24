package main

import (
	"flag"
	"github.com/itzloop/ipvpn/devices/tun"
	"github.com/itzloop/ipvpn/pkg/cache"
	"github.com/itzloop/ipvpn/pkg/config"
	"github.com/itzloop/ipvpn/pkg/log"
	"github.com/itzloop/ipvpn/transports/udp"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	serverMode := flag.Bool("server", false, "run in serverMode mode")
	cidr := flag.String("cidr", "10.0.3.0/24", "cidr used for the vpn")
	listerAddr := flag.String("l", "10.", "the address that the server listens on")
	remoteAddr := flag.String("r", "", "server address that client connects to")
	flag.Parse()

	ip, mask, err := net.ParseCIDR(*cidr)
	if err != nil {
		panic(err)
	}

	cfg := config.Config{
		CIDRConfig: config.CIDRConfig{
			CIDR: *cidr,
			Ip:   ip,
			Mask: mask,
		},
	}

	ch := cache.NewCache(&cfg)

	dev, err := tun.NewTunDevice(&cfg)
	if err != nil {
		panic(err)
	}

	defer dev.Close()

	if *serverMode {
		log.Logger.Info("running in server mode")
		srv, err := udp.NewServer(dev, ch, *listerAddr)
		if err != nil {
			panic(err)
		}

		log.Logger.Info("starting server...")
		if err = srv.Start(); err != nil {
			panic(err)
		}

		defer srv.Close()
	} else {
		log.Logger.Info("running in client mode")
		cl, err := udp.NewClient(dev, *remoteAddr)
		if err != nil {
			panic(err)
		}

		log.Logger.Info("starting client...")
		if err = cl.Start(); err != nil {
			panic(err)
		}

		defer cl.Close()
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	<-sig

}
