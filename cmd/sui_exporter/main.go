package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	rpcAddr   = flag.String("rpcURI", "https://testnet.sui.rpcpool.com", "Solana RPC URI (including protocol and path)")
	addr      = flag.String("addr", ":8080", "Listen address")
	version   = flag.Bool("version", false, "print version")
	frequency = flag.Int(("frequency"), 120, "frequency of metrics collection")
)

func main() {
	log.SetFlags(0)

	flag.Parse()

	if *version {
		fmt.Println("Version: 0.1")
		os.Exit(0)
	}

	if *rpcAddr == "" {
		log.Fatal("Please specify -rpcURI")
	}

	exporter := NewExporter(*rpcAddr, *frequency)
	prometheus.MustRegister(exporter)

	go exporter.WatchState()

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))

}
