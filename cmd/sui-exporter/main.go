package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/rpcpool/sui-exporter/internal/suirpc"
)

const exporterVersion = "0.2.0"

var (
	grpcAddr          = flag.String("grpc", "", "Sui gRPC address as host:port, e.g. mainnet.sui.rpcpool.com:443")
	grpcTLS           = flag.Bool("grpc-tls", true, "Use TLS for the gRPC connection")
	grpcToken         = flag.String("grpc-token", "", "Optional x-token header sent with every gRPC call")
	addr              = flag.String("addr", ":8080", "Listen address")
	validatorMetrics  = flag.Bool("export-validator-metrics", true, "Export validator metrics")
	checkpointMetrics = flag.Bool("export-checkpoint-metrics", true, "Export checkpoint metrics")
	validatorReports  = flag.Bool("export-validator-reports", true, "Export validator reports")
	version           = flag.Bool("version", false, "print version")
	frequency         = flag.Int("frequency", 120, "frequency of metrics collection in seconds")
	timeout           = flag.Int("timeout", 10, "gRPC timeout in seconds")
)

func main() {
	log.SetFlags(0)
	flag.Parse()

	if *version {
		fmt.Println("Version:", exporterVersion)
		os.Exit(0)
	}

	if *grpcAddr == "" {
		log.Fatal("Please specify -grpc")
	}

	conn, err := dial(*grpcAddr, *grpcTLS, *grpcToken)
	if err != nil {
		log.Fatalf("Failed to create gRPC client for %s: %v", *grpcAddr, err)
	}
	defer conn.Close()

	log.Println("SUI Exporter", exporterVersion)
	log.Println("Listening on:", *addr+"/metrics")
	log.Println("Using SUI gRPC:", *grpcAddr, "tls:", *grpcTLS)
	log.Println("Updating every", *frequency, "seconds, gRPC timeout", *timeout, "seconds")
	log.Println("Exporting validator metrics:", *validatorMetrics, ", checkpoint metrics:", *checkpointMetrics, ", validator reports:", *validatorReports)

	exporter := NewExporter(suirpc.NewLedgerServiceClient(conn),
		time.Duration(*frequency)*time.Second, time.Duration(*timeout)*time.Second,
		*validatorMetrics, *checkpointMetrics, *validatorReports)
	prometheus.MustRegister(exporter)

	go exporter.WatchState()

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func dial(target string, useTLS bool, token string) (*grpc.ClientConn, error) {
	creds := insecure.NewCredentials()
	if useTLS {
		creds = credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12})
	}

	opts := []grpc.DialOption{grpc.WithTransportCredentials(creds)}
	if token != "" {
		opts = append(opts, grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, callOpts ...grpc.CallOption) error {
			return invoker(metadata.AppendToOutgoingContext(ctx, "x-token", token), method, req, reply, cc, callOpts...)
		}))
	}

	return grpc.NewClient(target, opts...)
}
