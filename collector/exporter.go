package collector

import (
	"database/sql"
	"sync"
	"time"
	_ "github.com/SAP/go-hdb/driver"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"strings"
)

// Metric name parts.
const (
	// Subsystem(s).
	exporter = "exporter"
)
// SQL Queries.
const (	
		upQuery = `select 1 from dummy;`
	)
// Metric descriptors.
var (
	scrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, exporter, "collector_duration_seconds"),
		"Collector time duration.",
		[]string{"collector"}, nil,
	)
	 Hana_instance string
)

// Exporter collects HANA metrics. It implements prometheus.Collector.
type Exporter struct {
	dsn          string
	scrapers     []Scraper
	error        prometheus.Gauge
	totalScrapes prometheus.Counter
	scrapeErrors *prometheus.CounterVec
	hanaUp     prometheus.Gauge
}
// split string, use @ as delimiter to split the dsn to get hana instance

func stringsplit(s rune) bool {
	if s == '@' {
	 return true
	}
	return false
 }


// New returns a new HANA exporter for the provided DSN.
func New(dsn string, scrapers []Scraper) *Exporter {
	Hana_instance =  strings.FieldsFunc(dsn, stringsplit)[1]

	return &Exporter{
		dsn:      dsn,
		scrapers: scrapers,
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: exporter,
			Name:      "scrapes_total",
			Help:      "Total number of times HANA was scraped for metrics.",
		}),
		scrapeErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: exporter,
			Name:      "scrape_errors_total",
			Help:      "Total number of times an error occurred scraping a HANA.",
		}, []string{"collector"}),
		error: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: exporter,
			Name:      "last_scrape_error",
			Help:      "Whether the last scrape of metrics from HANA resulted in an error (1 for error, 0 for success).",
		}),
		
	hanaUp: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "up",
			Help:      "Whether the HANA server is up.",
			ConstLabels: prometheus.Labels{"hana_instance": Hana_instance},
			}),
	}
}

// Describe implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	// We cannot know in advance what metrics the exporter will generate
	// from HANA. So we use the poor man's describe method: Run a collect
	// and send the descriptors of all the collected metrics. The problem
	// here is that we need to connect to the HANA DB. If it is currently
	// unavailable, the descriptors will be incomplete. Since this is a
	// stand-alone exporter and not used as a library within other code
	// implementing additional metrics, the worst that can happen is that we
	// don't detect inconsistent metrics created by this exporter
	// itself. Also, a change in the monitored HANA instance may change the
	// exported metrics during the runtime of the exporter.

	metricCh := make(chan prometheus.Metric)
	doneCh := make(chan struct{})

	go func() {
		for m := range metricCh {
			ch <- m.Desc()
		}
		close(doneCh)
	}()

	e.Collect(metricCh)
	close(metricCh)
	<-doneCh
}

// Collect implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {

	e.scrape(ch)

	ch <- e.totalScrapes
	ch <- e.error
	ch <- e.hanaUp
	e.scrapeErrors.Collect(ch)
}

func (e *Exporter) scrape(ch chan<- prometheus.Metric) {
	e.totalScrapes.Inc()
	var err error
  scrapeTime := time.Now()
	db, err := sql.Open("hdb", e.dsn)
	if err != nil {
		log.Errorln("Error opening connection to database:", err)
		e.error.Set(1)
		return
	}
	defer db.Close()

	// By design exporter should use maximum one connection per request.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	// Set max lifetime for a connection.
	db.SetConnMaxLifetime(1 * time.Minute)

		isUpRows, err := db.Query(upQuery)
		if err != nil {
			log.Errorln("Error pinging hana:", err)
			e.hanaUp.Set(0)
			e.error.Set(1)
			return
		}
		isUpRows.Close()
	
		e.hanaUp.Set(1)
	 ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, time.Since(scrapeTime).Seconds(), "connection") 

	wg := &sync.WaitGroup{}
	defer wg.Wait()
	for _, scraper := range e.scrapers {
		wg.Add(1)
		go func(scraper Scraper) {
			defer wg.Done()
			label := "collect." + scraper.Name()
			scrapeTime := time.Now()
			if err := scraper.Scrape(db, ch); err != nil {
				log.Errorln("Error scraping for "+label+":", err)
				e.scrapeErrors.WithLabelValues(label).Inc()
				e.error.Set(1)
			}
			ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, time.Since(scrapeTime).Seconds(), label)
		}(scraper)
	}
}
