package acceptor

import (
	"crypto/tls"
	"fmt"
	"net"

	"github.com/colin1989/battery/errors"
)

const (
	acceptorRunning int32 = iota + 1
	acceptorStopped
)

func loadCertificate(certs ...string) []tls.Certificate {
	var certificates []tls.Certificate
	if len(certs) != 2 && len(certs) != 0 {
		panic(errors.ErrIncorrectNumberOfCertificates)
	} else if len(certs) == 2 && certs[0] != "" && certs[1] != "" {
		cert, err := tls.LoadX509KeyPair(certs[0], certs[1])
		if err != nil {
			panic(fmt.Errorf("%w: %v", errors.ErrInvalidCertificates, err))
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
