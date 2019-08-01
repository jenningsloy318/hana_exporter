// Scrape `logModeSystemQuery`.

package collector

import (
	"database/sql"

	_ "github.com/SAP/go-hdb/driver"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// Scrape query.
	logModeSystemQuery = `SELECT VALUE as LOG_MODE FROM M_INIFILE_CONTENTS WHERE FILE_NAME = 'global.ini' AND SECTION = 'persistence' AND  LAYER_NAME='DEFAULT' AND  KEY='log_mode';`
	// Subsystem.
	systemConfig = "system_config"
)

// Metric descriptors.
var (
	logModeSystemDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, systemConfig, "log_mode"),
		"log_mode of the current system layer,0 (normal), 1(overwrite)",
		nil, nil)
)

// ScrapeSystemConfig collects from `SYS.M_CS_TABLES;`.
type ScrapeSystemConfig struct{}

// Name of the Scraper. Should be unique.
func (ScrapeSystemConfig) Name() string {
	return systemConfig
}

// Help describes the role of the Scraper.
func (ScrapeSystemConfig) Help() string {
	return "Collect  info from  system ini config;"
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeSystemConfig) Scrape(db *sql.DB, ch chan<- prometheus.Metric) error {
	// scrape log mode
	logModeRows, err := db.Query(logModeSystemQuery)
	if err != nil {
		return err
	}

	defer logModeRows.Close()

	var log_mode sql.NullString

	for logModeRows.Next() {
		if err := logModeRows.Scan(&log_mode); err != nil {
			return err
		} else if log_mode.Valid {
			if log_mode_value, ok := parseConfigString(log_mode.String); ok {
				ch <- prometheus.MustNewConstMetric(logModeSystemDesc, prometheus.GaugeValue, log_mode_value)
			}
		} else {
			ch <- prometheus.MustNewConstMetric(logModeSystemDesc, prometheus.GaugeValue, 0)
		}
	}

	// also add other scrape to get the system config

	return nil

}
