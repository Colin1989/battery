package acceptor

import (
	"crypto/tls"
	"fmt"
	"github.com/colin1989/battery/constant"
	"net"
)

const (
	acceptorRunning int32 = iota + 1
	acceptorStopped
)

func loadCertificate(certs ...string) []tls.Certificate {
	var certificates []tls.Certificate
	if len(certs) != 2 && len(certs) != 0 {
		panic(constant.ErrIncorrectNumberOfCertificates)
	} else if len(certs) == 2 && certs[0] != "" && certs[1] != "" {
		cert, err := tls.LoadX509KeyPair(certs[0], certs[1])
		if err != nil {
			panic(fmt.Errorf("%w: %v", constant.ErrInvalidCertificates, err))
		}
		certificates = append(certificates, cert)
	}
	return certificates
}

func getListener(addr string, certs []tls.Certificate) (net.Listener, error) {
	if len(certs) == 0 {
		return net.Listen("tcp", addr)
	}

	tlsCfg := &tls.Config{
		Certificates: certs,
	}

	return tls.Listen("tcp", addr, tlsCfg)
}
