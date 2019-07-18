package collector

import (
	"database/sql"

	_ "github.com/SAP/go-hdb/driver"

	"github.com/prometheus/client_golang/prometheus"
)

// Scraper is minimal interface that let's you add new prometheus metrics to hana_exporter.
type Scraper interface {
	// Name of the Scraper. Should be unique.
	Name() string
	// Help describes the role of the Scraper.
	// Example: "Collect from SHOW ENGINE INNODB STATUS"
	Help() string
	// Scrape collects data from database connection and sends it over channel as prometheus metric.
	Scrape(db *sql.DB, ch chan<- prometheus.Metric) error
}
