package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	log "github.com/sirupsen/logrus"
)

var tlsCertificateValidity prometheus.Gauge

//Start ...
func Start(host string, addr string, channel chan int) {
	reg := prometheus.NewRegistry()
	tlsCertificateValidity = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "tls_certificate_validity",
		Help:        "How many days remain before certificate expire",
		ConstLabels: prometheus.Labels{"host": host},
	})
	reg.MustRegister(tlsCertificateValidity)
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	go updateMetrics(channel)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func updateMetrics(channel chan int) {
	for {
		value := <-channel
		log.Infof("metric received")
		tlsCertificateValidity.Set(float64(value))
	}
}
