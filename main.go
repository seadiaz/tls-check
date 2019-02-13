package main

import (
	"flag"
	"time"

	"github.com/seadiaz/tls-checker/metrics"
	"github.com/seadiaz/tls-checker/network"

	log "github.com/sirupsen/logrus"
)

var host = flag.String("host", "", "the domain name of the host to check")
var lookupTimeout = flag.Duration("lookup-timeout", 10*time.Second, "timeout for DNS lookups - see: https://golang.org/pkg/time/#ParseDuration")
var connectionTimeout = flag.Duration("connection-timeout", 30*time.Second, "timeout connection - see: https://golang.org/pkg/time/#ParseDuration")
var addr = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")

func main() {
	flag.Parse()

	log.SetLevel(log.DebugLevel)

	if *host == "" {
		flag.Usage()
		log.Panicf("host is required")
	}

	channel := make(chan int)

	go network.ValidateHost(*host, *lookupTimeout, channel)

	metrics.Start(*host, *addr, channel)
}
