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
	rpcAddr           = flag.String("rpcURI", "https://testnet.sui.rpcpool.com", "Solana RPC URI (including protocol and path)")
	addr              = flag.String("addr", ":8080", "Listen address")
	validatorMetrics  = flag.Bool("export-validator-metrics", true, "Export validator metrics")
	checkpointMetrics = flag.Bool("export-checkpoint-metrics", true, "Export checkpoint metrics")
	validatorReports  = flag.Bool("export-validator-reports", true, "Export validator reports")
	version           = flag.Bool("version", false, "print version")
	frequency         = flag.Int(("frequency"), 120, "frequency of metrics collection")
	timeout           = flag.Int(("timeout"), 10, "rpc timeout")
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

	log.Println("SUI Exporter v0.1")
	log.Println("Listening on: ", *addr+"/metrics")
	log.Println("Using SUI RPC: ", *rpcAddr)
	log.Println("Updating every", *frequency, "seconds, rpc timeout", *timeout, "seconds")
	log.Println("Exporting base metrics: true, Exporting validator metrics:", *validatorMetrics, ", Exporting checkpoint metrics:", *checkpointMetrics, ", Exporting validator reports:", *validatorReports)

	exporter := NewExporter(*rpcAddr, *frequency, *timeout, *validatorMetrics, *checkpointMetrics, *validatorReports)
	prometheus.MustRegister(exporter)

	go exporter.WatchState()

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))

}
