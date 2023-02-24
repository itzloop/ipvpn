package udp

import (
	"github.com/itzloop/ipvpn/devices/tun"
	"github.com/itzloop/ipvpn/pkg/log"
	"github.com/sirupsen/logrus"
	"net"
)

type Client struct {
	udpConn       *net.UDPConn
	networkDevice *tun.Interface
}

func NewClient(localNetworkDevice *tun.Interface, remote string) (*Client, error) {
	remoteAddr, err := net.ResolveUDPAddr("udp", remote)
	if err != nil {
		return nil, err
	}

	log.Logger.WithField("remote_addr", remoteAddr.String()).Info("udp client is dialing...")
	remoteConn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		return nil, err
	}

	client := &Client{
		udpConn:       remoteConn,
		networkDevice: localNetworkDevice,
	}

	return client, nil
}

// Start the client by creating 2 goroutines
// one for reading from network devices and writing
// to udp conn and the other is for reading from
// udp conn and reading it to network devices
func (client *Client) Start() error {
	go client.devToUdp()
	go client.udpToDev()

	return nil
}

// Close udp connection and network devices
func (client *Client) Close() error {
	err1 := client.udpConn.Close()
	err2 := client.networkDevice.Close()

	if err1 != nil {
		return err1
	}

	return err2
}

// send denotes receiving the packet from Client.networkDevice and
// sending it to server using Client.udpConn
func (client *Client) devToUdp() {
	bufLen := (2 << 15) - 1 // TODO config
	buf := make([]byte, bufLen)
	entry := log.Logger.WithFields(logrus.Fields{
		"dev_name": client.networkDevice.Name(),
		"dir":      "dev -> udp",
	})

	for {
		n, err := client.networkDevice.Read(buf)
		if err != nil {
			entry.WithField("error", err).
				Error("failed to read from devices")

			continue
		}

		if n == 0 {
			continue
		}

		// TODO encryption, authentication and compression

		entry.WithField("bytes_read", n).Info("read bytes from devices")
		// write to remote conn
		bufCopy := buf[:n]
		_, err = client.udpConn.Write(bufCopy)
		if err != nil {
			entry.WithField("error", err).
				Error("failed to write to conn")

			continue
		}
	}
}

// recv denotes receiving packets from server
func (client *Client) udpToDev() {
	bufLen := (2 << 15) - 1 // TODO config
	buf := make([]byte, bufLen)
	entry := log.Logger.WithFields(logrus.Fields{
		"dev_name": client.networkDevice.Name(),
		"dir":      "udp -> dev",
	})

	for {
		n, err := client.udpConn.Read(buf)
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

		b := buf[:n]
		_, err = client.networkDevice.Write(b)
		if err != nil {
			entry.WithField("error", err).
				Error("failed to write to devices")

			continue
		}
	}
}
