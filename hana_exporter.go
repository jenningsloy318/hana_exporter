package main

import (
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/ini.v1"

	"github.com/jenningsloy318/hana_exporter/collector"
)

// define  flag
var (
	listenAddress = kingpin.Flag(
		"web.listen-address",
		"Address to listen on for web interface and telemetry.",
	).Default(":9105").String()
	metricPath = kingpin.Flag(
		"web.telemetry-path",
		"Path under which to expose metrics.",
	).Default("/metrics").String()
	configHanacnf = kingpin.Flag(
		"config.hana-cnf",
		"Path to hana.cnf file to read HANA credentials from.",
	).Default(path.Join(os.Getenv("HOME"), ".hana/hana.cnf")).String()
	dsn string
)

// scrapers lists all possible collection methods and if they should be enabled by default.
var scrapers = map[collector.Scraper]bool{
	collector.ScrapeGlobalStatus{}: true ,
}

func parseHanacnf(config interface{}) (string, error) {
	var dsn string
	opts := ini.LoadOptions{
		// HANA ini file can have boolean keys.
		AllowBooleanKeys: true,
	}
	cfg, err := ini.LoadSources(opts, config)
	if err != nil {
		return dsn, fmt.Errorf("failed reading ini file: %s", err)
	}
	user := cfg.Section("client").Key("user").String()
	password := cfg.Section("client").Key("password").String()
	if (user == "") || (password == "") {
		return dsn, fmt.Errorf("no user or password specified under [client] in %s", config)
	}
	host := cfg.Section("client").Key("host").String()
	port,err := cfg.Section("client").Key("port").Uint()
	if (host == "") || ( err != nil ) {
		return dsn, fmt.Errorf("no host or port specified under [client] in %s", config)
	}
	dsn = fmt.Sprintf("hdb://%s:%s@%s:%d", user, password, host, port)

	log.Debugln(dsn)
	return dsn, nil
}

func init() {
	prometheus.MustRegister(version.NewCollector("hana_exporter"))
}

// define new http handleer
func newHandler(scrapers []collector.Scraper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filteredScrapers := scrapers
		params := r.URL.Query()["collect[]"] // query if the request contains string collect[],which is defined in the prometheus scrape_configs, use this filter to enable only monitor perticular metrics for this hana instance
		log.Debugln("collect query:", params)  

		// Check if we have some "collect[]" query parameters.
		if len(params) > 0 {
			filters := make(map[string]bool)
			for _, param := range params {
				filters[param] = true
			}

			filteredScrapers = nil
			for _, scraper := range scrapers {
				if filters[scraper.Name()] {
					filteredScrapers = append(filteredScrapers, scraper)
				}
			}
		}

		registry := prometheus.NewRegistry()
		registry.MustRegister(collector.New(dsn, filteredScrapers))

		gatherers := prometheus.Gatherers{
			prometheus.DefaultGatherer,
			registry,
		}
		// Delegate http serving to Prometheus client library, which will call collector.Collect.
		h := promhttp.HandlerFor(gatherers, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
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

	dsn = os.Getenv("DATA_SOURCE_NAME")
	host := os.Getenv("HANA_HOST")
	port := int(os.Getenv("HANA_PORT"))
	user := os.Getenv("HANA_USER")
	password := os.Getenv("HANA_PASSWORD")


	if len(dsn) == 0   { 
		if len(host) != 0  && len(port) != 0 &&  len(dsn) == 0 &&  len(user) == 0 &&  len(password) == 0 {
		dsn = fmt.Sprintf("hdb://%s:%s@%s:%d", user, password, host, port)
		} else {
		var err error
		if dsn, err = parseHanacnf(*configHanacnf); err != nil {
			log.Fatal(err)
		}
	}
	// Register only scrapers enabled by flag.
	log.Infof("Enabled scrapers:")
	enabledScrapers := []collector.Scraper{}
	for scraper, enabled := range scraperFlags {
		if *enabled {
			log.Infof(" --collect.%s", scraper.Name())
			enabledScrapers = append(enabledScrapers, scraper)
		}
	}
	http.HandleFunc(*metricPath, prometheus.InstrumentHandlerFunc("metrics", newHandler(enabledScrapers)))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(landingPage)
	})

	log.Infoln("Listening on", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
