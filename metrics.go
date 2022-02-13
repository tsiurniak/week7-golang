package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	numberOfRegistered = promauto.NewCounter(prometheus.CounterOpts{
		Name: "number_of_registered_users",
		Help: "The total number of registered users",
	})
	numberOfCakesGiven = promauto.NewCounter(prometheus.CounterOpts{
		Name: "number_of_cakes_given",
		Help: "The total number of cakes given",
	})
	serverWorkTime = promauto.NewCounter(prometheus.CounterOpts{
		Name: "server_work_time",
		Help: "The server uptime from startup",
	})
)

func metrics() {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)

	log.Println("If you see me, prometheus didn`t start")
}
