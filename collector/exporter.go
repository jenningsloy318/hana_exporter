package collector

import (
	"bytes"
	"database/sql"
	"fmt"
	"strconv"
	"sync"
	"time"

	_ "github.com/SAP/go-hdb/driver"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

// Metric name parts.
const (
	// Subsystem(s).
	exporter = "exporter"
)

// SQL Queries.
const (
	upQuery = `select  SYSTEM_ID AS SID, DATABASE_NAME AS DB_NAME,Version from "SYS"."M_DATABASE";`
)

// Metric descriptors.
var (
	HanaInfoLabelNames  = []string{"sid", "db_name", "db_version"}
	HanaInfoLabelValues = make([]string, 3, 3)
	scrapeDurationDesc  = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, exporter, "collector_duration_seconds"),
		"Collector time duration.",
		[]string{"collector"}, nil,
	)
	hanaUpDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "up"),
		"Collector time duration.",
		nil, nil,
	)
	hanaInfoDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "info"),
		"Collector time duration.",
		HanaInfoLabelNames, nil,
	)
)

// Exporter collects HANA metrics. It implements prometheus.Collector.
type Exporter struct {
	dsn          string
	scrapers     []Scraper
	error        prometheus.Gauge
	totalScrapes prometheus.Counter
	scrapeErrors *prometheus.CounterVec
}

// split string, use @ as delimiter to split the dsn to get hana instance

//func stringsplit(s rune) bool {
//	if s == '@' {
//		return true
//	}
//	return false
//}

// New returns a new HANA exporter for the provided DSN.
func New(host string, user string, password string, scrapers []Scraper) *Exporter {
	//	BaseLabelValues[0] = host
	return &Exporter{
		dsn:      fmt.Sprintf("hdb://%s:%s@%s", user, password, host),
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
		ch <- prometheus.MustNewConstMetric(hanaUpDesc, prometheus.GaugeValue, 0)
		e.error.Set(1)
		return
	} else {
		ch <- prometheus.MustNewConstMetric(hanaUpDesc, prometheus.GaugeValue, 1)
		var sid string
		var db_name string
		var db_version string
		for isUpRows.Next() {
			if err := isUpRows.Scan(&sid, &db_name, &db_version); err != nil {
				return 
			}
		}
		HanaInfoLabelValues = []string{sid, db_name, db_version}

		ch <- prometheus.MustNewConstMetric(hanaInfoDesc, prometheus.GaugeValue, 1, HanaInfoLabelValues...)

	}
	isUpRows.Close()

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

func parseStatus(data sql.RawBytes) (float64, bool) {

	// sys_m_service_statistics
	if bytes.Equal(data, []byte("YES")) {
		return 1, true
	}
	if bytes.Equal(data, []byte("NO")) {
		return 0, true
	}
	if bytes.Equal(data, []byte("UNKNOWN")) {
		return 2, true
	}
	if bytes.Equal(data, []byte("STARTING")) {
		return 3, true
	}
	if bytes.Equal(data, []byte("STOPPING")) {
		return 4, true
	}

	// sys_m_service_replication

	if bytes.Equal(data, []byte("TRUE")) {
		return 1, true
	}
	if bytes.Equal(data, []byte("FALSE")) {
		return 0, true
	}
	if bytes.Equal(data, []byte("ACTIVE")) {
		return 1, true
	}
	if bytes.Equal(data, []byte("ERROR")) {
		return 0, true
	}
	if bytes.Equal(data, []byte("INITIALIZING")) {
		return 3, true
	}
	if bytes.Equal(data, []byte("SYNCING")) {
		return 4, true
	}

	//default transform  to float64
	value, err := strconv.ParseFloat(string(data), 64)
	return value, err == nil
}

func parseConfigString(data string) (float64, bool) {
	// log_mode
	if bytes.Equal([]byte(data), []byte("overwrite")) {
		return 1, true
	}

	if bytes.Equal([]byte(data), []byte("normal")) {
		return 0, true
	}

	//default transform from string to float64
	value, err := strconv.ParseFloat(data, 64)
	return value, err == nil

}
