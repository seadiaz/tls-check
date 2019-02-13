package network

import (
	"crypto/tls"
	"fmt"
	"math"
	"net"
	"os"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

const evaluationPeriodInSeconds = 30 * time.Second

//ValidateHost ...
func ValidateHost(host string, timeout time.Duration, channel chan int) {
	ips := lookupIPWithTimeout(host, timeout)
	channel <- validateIPs(ips, host, timeout)
	timer := time.NewTicker(evaluationPeriodInSeconds)
	for range timer.C {
		log.Debugf("lookup result: %v", ips)
		channel <- validateIPs(ips, host, timeout)
	}
}

func lookupIPWithTimeout(host string, timeout time.Duration) []net.IP {
	timer := time.NewTimer(timeout)

	ch := make(chan []net.IP, 1)
	go resolveIP(host, ch)
	select {
	case ips := <-ch:
		return ips
	case <-timer.C:
		log.Errorf("timeout resolving %s", host)
	}
	return make([]net.IP, 0)
}

func resolveIP(host string, ch chan []net.IP) {
	r, err := net.LookupIP(host)
	if err != nil {
		log.Fatal(err)
	}
	ch <- r
}

func validateIPs(ips []net.IP, host string, timeout time.Duration) int {
	for _, ip := range ips {
		dialer := net.Dialer{Timeout: timeout, Deadline: time.Now().Add(timeout + 5*time.Second)}
		connection, err := tls.DialWithDialer(&dialer, "tcp", fmt.Sprintf("[%s]:443", ip), &tls.Config{ServerName: host})
		if err != nil {
			if ip.To4() == nil {
				switch err.(type) {
				case *net.OpError:
					// https://stackoverflow.com/questions/38764084/proper-way-to-handle-missing-ipv6-connectivity
					if err.(*net.OpError).Err.(*os.SyscallError).Err == syscall.EHOSTUNREACH {
						log.Infof("%-15s - ignoring unreachable IPv6 address", ip)
						continue
					}
				}
			}
			log.Errorf("%s: %s", ip, err)
			continue
		}
		defer connection.Close()

		checkedCerts := make(map[string]struct{})
		for _, chain := range connection.ConnectionState().VerifiedChains {
			for _, cert := range chain {
				if _, checked := checkedCerts[string(cert.Signature)]; checked {
					continue
				}
				checkedCerts[string(cert.Signature)] = struct{}{}
				if cert.IsCA {
					log.Debugf("%-15s - ignoring CA certificate %s", ip, cert.Subject.CommonName)
					continue
				}

				validityInHours := cert.NotAfter.Sub(time.Now())
				validityInDays := getDays(validityInHours)
				log.Infof("%s expiration: %d days", cert.Subject.CommonName, validityInDays)
				return validityInDays
			}
		}
	}

	return -1
}

func getDays(input time.Duration) int {
	return int(math.Floor(input.Hours() / 24))
}
