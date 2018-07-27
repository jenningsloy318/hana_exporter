package main

import (
	"fmt"
	"github.com/jenningsloy318/hana_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// define  flag
var (
	listenAddress = kingpin.Flag(
		"web.listen-address",
		"Address to listen on for web interface and telemetry.",
	).Default(":9460").String()
	metricPath = kingpin.Flag(
		"web.telemetry-path",
		"Path under which to expose metrics.",
	).Default("/hana").String()
	configFile = kingpin.Flag("config.file", "Path to configuration file.").Default("hana-exporter.yml").String()
	dsn        string
	sc         = &SafeConfig{
		C: &Config{},
	}
	reloadCh chan chan error
)

// scrapers lists all possible collection methods and if they should be enabled by default.
var scrapers = map[collector.Scraper]bool{
	collector.ScrapeHostResourceUtilization{}: true,
	collector.ScrapeServiceStatistics{}:       true,
	collector.ScrapeLicenseStatus{}:           true,
	collector.ScrapeDisks{}:                   true,
	collector.ScrapeSharedMemory{}:            true,
}

func init() {
	prometheus.MustRegister(version.NewCollector("hana_exporter"))
}

// define new http handleer
func newHandler(scrapers []collector.Scraper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.Query().Get("target")
		if target == "" {
			http.Error(w, "'target' parameter must be specified", 400)
			return
		}
		log.Debugf("Scraping target '%s'", target)
		var targetCredentials Credentials
		var err error
		if targetCredentials, err = sc.CredentialsForTarget(target); err != nil {
			log.Fatalf("Error getting credentialfor target %s file: %s", target, err)
		}
		user := targetCredentials.User
		password := targetCredentials.Password
		// construct dsn
		dsn = fmt.Sprintf("hdb://%s:%s@%s", user, password, target)

		registry := prometheus.NewRegistry()

		collector := collector.New(dsn, scrapers)
		registry.MustRegister(collector)

		gatherers := prometheus.Gatherers{
			prometheus.DefaultGatherer,
			registry,
		}
		// Delegate http serving to Prometheus client library, which will call collector.Collect.
		h := promhttp.HandlerFor(gatherers, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	}
}

func updateConfiguration(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		rc := make(chan error)
		reloadCh <- rc
		if err := <-rc; err != nil {
			http.Error(w, fmt.Sprintf("failed to reload config: %s", err), http.StatusInternalServerError)
		}
	default:
		log.Errorf("POST method expected")
		http.Error(w, "POST method expected", 400)
	}
}

func main() {
	// Generate ON/OFF flags for all scrapers.
	scraperFlags := map[collector.Scraper]*bool{}
	for scraper, enabledByDefault := range scrapers {
		defaultOn := "false"
		if enabledByDefault {
			defaultOn = "true"
		}

		f := kingpin.Flag(
			"collect."+scraper.Name(),
			scraper.Help(),
		).Default(defaultOn).Bool()

		scraperFlags[scraper] = f
	}

	// Parse flags.
	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("hana_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	if err := sc.ReloadConfig(*configFile); err != nil {
		log.Fatalf("Error parsing config file: %s", err)
	}

	// landingPage contains the HTML served at '/'.
	// TODO: Make this nicer and more informative.
	var landingPage = []byte(`<html>
<head><title>HANA exporter</title></head>
<body>
<h1>HANA exporter</h1>
<p><a href='` + *metricPath + `'>Metrics</a></p>
</body>
</html>
`)

	log.Infoln("Starting hana_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	// Register only scrapers enabled by flag.
	log.Infof("Enabled scrapers:")
	enabledScrapers := []collector.Scraper{}
	for scraper, enabled := range scraperFlags {
		if *enabled {
			log.Infof(" --collect.%s", scraper.Name())
			enabledScrapers = append(enabledScrapers, scraper)
		}
	}

	// load config  first time
	hup := make(chan os.Signal)
	reloadCh = make(chan chan error)
	signal.Notify(hup, syscall.SIGHUP)

	go func() {
		for {
			select {
			case <-hup:
				if err := sc.ReloadConfig(*configFile); err != nil {
					log.Errorf("Error reloading config: %s", err)
				}
			case rc := <-reloadCh:
				if err := sc.ReloadConfig(*configFile); err != nil {
					log.Errorf("Error reloading config: %s", err)
					rc <- err
				} else {
					rc <- nil
				}
			}
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc(*metricPath, prometheus.InstrumentHandlerFunc("metrics", newHandler(enabledScrapers)))
	http.HandleFunc("/-/reload", updateConfiguration) // reload config
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(landingPage)
	})

	log.Infoln("Listening on", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
