package main

import (
	"context"
	"fmt"
	_ "github.com/SAP/go-hdb/driver"
	"github.com/jenningsloy318/hana_exporter/collector"
	"github.com/jenningsloy318/hana_exporter/config"
	"go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"go.opencensus.io/zpages"
	"gopkg.in/alecthomas/kingpin.v2"
	"html/template"
	"log"
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
	version string 
	branch string
	commit string
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

	dsn := fmt.Sprintf("hdb://%s:%s@%s", user, password, target)

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

	// configure enabled collectors
	var viewCollectors = map[collector.ViewCollector]bool{
		collector.DisksCollector{}:   true,
		collector.LicenseCollector{}: true,
	}
	enabledCollectors := map[collector.ViewCollector]bool{}

	// register views according to enabled collector
	for viewCollector, flag := range viewCollectors {
		if flag {
			enabledCollectors[viewCollector] = flag
			if err := view.Register(viewCollector.Views()...); err != nil {
				log.Fatalf("Failed to register hana views for metrics: %v", err)
			}

		}
	}

	// start strace and collect data
	ctx, span := trace.StartSpan(context.Background(), "metrics_data_trace")
	span.Annotate([]trace.Attribute{trace.StringAttribute("step", "Initial")}, "This is first span of the trace")
	stime := time.Now().String()
	log.Printf("Starting to fectch data at %s", stime)
	span.Annotate([]trace.Attribute{}, fmt.Sprintf("Starting to fectch data at %s.", stime))

	defer span.End()
	collector.Collect(ctx, dsn, enabledCollectors)

	view.SetReportingPeriod(1 * time.Second)
	prometheusExporter.ServeHTTP(w, r)

}

func main() {
	// Parse flags.
	Version := fmt.Sprintf("hana_exporter Version %s , Branch %s ,Commit %s", version, branch, commit)
	kingpin.Version(Version)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	

	if err := sc.ReloadConfig(*configFile); err != nil {
		log.Fatalf("Error parsing config file: %s", err)
	}
	// create jaeger exporter
	var jaegerConfig config.JaegerConfig
	var err error 
	if jaegerConfig, err = sc.ParseJaegerConfig(); err != nil {
		log.Fatalf("Error getting jarger config,%s",  err)
	}

	JaegerExporter, err := jaeger.NewExporter(jaeger.Options{
		AgentEndpoint: jaegerConfig.AgentEndpointURI,
		Endpoint:      jaegerConfig.CollectorEndpointURI,
		ServiceName:   "hana_exporter",
	})
	if err != nil {
		log.Fatalf("Failed to create the Jaeger exporter: %v", err)
	}
	trace.RegisterExporter(JaegerExporter)

	// apply trace config
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	// Ensure that the Prometheus endpoint is exposed for scraping

	// http server mux
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
