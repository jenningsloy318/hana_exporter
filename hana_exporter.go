package main

import (
	"fmt"
	_ "github.com/SAP/go-hdb/driver"
	"github.com/jenningsloy318/hana_exporter/collector"
	"github.com/jenningsloy318/hana_exporter/config"
	"go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/zpages"
	"gopkg.in/alecthomas/kingpin.v2"
	"html/template"
	"log"
	"net"
	"net/http"
	"time"
)

const (
	html = `<html>
	<head><title>opencensus example</title></head>
	<body>
	<h1>OpenCensus Example</h1>
	<p><a href="/metrics">metrics</a></p>
	<p><a href="/debug/rpcz">rpcz</a></p>
	<p><a href="/debug/tracez">tracez</a></p>
	</body>
	</html>`
)

var (
	listenAddress = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9460").String()
	metricPath    = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/hana").String()
	configFile    = kingpin.Flag("config.file", "Path to configuration file.").Default("hana.yml").String()
	sc            = &config.SafeConfig{
		C: &config.Config{},
	}
	version   string
	branch    string
	commit    string
	buildUser string
	buildHost string
	dsn string
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("home").Parse(html)
	if err != nil {
		log.Fatalf("Cannot parse template: %v", err)
	}
	t.Execute(w, "")
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	// retrieve target from request

	target := r.URL.Query().Get("target")

	if target == "" {
		http.Error(w, "'target' parameter must be specified", 400)
		return
	}
	log.Printf("Scraping target '%s'", target)

	// Get credentials for target
	var targetCredentials config.Credentials
	var err error
	if targetCredentials, err = sc.CredentialsForTarget(target); err != nil {
		log.Fatalf("Error getting credentialfor target %s file: %s", target, err)
	}
	user := targetCredentials.User
	password := targetCredentials.Password

	dsn = fmt.Sprintf("hdb://%s:%s@%s", user, password, target)

	// create prometheus exporter

	prometheusExporter, err := prometheus.NewExporter(prometheus.Options{
		Namespace: "hana",
	})

	if err != nil {
		log.Fatalf("Failed to create the Prometheus exporter: %v", err)
	}

	// register prometheus exporter to view

	view.RegisterExporter(prometheusExporter)

	// register default DefaultServerViews to view

	if err := view.Register(ochttp.DefaultServerViews...); err != nil {
		log.Fatalf("Failed to register http default server views for metrics: %v", err)
	}
	hanaViews := collector.NewHanaViews() 
	// register views according to enabled collector
			if err := view.Register(hanaViews...); err != nil {
				log.Fatalf("Failed to register hana views for metrics: %v", err)
			}


	// retrieve the remote client address and add it as attribute to the root span
	IP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		fmt.Fprintf(w, "userip: %q is not IP:port", r.RemoteAddr)
	}
	clientIP := net.ParseIP(IP)
	if clientIP == nil {
		//return nil, fmt.Errorf("userip: %q is not IP:port", req.RemoteAddr)
		fmt.Fprintf(w, "userip: %q is not IP:port", r.RemoteAddr)
	} else {
	}

	// start  collect data
	stime := time.Now().String()
	log.Printf("Starting to fectch data at %s", stime)
	view.SetReportingPeriod(1 * time.Second)
	prometheusExporter.ServeHTTP(w, r)

}

func main() {
	// Parse flags.
	versionInfo := fmt.Sprintf("hana_exporter version: %s , Branch: %s ,Commit ID: %s , Built by %s on %s at %s", version, branch, commit, buildUser, buildHost, time.Now().String())
	kingpin.Version(versionInfo)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	if err := sc.ReloadConfig(*configFile); err != nil {
		log.Fatalf("Error parsing config file: %s", err)
	}



	// http server mux

	go collector.NewHanaMeasurements(dsn)

	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metricsHandler(w, r)
	})
	mux.HandleFunc("/", homeHandler)
	mux.Handle("/debug/", http.StripPrefix("/debug", zpages.Handler))
	h := &ochttp.Handler{Handler: mux}

	if err := http.ListenAndServe(*listenAddress, h); err != nil {
		log.Fatalf("HTTP server ListenAndServe error: %v", err)
	}

}
