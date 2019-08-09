// Scrape `sys_m_cs_unloads`.

package collector

import (
	"database/sql"
	_ "github.com/SAP/go-hdb/driver"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// Scrape query.
	csUnloadsQuery = `SELECT COUNT(*) as CS_UNLOAD_COUNT,SCHEMA_NAME FROM "SYS"."M_CS_UNLOADS"  WHERE SCHEMA_NAME NOT LIKE 'SYS%'  AND SCHEMA_NAME NOT LIKE '_SYS%' AND SCHEMA_NAME NOT LIKE 'HANA%' AND  SCHEMA_NAME NOT LIKE 'UI%' GROUP BY SCHEMA_NAME ORDER BY COUNT(*) DESC;`
	// Subsystem.
	csUnloads = "sys_m_cs_unloads"
)

// Metric descriptors.
var (
	csUnloadsLabels                    = []string{"schema"}
	csUnloadsCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, csUnloads, "count"),
		"column unloads count.",
		csUnloadsLabels, nil)
)

// Scrapedisks collects from `M_CS_UNLOADS;`.
type ScrapeCsUnloads struct{}

// Name of the Scraper. Should be unique.
func (ScrapeCsUnloads) Name() string {
	return csUnloads
}

// Help describes the role of the Scraper.
func (ScrapeCsUnloads) Help() string {
	return "Collect  info from  SYS.M_CS_UNLOADS;"
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeCsUnloads) Scrape(db *sql.DB, ch chan<- prometheus.Metric) error {
	csUnloadsRows, err := db.Query(csUnloadsQuery)
	if err != nil {
		return err
	}
	defer csUnloadsRows.Close()

	var schema string
	var cs_unload_count  float64
	for csUnloadsRows.Next() {
		if err := csUnloadsRows.Scan(&cs_unload_count, &schema); err != nil {
			return err

		}
		csUnloadsLabelValues := []string{schema}
			ch <- prometheus.MustNewConstMetric(csUnloadsCountDesc, prometheus.GaugeValue, cs_unload_count, csUnloadsLabelValues...)
	}
	return nil

}
