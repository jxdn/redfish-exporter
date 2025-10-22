package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"redfish-exporter/internal/exporter"
)

func main() {
	configFile := flag.String("config.file", "", "Path to YAML config file")
	listenAddress := flag.String("web.listen-address", ":9102", "Address to listen on for HTTP requests.")
	redfishHost := flag.String("redfish.host", "", "Redfish host base URL, e.g. https://<ip or host>")
	username := flag.String("redfish.username", "", "Redfish username")
	password := flag.String("redfish.password", "", "Redfish password")
	insecure := flag.Bool("redfish.insecure", false, "Skip TLS verification")
	scrapeTimeout := flag.Duration("scrape.timeout", 10*time.Second, "Timeout for Redfish requests")
	chassisID := flag.String("redfish.chassis-id", "", "Optional chassis ID to restrict scraping; if empty, scrape all")
	flag.Parse()

	// Load config file if provided
	if *configFile != "" {
		cfg, err := exporter.LoadConfig(*configFile)
		if err != nil {
			log.Fatalf("failed to load config: %v", err)
		}
		if cfg.Web.ListenAddress != "" && (listenAddress != nil && *listenAddress == ":9102") {
			*listenAddress = cfg.Web.ListenAddress
		}
		if cfg.Redfish.Host != "" && *redfishHost == "" {
			*redfishHost = cfg.Redfish.Host
		}
		if cfg.Redfish.Username != "" && *username == "" {
			*username = cfg.Redfish.Username
		}
		if cfg.Redfish.Password != "" && *password == "" {
			*password = cfg.Redfish.Password
		}
		if cfg.Redfish.ChassisID != "" && *chassisID == "" {
			*chassisID = cfg.Redfish.ChassisID
		}
		if cfg.Redfish.TimeoutSec > 0 && (*scrapeTimeout == 10*time.Second || *scrapeTimeout == 0) {
			*scrapeTimeout = time.Duration(cfg.Redfish.TimeoutSec) * time.Second
		}
		// insecure flag: config only applies if flag not explicitly set true
		if *insecure == false {
			*insecure = cfg.Redfish.InsecureTLS
		}
	}

	if *redfishHost == "" || *username == "" || *password == "" {
		log.Fatal("Redfish host, username and password are required via flags or config file")
	}

	collector, err := exporter.NewRedfishCollector(exporter.RedfishCollectorConfig{
		BaseURL:       *redfishHost,
		Username:      *username,
		Password:      *password,
		InsecureTLS:   *insecure,
		ScrapeTimeout: *scrapeTimeout,
		ChassisID:     *chassisID,
	})
	if err != nil {
		log.Fatalf("failed to create collector: %v", err)
	}

	reg := prometheus.NewRegistry()
	reg.MustRegister(collector)

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	log.Printf("Starting redfish exporter on %s", *listenAddress)
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}
