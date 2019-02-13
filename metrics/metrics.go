package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	log "github.com/sirupsen/logrus"
)

var tlsCertificateValidityInDays prometheus.Gauge
var tlsCertificateValidityInHours prometheus.Gauge

//Start ...
func Start(host string, addr string, channel chan []int) {
	reg := prometheus.NewRegistry()
	tlsCertificateValidityInDays = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "tls_certificate_validity_days",
		Help:        "How many days remain before certificate expire",
		ConstLabels: prometheus.Labels{"host": host},
	})
	tlsCertificateValidityInHours = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "tls_certificate_validity_hours",
		Help:        "How many hours remain before certificate expire",
		ConstLabels: prometheus.Labels{"host": host},
	})
	reg.MustRegister(tlsCertificateValidityInDays)
	reg.MustRegister(tlsCertificateValidityInHours)
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	go updateMetrics(channel)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func updateMetrics(channel chan []int) {
	for {
		value := <-channel
		log.Infof("metric received")
		tlsCertificateValidityInDays.Set(float64(value[1]))
		tlsCertificateValidityInHours.Set(float64(value[0]))
	}
}
