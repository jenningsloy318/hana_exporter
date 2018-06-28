// Scrape `SHOW GLOBAL STATUS`.

package collector

import (
	"database/sql"
	"regexp"
	_ "github.com/SAP/go-hdb/driver"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// Scrape query.
	upStatusQuery = `select  1 as status  from dummy ;`
	// Subsystem.
	UPStatus = "up"
)

// Regexp to match various groups of status vars.
var upStatusRE = regexp.MustCompile(`status`)

// Metric descriptors.
var (
	upStatusDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, UPStatus, "state"),
		"the HANA server is up",
		[]string{"hana_instance"}, nil,
	)
)

// ScrapeUPStatus collects from `select  1 as status  from dummy`.
type ScrapeUPStatus struct{}

// Name of the Scraper. Should be unique.
func (ScrapeUPStatus) Name() string {
	return UPStatus
}

// Help describes the role of the Scraper.
func (ScrapeUPStatus) Help() string {
	return "Whether the HANA server is up."
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeUPStatus) Scrape(db *sql.DB, ch chan<- prometheus.Metric) error {
	upStatusRows, err := db.Query(upStatusQuery)
	if err != nil {
		return err
	}
	defer upStatusRows.Close()

//	var key int
	var key sql.RawBytes


	for upStatusRows.Next() {
		if err := upStatusRows.Scan(&key); err != nil {
			return err
		}
		if floatVal, ok := parseStatus(key); ok { // Unparsable values are silently skipped.

				ch <- prometheus.MustNewConstMetric(upStatusDesc, prometheus.GaugeValue,floatVal, Hana_instance, )
			}

		}

		return nil

}


