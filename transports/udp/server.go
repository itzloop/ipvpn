package udp

import (
	"github.com/itzloop/ipvpn/devices/tun"
	"github.com/itzloop/ipvpn/pkg/log"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

type Server struct {
	udpConn       *net.UDPConn
	networkDevice *tun.Interface
	ipCache       *cache.Cache
}

func NewServer(networkDevice *tun.Interface, ipCache *cache.Cache, local string) (*Server, error) {
	localAddr, err := net.ResolveUDPAddr("udp", local)
	if err != nil {
		return nil, err
	}

	log.Logger.WithField("local_addr", localAddr.String()).Info("udp server is listening...")
	localonn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		return nil, err
	}

	client := &Server{
		udpConn:       localonn,
		networkDevice: networkDevice,
		ipCache:       ipCache,
	}

	return client, nil
}

func (server *Server) Start() error {
	go server.devToUdp()
	go server.udpToDev()

	return nil
}

func (server *Server) Close() error {
	err1 := server.udpConn.Close()
	err2 := server.networkDevice.Close()

	if err1 != nil {
		return err1
	}

	return err2
}

func (server *Server) devToUdp() {
	bufLen := (2 << 15) - 1 // TODO config
	buf := make([]byte, bufLen)
	entry := log.Logger.WithFields(logrus.Fields{
		"dev_name": server.networkDevice.Name(),
		"dir":      "dev -> udp",
	})

	for {
		n, err := server.networkDevice.Read(buf)
		if err != nil {
			entry.WithField("error", err).
				Error("failed to read from devices")

			continue
		}

		if n == 0 {
			continue
		}

		// TODO encryption, authentication and compression

		entry.WithField("bytes_read", n).Info("read bytes from conn")

		bufCopy := buf[:n]

		if len(bufCopy) < 20 || !isIpv4(bufCopy) {
			entry.Error("invalid ipv4 header")
			continue
		}

		src, dst := getSrcDst(bufCopy)

		tempEntry := entry.WithFields(logrus.Fields{
			"src": src.String(),
			"dst": dst.String(),
		})

		v, exists := server.ipCache.Get(dst.String())
		if !exists {
			tempEntry.Error("can't find address in cache")
			continue
		}

		clientAddr, ok := v.(*net.UDPAddr)
		if !ok {
			tempEntry.Error("expected v to be of type *net.UDPAddr but got %T", v)
		}

		entry.WithFields(logrus.Fields{
			"client_addr": clientAddr.String(),
			"src":         src.String(),
			"dst":         dst.String(),
		}).Info("sending packet to client")

		if _, err = server.udpConn.WriteTo(bufCopy, clientAddr); err != nil {
			entry.WithField("error", err).
				Error("failed to write to conn")

			continue
		}

	}
}

func (server *Server) udpToDev() {
	bufLen := (2 << 15) - 1 // TODO config
	buf := make([]byte, bufLen)
	entry := log.Logger.WithFields(logrus.Fields{
		"dev_name": server.networkDevice.Name(),
		"dir":      "udp -> dev",
	})

	for {
		n, clientAddr, err := server.udpConn.ReadFromUDP(buf)
		if err != nil {
			entry.WithField("error", err).
				Error("failed to read from conn")
			continue
		}

		if n == 0 {
			continue
		}

		// TODO decryption, authentication and decompression

		entry.WithField("bytes_read", n).Info("read bytes from conn")

		bufCopy := buf[:n]

		if len(bufCopy) < 20 || !isIpv4(bufCopy) {
			entry.Error("invalid ipv4 header")
			continue
		}

		src, dst := getSrcDst(bufCopy)

		entry.WithFields(logrus.Fields{
			"client_addr": clientAddr.String(),
			"src":         src.String(),
			"dst":         dst.String(),
		}).Info("sending packet to dst")

		server.ipCache.Set(src.String(), clientAddr, time.Hour)
		if _, err = server.networkDevice.Write(bufCopy); err != nil {
			entry.WithField("error", err).
				Error("failed to write to devices")

			continue
		}

	}
}

func isIpv4(b []byte) bool {
	return (b[0] >> 4) == 4
}

func getSrcDst(b []byte) (net.IP, net.IP) {
	return net.IPv4(b[12], b[13], b[14], b[15]), net.IPv4(b[16], b[17], b[18], b[19])
}
